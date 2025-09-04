package types

type Server interface {
	GetAddr() string
	GetWeight() int
	GetConnectionsCount() int
	IsHealthy() bool
}
