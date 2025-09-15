package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mohits-git/load-balancer/internal/l7lb"
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
	if port := os.Getenv("PORT"); port != "" {
		PORT = ":" + port
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", l7lb.HTTPRequestLogger(HandleHome))
	mux.HandleFunc("GET /health", l7lb.HTTPRequestLogger(HandleHealthCheck))

	fmt.Println("Server Listening on port", PORT)
	http.ListenAndServe(PORT, mux)
}
