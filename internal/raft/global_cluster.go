package raft

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// GlobalCluster manages all Raft nodes in a shared registry
type GlobalCluster struct {
	nodes map[string]*RaftNode
	mu    sync.RWMutex
}

var globalCluster *GlobalCluster
var once sync.Once

// GetGlobalCluster returns the singleton global cluster instance
func GetGlobalCluster() *GlobalCluster {
	once.Do(func() {
		globalCluster = &GlobalCluster{
			nodes: make(map[string]*RaftNode),
		}
	})
	return globalCluster
}

// RegisterNode registers a node with the global cluster
func (gc *GlobalCluster) RegisterNode(node *RaftNode) error {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	if _, exists := gc.nodes[node.GetID()]; exists {
		return fmt.Errorf("node %s already registered", node.GetID())
	}

	gc.nodes[node.GetID()] = node
	log.Printf("Registered node %s with global cluster", node.GetID())
	return nil
}

// UnregisterNode removes a node from the global cluster
func (gc *GlobalCluster) UnregisterNode(nodeID string) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	if node, exists := gc.nodes[nodeID]; exists {
		node.Stop()
		delete(gc.nodes, nodeID)
		log.Printf("Unregistered node %s from global cluster", nodeID)
	}
}

// GetNode returns a node by ID
func (gc *GlobalCluster) GetNode(nodeID string) (*RaftNode, error) {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	node, exists := gc.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}
	return node, nil
}

// GetLeader returns the current leader node
func (gc *GlobalCluster) GetLeader() (*RaftNode, error) {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	for _, node := range gc.nodes {
		if node.IsLeader() {
			return node, nil
		}
	}
	return nil, fmt.Errorf("no leader found")
}

// GetAllNodes returns all registered nodes
func (gc *GlobalCluster) GetAllNodes() map[string]*RaftNode {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	nodes := make(map[string]*RaftNode)
	for k, v := range gc.nodes {
		nodes[k] = v
	}
	return nodes
}

// GetClusterInfo returns information about the cluster
func (gc *GlobalCluster) GetClusterInfo() map[string]interface{} {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	info := make(map[string]interface{})
	nodes := make(map[string]interface{})

	for id, node := range gc.nodes {
		state, term := node.GetState()
		nodes[id] = map[string]interface{}{
			"id":      id,
			"address": node.GetAddress(),
			"state":   state.String(),
			"term":    term,
			"leader":  node.IsLeader(),
		}
	}

	info["nodes"] = nodes
	info["total_nodes"] = len(gc.nodes)

	// Find leader
	for _, node := range gc.nodes {
		if node.IsLeader() {
			info["leader"] = node.GetID()
			break
		}
	}

	return info
}

// StopAll stops all nodes in the cluster
func (gc *GlobalCluster) StopAll() {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	for id, node := range gc.nodes {
		log.Printf("Stopping node %s", id)
		node.Stop()
	}
	gc.nodes = make(map[string]*RaftNode)
}

// StartHeartbeatMonitor monitors the cluster and ensures only one leader
func (gc *GlobalCluster) StartHeartbeatMonitor() {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			gc.mu.RLock()
			leaders := make([]*RaftNode, 0)
			for _, node := range gc.nodes {
				if node.IsLeader() {
					leaders = append(leaders, node)
				}
			}
			gc.mu.RUnlock()

			// If multiple leaders, step down all but the one with highest term
			if len(leaders) > 1 {
				log.Printf("WARNING: Multiple leaders detected (%d), resolving conflict", len(leaders))

				// Find leader with highest term
				var highestTermLeader *RaftNode
				highestTerm := -1
				for _, leader := range leaders {
					_, term := leader.GetState()
					if term > highestTerm {
						highestTerm = term
						highestTermLeader = leader
					}
				}

				// Step down all other leaders
				for _, leader := range leaders {
					if leader != highestTermLeader {
						leader.StepDown()
						log.Printf("Stepped down leader %s", leader.GetID())
					}
				}
			}
		}
	}()
}
