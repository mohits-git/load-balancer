package l7lb

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/mohits-git/load-balancer/internal/types"
)

type L7LoadBalancer struct {
	servers             []*HTTPServer
	algo                types.LoadBalancingAlgorithm
	wg                  *sync.WaitGroup
	healthCheckInterval time.Duration
	retryLimit          int
}

func NewL7LoadBalancer(lbalgo types.LoadBalancingAlgorithm, healthCheckInterval time.Duration, retryLimit int) *L7LoadBalancer {
	if healthCheckInterval <= 0 {
		healthCheckInterval = 10 * time.Second
	}
	return &L7LoadBalancer{
		servers:             []*HTTPServer{},
		algo:                lbalgo,
		wg:                  &sync.WaitGroup{},
		healthCheckInterval: healthCheckInterval,
		retryLimit:          retryLimit,
	}
}

func (lb *L7LoadBalancer) AddServer(server *HTTPServer) {
	lb.servers = append(lb.servers, server)
	lb.algo.AddServer(server)
}

func (lb *L7LoadBalancer) Start(port int) error {
	go lb.startHealthCheck()
	mux := http.NewServeMux()
	mux.HandleFunc("/", HTTPRequestLogger(lb.handleNewRequests))

	log.Println("Started load balancer on port 8080")
	if err := http.ListenAndServe(":"+strconv.Itoa(port), mux); err != nil {
		return fmt.Errorf("Error while starting the loadbalancer")
	}

	return nil
}

func (lb *L7LoadBalancer) Stop() {
	log.Println("Stoping L7 Load Balancer...\nWaiting for previous requests to complete")
	lb.wg.Wait()
	os.Exit(0)
}

func (lb *L7LoadBalancer) startHealthCheck() {
	for {
		<-time.After(lb.healthCheckInterval)
		fmt.Println()
		for _, server := range lb.servers {
			go lb.handleHealthCheck(server)
		}
	}
}

func (lb *L7LoadBalancer) handleNewRequests(w http.ResponseWriter, r *http.Request) {
	lb.wg.Add(1)
	defer lb.wg.Done()
	resp := lb.doRequestWithRetry(r)
	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error unable to do request\n"))
		return
	}

	// write resp headers
	for k, v := range resp.Header {
		for _, val := range v {
			w.Header().Add(k, val)
		}
	}

	// write resp received from server
	defer resp.Body.Close()
	if _, err := io.Copy(w, resp.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error unable to forward response\n"))
		return
	}
	log.Println("Replied with response")
}

func (lb *L7LoadBalancer) pickServer() *HTTPServer {
	server := lb.algo.NextServer()
	if server == nil {
		return nil
	}
	httpServer, ok := server.(*HTTPServer)
	if !ok {
		return nil
	}
	return httpServer
}

func (lb *L7LoadBalancer) handleHealthCheck(server types.Server) bool {
	if !server.IsHealthy() {
		server.SetActive(false)
		lb.algo.RemoveServer(server)
		return false
	}
	if !server.IsActive() {
		log.Printf("Adding Server %s Back\n", server.GetAddr())
		server.SetActive(true)
		lb.algo.AddServer(server)
	}
	return true
}

func (lb *L7LoadBalancer) doRequestWithRetry(r *http.Request) *http.Response {
	var resp *http.Response
	var err error
	retryLimit := lb.retryLimit
	if retryLimit < 1 {
		retryLimit = len(lb.servers)
	}
	for range retryLimit {
		server := lb.pickServer()
		if server == nil {
			continue // retry
		}

		resp, err = server.DoRequest(r)
		if err != nil {
			continue // retry
		}

		if resp != nil {
			break
		}
	}

	return resp
}
