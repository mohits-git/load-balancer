package l7lb

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

type HTTPServer struct {
	addr                string
	active              bool
	healthCheckEndpoint string
	weight              int
	connections         atomic.Int32
	client              http.Client
}

func NewHTTPServer(addr, healthCheckEndpoint string) *HTTPServer {
	return &HTTPServer{
		addr:                addr,
		healthCheckEndpoint: healthCheckEndpoint,
		active:              true,
		weight:              1,
		connections:         atomic.Int32{},
		client: http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				IdleConnTimeout:   90 * time.Second,
				DisableKeepAlives: false,
			},
		},
	}
}

func (s *HTTPServer) IsHealthy() bool {
	reqUrl, err := url.JoinPath("http://", s.addr, s.healthCheckEndpoint)
	if err != nil {
		log.Println("Invalid address or health check endpoint", err)
		return false
	}
	resp, err := http.Get(reqUrl)
	if err != nil {
		log.Println("Server found to be not healthy", s.addr, s.healthCheckEndpoint)
		return false
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Server found to be not healthy", s.addr, s.healthCheckEndpoint)
		return false
	}
	log.Printf("Server %s active", s.addr)
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
	clientIP, clientPort := GetHTTPClientRemoteAddrInfo(r)

	reqUrl, err := url.JoinPath("http://", s.addr, r.URL.Path)
	if err != nil {
		return nil, fmt.Errorf("Invalid address or health check endpoint: %w", err)
	}
	reqUrlWithQuery := reqUrl + "?" + r.URL.RawQuery

	newReq, err := http.NewRequest(r.Method, reqUrlWithQuery, r.Body)
	if err != nil {
		return nil, err
	}

	for key, vals := range r.Header {
		for _, val := range vals {
			newReq.Header.Add(key, val)
		}
	}
	newReq.Header.Set("Host", s.addr)
	newReq.Header.Set("X-Forwarded-For", net.JoinHostPort(clientIP, clientPort))

	resp, err := s.client.Do(newReq)
	if err != nil {
		log.Println("Error doing request: ", err)
		return nil, err
	}

	return resp, nil
}
