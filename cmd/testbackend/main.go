package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mohits-git/load-balancer/internal/utils"
)

var PORT = ":8081"

func HandleHome(w http.ResponseWriter, r *http.Request) {
	log.Printf("Replied with a Hello message\n\n")
	fmt.Fprintf(w, "Hello From Backend Server at %s\n", PORT)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", utils.RequestLogger(HandleHome))

	fmt.Println("Server Listening on port", PORT)
	http.ListenAndServe(PORT, mux)
}
