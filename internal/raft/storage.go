package raft

import (
	"fmt"
	"sync"
	"time"
)

// RaftStorage implements the storage.Storage interface using Raft consensus
type RaftStorage struct {
	cluster *GlobalCluster
	nodeID  string
	mu      sync.RWMutex
}

// NewRaftStorage creates a new Raft-based storage
func NewRaftStorage(cluster *GlobalCluster, nodeID string) *RaftStorage {
	return &RaftStorage{
		cluster: cluster,
		nodeID:  nodeID,
	}
}

// Put stores a key-value pair using Raft consensus
func (rs *RaftStorage) Put(key, value []byte) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	node, err := rs.cluster.GetNode(rs.nodeID)
	if err != nil {
		return fmt.Errorf("failed to get node: %v", err)
	}

	// Only the leader can handle writes
	if !node.IsLeader() {
		// Try to find the leader
		leader, err := rs.cluster.GetLeader()
		if err != nil {
			// No leader found, wait a bit and retry
			rs.mu.Unlock()
			time.Sleep(100 * time.Millisecond)
			rs.mu.Lock()

			// Try again
			leader, err = rs.cluster.GetLeader()
			if err != nil {
				return fmt.Errorf("no leader available: %v", err)
			}
		}

		// Redirect to leader (in a real implementation, you'd forward the request)
		return fmt.Errorf("not the leader, leader is at %s", leader.GetAddress())
	}

	return node.Put(key, value)
}

// Get retrieves a value for a key
func (rs *RaftStorage) Get(key []byte) ([]byte, error) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	node, err := rs.cluster.GetNode(rs.nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %v", err)
	}

	return node.Get(key)
}

// Delete removes a key-value pair using Raft consensus
func (rs *RaftStorage) Delete(key []byte) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	node, err := rs.cluster.GetNode(rs.nodeID)
	if err != nil {
		return fmt.Errorf("failed to get node: %v", err)
	}

	// Only the leader can handle writes
	if !node.IsLeader() {
		// Try to find the leader
		leader, err := rs.cluster.GetLeader()
		if err != nil {
			// No leader found, wait a bit and retry
			rs.mu.Unlock()
			time.Sleep(100 * time.Millisecond)
			rs.mu.Lock()

			// Try again
			leader, err = rs.cluster.GetLeader()
			if err != nil {
				return fmt.Errorf("no leader available: %v", err)
			}
		}

		// Redirect to leader (in a real implementation, you'd forward the request)
		return fmt.Errorf("not the leader, leader is at %s", leader.GetAddress())
	}

	return node.Delete(key)
}

// Close closes the Raft storage
func (rs *RaftStorage) Close() error {
	// The cluster manages the lifecycle of nodes
	// Individual storage instances don't need to close the cluster
	return nil
}

// Size returns the number of keys (not implemented for Raft storage)
func (rs *RaftStorage) Size() int {
	// This would require a special Raft command to count keys
	// For now, return -1 to indicate not supported
	return -1
}

// GetClusterInfo returns information about the Raft cluster
func (rs *RaftStorage) GetClusterInfo() map[string]interface{} {
	return rs.cluster.GetClusterInfo()
}

// IsLeader returns true if this node is the leader
func (rs *RaftStorage) IsLeader() bool {
	node, err := rs.cluster.GetNode(rs.nodeID)
	if err != nil {
		return false
	}
	return node.IsLeader()
}

// GetLeaderAddress returns the address of the current leader
func (rs *RaftStorage) GetLeaderAddress() (string, error) {
	leader, err := rs.cluster.GetLeader()
	if err != nil {
		return "", err
	}
	return leader.GetAddress(), nil
}
