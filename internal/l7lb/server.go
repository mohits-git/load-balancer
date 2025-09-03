package l7lb

import (
	"log"
	"net/http"
	"time"

	"github.com/mohits-git/load-balancer/internal/utils"
)

type HTTPServer struct {
	addr                string
	active              bool
	healthCheckEndpoint string
}

func (s *HTTPServer) IsHealthy() bool {
	resp, err := http.Get(s.addr + s.healthCheckEndpoint)
	if err != nil {
		return false
	}
	if resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

// forwards the request to the backend server
// copys and build a new request
func (s *HTTPServer) DoRequest(r *http.Request) (*http.Response, error) {
	clientIP, clientPort := utils.GetHTTPClientRemoteAddrInfo(r)

	client := http.Client{
		Timeout: 60 * time.Second, // timeout duration
	}

	url := "http://" + s.addr + r.URL.Path + "?" + r.URL.RawQuery
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
