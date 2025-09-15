package l7lb

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mohits-git/load-balancer/internal/types"
	"github.com/mohits-git/load-balancer/internal/utils"
)

type L7LoadBalancer struct {
	servers []types.Server
	algo    types.LoadBalancingAlgorithm
}

func NewL7LoadBalancer(lbalgo types.LoadBalancingAlgorithm) types.LoadBalancer {
	return &L7LoadBalancer{
		servers: []types.Server{},
		algo:    lbalgo,
	}
}

func (lb *L7LoadBalancer) AddServer(server types.Server) {
	lb.servers = append(lb.servers, server)
	lb.algo.AddServer(server)
}

func (lb *L7LoadBalancer) PickServer() types.Server {
	return lb.algo.NextServer()
}

func (lb *L7LoadBalancer) StartHealthCheck() {
	for {
		<-time.After(10 * time.Second)
		for _, server := range lb.servers {
			go func() {
				if !server.IsHealthy() {
					server.SetActive(false)
					lb.algo.RemoveServer(server)
				} else {
					if !server.IsActive() {
						server.SetActive(true)
						lb.algo.AddServer(server)
					}
				}
			}()
		}
	}
}

func (lb *L7LoadBalancer) Start() error {
	go lb.StartHealthCheck()
	mux := http.NewServeMux()
	mux.HandleFunc("/", utils.HTTPRequestLogger(lb.handleNewRequests))

	log.Println("Started load balancer on port 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		return fmt.Errorf("Error while starting the loadbalancer")
	}

	return nil
}

func (lb *L7LoadBalancer) handleNewRequests(w http.ResponseWriter, r *http.Request) {
	server := lb.PickServer()
	resp, err := server.(*HTTPServer).DoRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error, unable to do request"))
		return
	}

	// copy headers
	for k, v := range resp.Header {
		for _, val := range v {
			w.Header().Add(k, val)
		}
	}

	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error, unable to forward response"))
		return
	}
	log.Println("Replied with response")
}

func (lb *L7LoadBalancer) Stop() {
	log.Println("Stoping L7 Load Balancer...")
	os.Exit(0)
}
