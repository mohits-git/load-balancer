package l4lb

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/mohits-git/load-balancer/internal/types"
)

// Layer 4 Load balancer, distributes tcp requests load to multiple backend servers
type L4LoadBalancer struct {
	servers  []types.Server
	listener net.Listener
	connWg   *sync.WaitGroup
	algo     types.LoadBalancingAlgorithm
}

// returns new l4 load balancer
func NewL4LoadBalancer(lbalgo types.LoadBalancingAlgorithm) types.LoadBalancer {
	return &L4LoadBalancer{
		servers:  []types.Server{},
		listener: nil,
		connWg:   &sync.WaitGroup{},
		algo:     lbalgo,
	}
}

// adds a new tcp server with address as 'addr'
func (lb *L4LoadBalancer) AddServer(server types.Server) {
	lb.algo.AddServer(server)
	lb.servers = append(lb.servers, server)
}

// uses load balancing algorithms to pick a server to forward next req to
func (lb *L4LoadBalancer) pickServer() types.Server {
	return lb.algo.NextServer()
}

// starts the load balancer tcp server
func (lb *L4LoadBalancer) Start() error {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		return fmt.Errorf("Error starting a tcp server: %w", err)
	}
	lb.listener = ln

	connChan := make(chan net.Conn, 100)
	go lb.acceptConnections(connChan)

	for conn := range connChan {
		lb.connWg.Add(1)
		go lb.handleConn(conn)
	}
	return nil
}

func (lb *L4LoadBalancer) acceptConnections(connChan chan net.Conn) {
	defer close(connChan)
	for {
		conn, err := lb.listener.Accept()
		if err != nil && errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			panic("Error while Accepting a new tcp connection")
		}
		connChan <- conn
	}
}

func (lb *L4LoadBalancer) handleConn(conn net.Conn) {
	defer conn.Close()
	defer lb.connWg.Done()

	log.Printf("Got request from %s\n\n", conn.RemoteAddr().String())

	reqBuf := make([]byte, 1024)
	n, err := conn.Read(reqBuf)
	if err != nil {
		log.Println("Error reading request")
		return
	}
	fmt.Println(string(reqBuf[:n]))

	server := lb.pickServer()

	resp, err := server.(*TCPServer).DoRequest(reqBuf)
	if err != nil {
		conn.Write([]byte("HTTP/1.1 500 Internal Server Error"))
	}

	if _, err := conn.Write(resp); err != nil {
		log.Println("Error while writing response to client")
		return
	}

	log.Printf("Replied client with response:\n\n")
	fmt.Println(string(resp))
}

func (lb *L4LoadBalancer) Stop() {
	if lb.listener == nil {
		os.Exit(0)
	}
	log.Println("Stoping the L4 LoadBalancer...")
	if err := lb.listener.Close(); err != nil {
		panic(fmt.Errorf("Error closing the tcp listener %w", err))
	}
	log.Println("Waiting for client requests to complete...")
	lb.connWg.Wait()
	os.Exit(0)
}
