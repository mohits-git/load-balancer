package types

type Server interface {
	GetAddr() string
	GetWeight() int
	SetWeight(int)
	GetConnectionsCount() int
	IsHealthy() bool
}
