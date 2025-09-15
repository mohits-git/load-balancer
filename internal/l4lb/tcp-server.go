package l4lb

import (
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/mohits-git/load-balancer/internal/types"
)

// TCPServer is types.Server implementation for TCP servers
type TCPServer struct {
	addr        string
	active      bool
	weight      int
	connections atomic.Int32
}

func NewTCPServer(addr, healthCheckEndpoint string) types.Server {
	return &TCPServer{
		addr:        addr,
		active:      true,
		weight:      1,
		connections: atomic.Int32{},
	}
}

// TCPServer.IsHealthy returns true if server is running
// else return false if not running
func (s *TCPServer) IsHealthy() bool {
	conn, err := net.Dial("tcp", s.addr)
	if err != nil || conn == nil {
		return false
	}
	conn.Close()
	return true
}

// return server's remote addr
func (s *TCPServer) GetAddr() string {
	return s.addr
}

// returns servers weightage
func (s *TCPServer) GetWeight() int {
	return s.weight
}

// sets servers weightage
func (s *TCPServer) SetWeight(weight int) {
	s.weight = weight
}

// returns number of active connections to the server
func (s *TCPServer) GetConnectionsCount() int {
	return int(s.connections.Load())
}

// TCPServer.DoRequest dial tcp connection with the server
// writes request (reqBuf) to the connection
// and reads and returns response returned from the server
func (s *TCPServer) DoRequest(reqBuf []byte) ([]byte, error) {
	serverConn, err := net.Dial("tcp", s.addr)

	if err != nil {
		log.Println("Error connecting to the server")
		return nil, err
	}

	if _, err := serverConn.Write(reqBuf); err != nil {
		log.Println("Error while writing client request to the backend server", s.addr)
		return nil, err

	}

	buf := make([]byte, 1024)
	n, err := serverConn.Read(buf)
	if err != nil && err != io.EOF {
		log.Println("Error reading backend's response")
		return nil, err
	}

	return buf[:n], nil
}
