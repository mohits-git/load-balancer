package l4lb

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/mohits-git/load-balancer/internal/types"
)

// Layer 4 Load balancer, distributes tcp requests load to multiple backend servers
type L4LoadBalancer struct {
	servers             []*TCPServer
	listener            net.Listener
	connWg              *sync.WaitGroup
	algo                types.LoadBalancingAlgorithm
	healthCheckInterval time.Duration
	retryLimit          int
}

// returns new l4 load balancer
func NewL4LoadBalancer(lbalgo types.LoadBalancingAlgorithm, healthCheckInterval time.Duration, retryLimit int) *L4LoadBalancer {
	if healthCheckInterval <= 0 {
		healthCheckInterval = 10 * time.Second
	}
	return &L4LoadBalancer{
		servers:             []*TCPServer{},
		listener:            nil,
		connWg:              &sync.WaitGroup{},
		algo:                lbalgo,
		healthCheckInterval: healthCheckInterval,
		retryLimit:          retryLimit,
	}
}

// adds a new tcp server with address as 'addr'
func (lb *L4LoadBalancer) AddServer(server *TCPServer) {
	lb.algo.AddServer(server)
	lb.servers = append(lb.servers, server)
}

// uses load balancing algorithms to pick a server to forward next req to
func (lb *L4LoadBalancer) pickServer() *TCPServer {
	server := lb.algo.NextServer()
	if server == nil {
		return nil
	}
	tcpServer, ok := server.(*TCPServer)
	if !ok {
		return nil
	}
	return tcpServer
}

// starts the load balancer tcp server
func (lb *L4LoadBalancer) Start(port int) error {
	go lb.startHealthCheck()
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
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

func (lb *L4LoadBalancer) handleHealthCheck(server types.Server) bool {
	if !server.IsHealthy() {
		server.SetActive(false)
		lb.algo.RemoveServer(server)
		return false
	}
	if !server.IsActive() {
		server.SetActive(true)
		lb.algo.AddServer(server)
	}
	return true
}

func (lb *L4LoadBalancer) startHealthCheck() {
	for {
		<-time.After(lb.healthCheckInterval)
		for _, server := range lb.servers {
			go lb.handleHealthCheck(server)
		}
	}
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

func (lb *L4LoadBalancer) doRequestWithRetry(reqBuf []byte) []byte {
	var resp []byte
	var err error

	retryLimit := lb.retryLimit
	if retryLimit <= 0 {
		retryLimit = len(lb.servers)
	}
	for range retryLimit {
		server := lb.pickServer()
		if server == nil {
			continue
		}

		resp, err = server.DoRequest(reqBuf)
		if err != nil {
			continue
		}

		if resp != nil {
			break
		}
	}

	return resp
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

	resp := lb.doRequestWithRetry(reqBuf)
	if resp == nil {
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
