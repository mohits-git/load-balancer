package types

type Server interface {
	IsHealthy() bool
}
