package cluster

import (
	"crypto/sha1"
	"math/big"
	"sync"
)

const numSlots = 16384

type ShardingManager struct {
	slots [numSlots]*Node
	nodes []*Node
	mu    sync.RWMutex
}

func NewShardingManager() *ShardingManager {
	return &ShardingManager{
		slots: [numSlots]*Node{}, // Initialize array with zero values
	}
}

func (sm *ShardingManager) AddNode(node *Node) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.nodes = append(sm.nodes, node)
	sm.rebalanceSlots()
}

func (sm *ShardingManager) RemoveNode(node *Node) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	newNodes := make([]*Node, 0, len(sm.nodes))
	for _, n := range sm.nodes {
		if n != node {
			newNodes = append(newNodes, n)
		}
	}
	sm.nodes = newNodes
	sm.rebalanceSlots()
}

func (sm *ShardingManager) GetNodeForKey(key string) (*Node, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	slot := sm.calculateHashSlot(key)
	return sm.slots[slot], nil
}

func (sm *ShardingManager) rebalanceSlots() {
	// Distribute slots evenly among nodes
	sm.mu.Lock()
	defer sm.mu.Unlock()
	totalSlots := numSlots
	numNodes := len(sm.nodes)
	if numNodes == 0 {
		return
	}

	slotsPerNode := totalSlots / numNodes
	remainingSlots := totalSlots % numNodes

	currentSlot := 0
	for i, node := range sm.nodes {
		nodeSlots := slotsPerNode
		if i < remainingSlots {
			nodeSlots++
		}

		for j := 0; j < nodeSlots && currentSlot < totalSlots; j++ {
			sm.slots[currentSlot] = node
			currentSlot++
		}
	}
}

func (sm *ShardingManager) calculateHashSlot(key string) int {
	// Calculate hash slot for the key
	hash := sha1.Sum([]byte(key))
	hashInt := big.Int{}
	hashInt.SetBytes(hash[:])

	// Redis uses the higher 14 bits for slot calculation
	return int(hashInt.Uint64() % uint64(numSlots))
}
