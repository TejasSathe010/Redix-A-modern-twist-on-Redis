package cluster

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

type Node struct {
	addr        string
	conn        net.Conn
	mu          sync.Mutex
	lastSeen    time.Time
	healthy     bool
	shutdown    bool
	shutdownMu  sync.Mutex
	commandChan chan []string
}

func NewNode(addr string) *Node {
	return &Node{
		addr:        addr,
		commandChan: make(chan []string, 100),
	}
}

func (n *Node) Connect() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	var err error
	n.conn, err = net.Dial("tcp", n.addr)
	n.lastSeen = time.Now()
	n.healthy = err == nil
	return err
}

func (n *Node) IsHealthy() bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.healthy
}

func (n *Node) MarkHealthy(healthy bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.healthy = healthy
}

func (n *Node) SendCommand(ctx context.Context, cmd []string) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.conn == nil {
		return fmt.Errorf("not connected to node")
	}

	// Format command for transmission
	// Using Redis protocol format
	writer := bufio.NewWriter(n.conn)

	// Start array
	writer.WriteString(fmt.Sprintf("*%d\r\n", len(cmd)))

	for _, arg := range cmd {
		writer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg))
	}

	writer.Flush()

	return nil
}

func (n *Node) StartCommandProcessor(handler network.CommandHandler) {
	go func() {
		for cmd := range n.commandChan {
			// Process incoming commands from other nodes
			// This is a simplified implementation
			// In a real scenario, you would execute the command
			// and send the response back
			fmt.Printf("Received command from cluster: %v\n", cmd)
		}
	}()
}

func (n *Node) Shutdown() {
	n.shutdownMu.Lock()
	defer n.shutdownMu.Unlock()
	if n.shutdown {
		return
	}

	n.shutdown = true
	close(n.commandChan)
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.conn != nil {
		n.conn.Close()
		n.conn = nil
	}
}
