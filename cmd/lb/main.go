package main

import (
	"log"

	"github.com/mohits-git/load-balancer/internal/l7lb"
)

func main() {
	lb := l7lb.NewL7LoadBalancer()
	lb.AddServer("127.0.0.1:8081")
	log.Println("Starting Layer 7 HTTP Load Balancer at port :8080")
	if err := lb.Start(); err != nil {
		log.Println("Error starting the load balancer", err)
	}
}
