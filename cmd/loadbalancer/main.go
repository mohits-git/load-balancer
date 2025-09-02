package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/mohits-git/load-balancer/internal/utils"
)

const BACKEND_ADDR = "127.0.0.1:8081"

func HandleNewRequests(w http.ResponseWriter, r *http.Request) {
	// forward the request to the backend server
	// copy and build new request
	clientIP, clientPort := utils.GetClientRemoteAddrInfo(r)

	client := http.Client{
		Timeout: 60 * time.Second, // timeout duration
	}

	url := "http://" + BACKEND_ADDR + r.URL.Path + "?" + r.URL.RawQuery
	newReq, err := http.NewRequest(r.Method, url, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error, unable create a new request"))
		return
	}

	for key, vals := range r.Header {
		for _, val := range vals {
			newReq.Header.Add(key, val)
		}
	}
	newReq.Header.Set("Host", BACKEND_ADDR)
	newReq.Header.Set("X-Forwarded-For", clientIP+":"+clientPort)

	// get the response
	resp, err := client.Do(newReq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error, unable to do request"))
		return
	}

	// forward response
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

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", utils.RequestLogger(HandleNewRequests))

	fmt.Println("Started load balancer on port 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Println("Error while starting the loadbalancer")
	}
}
