package cluster

import (
	"sync"
	"time"
)

type Coordinator struct {
	nodes      map[string]*Node
	nodesMu    sync.RWMutex
	leader     *Node
	leaderMu   sync.RWMutex
	shutdown   bool
	shutdownMu sync.Mutex
}

func NewCoordinator() *Coordinator {
	return &Coordinator{
		nodes: make(map[string]*Node),
	}
}

func (c *Coordinator) AddNode(addr string) {
	c.nodesMu.Lock()
	defer c.nodesMu.Unlock()
	if _, exists := c.nodes[addr]; exists {
		return
	}

	node := NewNode(addr)
	c.nodes[addr] = node
	go c.monitorNode(node)
}

func (c *Coordinator) RemoveNode(addr string) {
	c.nodesMu.Lock()
	defer c.nodesMu.Unlock()
	if node, exists := c.nodes[addr]; exists {
		node.Shutdown()
		delete(c.nodes, addr)
	}
}

func (c *Coordinator) GetNode(addr string) *Node {
	c.nodesMu.RLock()
	defer c.nodesMu.RUnlock()
	node, exists := c.nodes[addr]
	if !exists {
		return nil
	}
	return node
}

func (c *Coordinator) GetAllNodes() []*Node {
	c.nodesMu.RLock()
	defer c.nodesMu.RUnlock()
	nodes := make([]*Node, 0, len(c.nodes))
	for _, node := range c.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

func (c *Coordinator) monitorNode(node *Node) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if node.IsHealthy() {
				continue
			}

			// Attempt to reconnect
			if err := node.Connect(); err == nil {
				node.MarkHealthy(true)
			}
		case <-c.getShutdownChan():
			return
		}
	}
}

func (c *Coordinator) getShutdownChan() <-chan struct{} {
	c.shutdownMu.Lock()
	defer c.shutdownMu.Unlock()
	if c.shutdown {
		return make(chan struct{})
	}
	return nil
}

func (c *Coordinator) Shutdown() {
	c.shutdownMu.Lock()
	defer c.shutdownMu.Unlock()
	if c.shutdown {
		return
	}

	c.shutdown = true
	c.nodesMu.Lock()
	defer c.nodesMu.Unlock()
	for _, node := range c.nodes {
		node.Shutdown()
	}
}
