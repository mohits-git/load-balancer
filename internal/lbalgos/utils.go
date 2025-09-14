package lbalgos

import "github.com/mohits-git/load-balancer/internal/types"

func isSameAddr(server types.Server) func(e types.Server) bool {
	return func(e types.Server) bool {
		return e.GetAddr() == server.GetAddr()
	}
}
