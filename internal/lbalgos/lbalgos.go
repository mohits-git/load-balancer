package lbalgos

import "github.com/mohits-git/load-balancer/internal/types"

func NewLoadBalancerAlgorithm(algoType string) types.LoadBalancingAlgorithm {
	var algo types.LoadBalancingAlgorithm
	switch algoType {
	case "Round Robin":
		algo = NewRoundRobinAlgo()
	case "Weighted Round Robin":
		algo = NewWeightedRoundRobin()
	}
	return algo
}
