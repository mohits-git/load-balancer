package types

type LoadBalancer interface {
	Start(port int) error
	Stop()
}
