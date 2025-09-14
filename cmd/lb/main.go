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
	roundRobin := lbalgos.NewRoundRobinAlgo()

	// lb := l7lb.NewL7LoadBalancer()
	lb := l7lb.NewL7LoadBalancer(roundRobin)
	lb.AddServer("127.0.0.1:8081")
	lb.AddServer("127.0.0.1:8082")
	lb.AddServer("127.0.0.1:8083")

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
