package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mohits-git/load-balancer/internal/config"
	"github.com/mohits-git/load-balancer/internal/l4lb"
	"github.com/mohits-git/load-balancer/internal/l7lb"
	"github.com/mohits-git/load-balancer/internal/lbalgos"
	"github.com/mohits-git/load-balancer/internal/types"
)

func main() {
	cfg := config.LoadConfig()

	algo := lbalgos.NewLoadBalancerAlgorithm(cfg.Algorithm)

	var lb types.LoadBalancer
	switch cfg.Protocol {
	case "http":
		lb = SetupL7LoadBalancer(cfg, algo)
	case "tcp":
		lb = SetupL4LoadBalancer(cfg, algo)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		lb.Stop()
	}()

	log.Printf("Starting %s Load Balancer at port :%d\n", cfg.Protocol, cfg.Port)
	if err := lb.Start(cfg.Port); err != nil {
		log.Println("Error starting the load balancer", err)
	}
}

func SetupL7LoadBalancer(cfg *config.Config, algo types.LoadBalancingAlgorithm) *l7lb.L7LoadBalancer {
	lb := l7lb.NewL7LoadBalancer(
		algo,
		time.Duration(cfg.HealthCheckInterval)*time.Second,
		cfg.RetryLimit,
	)
	for _, server := range cfg.Servers {
		httpServer := l7lb.NewHTTPServer(server.Addr, server.HealthCheckHTTPEndpoint)
		httpServer.SetWeight(server.Weight)
		lb.AddServer(httpServer)

	}
	return lb
}

func SetupL4LoadBalancer(cfg *config.Config, algo types.LoadBalancingAlgorithm) *l4lb.L4LoadBalancer {
	lb := l4lb.NewL4LoadBalancer(
		algo,
		time.Duration(cfg.HealthCheckInterval)*time.Second,
		cfg.RetryLimit,
	)
	for _, server := range cfg.Servers {
		tcpServer := l4lb.NewTCPServer(server.Addr)
		tcpServer.SetWeight(server.Weight)
		lb.AddServer(tcpServer)
	}
	return lb
}
