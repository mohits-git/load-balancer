package l4lb

import (
	"io"
	"log"
	"net"
)

// TCPServer is types.Server implementation for TCP servers
type TCPServer struct {
	addr   string
	active bool
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
