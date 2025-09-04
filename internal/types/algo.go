package types

type LoadBalancingAlgorithm interface {
	AddServer(Server)
	RemoveServer(Server)
	NextServer() Server
}
