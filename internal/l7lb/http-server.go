package l7lb

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/mohits-git/load-balancer/internal/utils"
)

type HTTPServer struct {
	addr                string
	active              bool
	healthCheckEndpoint string
	weight              int
	connections         atomic.Int32
}

func NewHTTPServer(addr, healthCheckEndpoint string) *HTTPServer {
	return &HTTPServer{
		addr:                addr,
		healthCheckEndpoint: healthCheckEndpoint,
		active:              true,
		weight:              1,
		connections:         atomic.Int32{},
	}
}

func (s *HTTPServer) IsHealthy() bool {
	resp, err := http.Get("http://" + s.addr + s.healthCheckEndpoint) // TODO: join path safely
	if err != nil {
		fmt.Println("Server found to be not healthy", s.addr, s.healthCheckEndpoint)
		return false
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Server found to be not healthy", s.addr, s.healthCheckEndpoint)
		return false
	}
	fmt.Println("Server active")
	return true
}

func (s *HTTPServer) IsActive() bool {
	return s.active
}

func (s *HTTPServer) SetActive(active bool) {
	s.active = active
}

// returns server's remote addr
func (s *HTTPServer) GetAddr() string {
	return s.addr
}

// returns servers weightage
func (s *HTTPServer) GetWeight() int {
	return s.weight
}

// sets servers weightage
func (s *HTTPServer) SetWeight(weight int) {
	s.weight = weight
}

// returns number of active connections to the server
func (s *HTTPServer) GetConnectionsCount() int {
	return int(s.connections.Load())
}

// forwards the request to the backend server
// copys and build a new request
func (s *HTTPServer) DoRequest(r *http.Request) (*http.Response, error) {
	clientIP, clientPort := utils.GetHTTPClientRemoteAddrInfo(r)

	client := http.Client{Timeout: 60 * time.Second}

	url := "http://" + s.addr + r.URL.Path + "?" + r.URL.RawQuery // TODO: safe join path
	newReq, err := http.NewRequest(r.Method, url, r.Body)
	if err != nil {
		return nil, err
	}

	for key, vals := range r.Header {
		for _, val := range vals {
			newReq.Header.Add(key, val)
		}
	}
	newReq.Header.Set("Host", s.addr)
	newReq.Header.Set("X-Forwarded-For", clientIP+":"+clientPort)

	resp, err := client.Do(newReq)
	if err != nil {
		log.Println("Error doing request: ", err)
	}

	return resp, err
}
