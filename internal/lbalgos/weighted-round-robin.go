package lbalgos

import (
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

func (w *WeightedRoundRobin) AddServer(server types.Server) {
	w.mu.Lock()
	defer w.mu.Unlock()
	i := slices.IndexFunc(w.servers, func(s types.Server) bool {
		return s.GetAddr() == server.GetAddr()
	})
	if i != -1 {
		w.servers = append(w.servers, server)
		w.weights = append(w.weights, server.GetWeight())
	}
}

func (w *WeightedRoundRobin) RemoveServer(server types.Server) {
	w.mu.Lock()
	defer w.mu.Unlock()
	i := slices.IndexFunc(w.servers, func(s types.Server) bool {
		return s.GetAddr() == server.GetAddr()
	})
	w.servers = append(w.servers[:i], w.servers[i+1:]...)
	w.weights = append(w.weights[:i], w.weights[i+1:]...)
}

func (w *WeightedRoundRobin) NextServer() types.Server {
	w.mu.Lock()
	defer w.mu.Unlock()

	currIndex := w.current
	nextIndex := currIndex

	if currIndex >= len(w.servers) {
		currIndex = 0
	}

	currWeight := w.weights[currIndex]
	if currWeight == 0 {
		// TODO: do weighted load balancing
		w.weights[currIndex] = w.servers[currIndex].GetWeight()
	}

	nextIndex = (int(currIndex) + 1) % len(w.servers)
	w.current = nextIndex
	return w.servers[nextIndex]
}
