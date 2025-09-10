package raft

import (
	"fmt"
	"log"
	"sync"

	"godatabase/internal/storage"
)

// Cluster represents a Raft cluster
type Cluster struct {
	nodes map[string]*RaftNode
	mu    sync.RWMutex
}

// NewCluster creates a new Raft cluster
func NewCluster() *Cluster {
	return &Cluster{
		nodes: make(map[string]*RaftNode),
	}
}

// AddNode adds a node to the cluster
func (c *Cluster) AddNode(id, address string, peers map[string]string, storage storage.Storage) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.nodes[id]; exists {
		return fmt.Errorf("node %s already exists", id)
	}

	node := NewRaftNode(id, address, peers, storage)
	c.nodes[id] = node

	// Start the node
	if err := node.Start(); err != nil {
		delete(c.nodes, id)
		return fmt.Errorf("failed to start node %s: %v", id, err)
	}

	// Start RPC server
	if err := node.StartRPCServer(); err != nil {
		node.Stop()
		delete(c.nodes, id)
		return fmt.Errorf("failed to start RPC server for node %s: %v", id, err)
	}

	log.Printf("Added node %s to cluster", id)
	return nil
}

// RemoveNode removes a node from the cluster
func (c *Cluster) RemoveNode(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, exists := c.nodes[id]
	if !exists {
		return fmt.Errorf("node %s not found", id)
	}

	node.Stop()
	delete(c.nodes, id)
	log.Printf("Removed node %s from cluster", id)
	return nil
}

// GetNode returns a node by ID
func (c *Cluster) GetNode(id string) (*RaftNode, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	node, exists := c.nodes[id]
	if !exists {
		return nil, fmt.Errorf("node %s not found", id)
	}

	return node, nil
}

// GetLeader returns the current leader node
func (c *Cluster) GetLeader() (*RaftNode, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, node := range c.nodes {
		if node.IsLeader() {
			return node, nil
		}
	}

	return nil, fmt.Errorf("no leader found")
}

// GetNodes returns all nodes in the cluster
func (c *Cluster) GetNodes() map[string]*RaftNode {
	c.mu.RLock()
	defer c.mu.RUnlock()

	nodes := make(map[string]*RaftNode)
	for k, v := range c.nodes {
		nodes[k] = v
	}
	return nodes
}

// Stop stops all nodes in the cluster
func (c *Cluster) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for id, node := range c.nodes {
		log.Printf("Stopping node %s", id)
		node.Stop()
	}
}

// GetClusterInfo returns information about the cluster
func (c *Cluster) GetClusterInfo() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	info := make(map[string]interface{})
	nodes := make(map[string]interface{})

	for id, node := range c.nodes {
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
	info["total_nodes"] = len(c.nodes)

	// Find leader
	for _, node := range c.nodes {
		if node.IsLeader() {
			info["leader"] = node.GetID()
			break
		}
	}

	return info
}
