package types

type LoadBalancer interface {
	AddServer(addr string)
	Start() error
	PickServer() Server
}
