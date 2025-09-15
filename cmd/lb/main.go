package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mohits-git/load-balancer/internal/l7lb"
	"github.com/mohits-git/load-balancer/internal/lbalgos"
)

func main() {
	weightedRoundRobin := lbalgos.NewWeightedRoundRobin()

	// servers
	server1 := l7lb.NewHTTPServer("127.0.0.1:8081", "/health")
	server1.SetWeight(1)
	server2 := l7lb.NewHTTPServer("127.0.0.1:8082", "/health")
	server2.SetWeight(1)
	server3 := l7lb.NewHTTPServer("127.0.0.1:8083", "/health")

	// lb := l7lb.NewL7LoadBalancer()
	lb := l7lb.NewL7LoadBalancer(weightedRoundRobin)
	lb.AddServer(server1)
	lb.AddServer(server2)
	lb.AddServer(server3)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		lb.Stop()
	}()

	log.Println("Starting Layer 4 Load Balancer at port :8080")
	if err := lb.Start(); err != nil {
		log.Println("Error starting the load balancer", err)
	}
}
