package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mohits-git/load-balancer/internal/l4lb"
)

var PORT = ":8081"

func HandleHome(w http.ResponseWriter, r *http.Request) {
	log.Printf("Replied with a Hello message\n\n")
	fmt.Fprintf(w, "Hello From Backend Server at %s\n", PORT)
}

func HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	log.Printf("Replied with a Ok status\n\n")
	w.WriteHeader(http.StatusOK)
}

func main() {
	lb := l4lb.NewL4LoadBalancer()

	// configure servers
	lb.AddServer("127.0.0.1:8081")

	log.Println("Starting Layer 4 TCP Load Balancer at port :8080")
	if err := lb.Start(); err != nil {
		log.Println("Error starting the load balancer", err)
	}
}


// func main() {
// 	mux := http.NewServeMux()
// 	mux.HandleFunc("GET /", utils.HTTPRequestLogger(HandleHome))
// 	mux.HandleFunc("GET /health", utils.HTTPRequestLogger(HandleHealthCheck))
//
// 	fmt.Println("Server Listening on port", PORT)
// 	http.ListenAndServe(PORT, mux)
// }
