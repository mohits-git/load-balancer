package types

type LoadBalancer interface {
	AddServer(Server)
	Start() error
	Stop()
	PickServer() Server
}
