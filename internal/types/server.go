package types

type Server interface {
	IsActive() bool
	SetActive(active bool)
	GetAddr() string
	GetWeight() int
	SetWeight(int)
	GetConnectionsCount() int
	IsHealthy() bool
}
