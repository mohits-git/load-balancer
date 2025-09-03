package l4lb

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/mohits-git/load-balancer/internal/types"
)

// Layer 4 Load balancer, distributes tcp requests load to multiple backend servers
type L4LoadBalancer struct {
	servers  []types.Server
	listener net.Listener
	connWg   *sync.WaitGroup
}

// returns new l4 load balancer
func NewL4LoadBalancer() types.LoadBalancer {
	return &L4LoadBalancer{
		servers:  []types.Server{},
		listener: nil,
		connWg:   &sync.WaitGroup{},
	}
}

// adds a new tcp server with address as 'addr'
func (lb *L4LoadBalancer) AddServer(addr string) {
	lb.servers = append(lb.servers, &TCPServer{addr, true})
}

// uses load balancing algorithms to pick a server to forward next req to
func (lb *L4LoadBalancer) PickServer() types.Server {
	// TODO: from load balancing algo -> lb.servers[i]
	return &TCPServer{"127.0.0.1:8081", true}
}

// starts the load balancer tcp server
func (lb *L4LoadBalancer) Start() error {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		return fmt.Errorf("Error starting a tcp server: %w", err)
	}
	lb.listener = ln

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	connChan := make(chan net.Conn, 100)
	go lb.acceptConnections(connChan)

	for {
		select {
		case <-sigChan:
			lb.stop()
		case conn := <-connChan:
			lb.connWg.Add(1)
			go lb.handleConn(conn)
		default:
			continue
		}
	}
}

func (lb *L4LoadBalancer) acceptConnections(connChan chan net.Conn) {
	defer close(connChan)
	for {
		conn, err := lb.listener.Accept()
		if err != nil && !errors.Is(err, net.ErrClosed) {
			panic("Error while Accepting a new tcp connection")
		}
		if err != nil && errors.Is(err, net.ErrClosed) {
			return
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

	server := lb.PickServer()

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

func (lb *L4LoadBalancer) stop() {
	log.Println("Stoping the L4 LoadBalancer...")
	if err := lb.listener.Close(); err != nil {
		panic(fmt.Errorf("Error closing the tcp listener %w", err))
	}
	log.Println("Waiting for client requests to complete...")
	lb.connWg.Wait()
	os.Exit(0)
}
