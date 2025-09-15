package lbalgos

import (
	"fmt"
	"slices"
	"sync"

	"github.com/mohits-git/load-balancer/internal/types"
)

type WeightedRoundRobin struct {
	current int
	servers []types.Server
	weights []int
	mu      *sync.Mutex
}

func NewWeightedRoundRobin() *WeightedRoundRobin {
	return &WeightedRoundRobin{
		current: 0,
		servers: []types.Server{},
		weights: []int{},
		mu:      &sync.Mutex{},
	}
}

func (w *WeightedRoundRobin) AddServer(server types.Server) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if i := slices.IndexFunc(w.servers, isSameAddr(server)); i == -1 {
		w.servers = append(w.servers, server)
		w.weights = append(w.weights, server.GetWeight())
	}
}

func (w *WeightedRoundRobin) RemoveServer(server types.Server) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if i := slices.IndexFunc(w.servers, isSameAddr(server)); i != -1 {
		fmt.Println("Removing server", server.GetAddr())
		w.servers = append(w.servers[:i], w.servers[i+1:]...)
		w.weights = append(w.weights[:i], w.weights[i+1:]...)
	}
}

func (w *WeightedRoundRobin) NextServer() types.Server {
	w.mu.Lock()
	defer w.mu.Unlock()

	currIndex := w.current

	if currIndex >= len(w.servers) {
		currIndex = 0
	}

	currWeight := w.weights[currIndex]
	if currWeight == 0 {
		w.weights[currIndex] = w.servers[currIndex].GetWeight()
		currIndex = (currIndex + 1) % len(w.servers)
		currWeight = w.weights[currIndex]
	}

	w.weights[currIndex] = currWeight - 1

	w.current = currIndex
	return w.servers[currIndex]
}
