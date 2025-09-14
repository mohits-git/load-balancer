package lbalgos

import (
	"slices"
	"sync/atomic"

	"github.com/mohits-git/load-balancer/internal/types"
)

type RoundRobin struct {
	current atomic.Int32
	servers []types.Server
}

func NewRoundRobinAlgo() types.LoadBalancingAlgorithm {
	return &RoundRobin{
		current: atomic.Int32{},
		servers: make([]types.Server, 0),
	}
}

func (rb *RoundRobin) AddServer(server types.Server) {
	if i := slices.IndexFunc(rb.servers, isSameAddr(server)); i == -1 {
		rb.servers = append(rb.servers, server)
	}
}

func (rb *RoundRobin) RemoveServer(server types.Server) {
	rb.servers = slices.DeleteFunc(rb.servers, isSameAddr(server))
}

func (rb *RoundRobin) NextServer() types.Server {
	currIndex := rb.current.Load()
	nextIndex := (int(currIndex) + 1) % len(rb.servers)
	rb.current.Store(int32(nextIndex))
	return rb.servers[currIndex]
}
