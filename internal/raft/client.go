package raft

import (
	"bytes"
	"fmt"
	"log"
	"time"
)

// handleClientRequest handles client requests
func (n *RaftNode) handleClientRequest(req ClientRequest) {
	n.mu.RLock()
	state := n.state
	n.mu.RUnlock()

	// Only the leader can handle client requests
	if state != Leader {
		req.Response <- ClientResponse{
			Success: false,
			Error:   fmt.Errorf("not the leader"),
		}
		return
	}

	// Create log entry for the command
	var command []byte
	switch req.Operation {
	case "put":
		command = append([]byte("PUT "), req.Key...)
		command = append(command, ' ')
		command = append(command, req.Value...)
	case "delete":
		command = append([]byte("DEL "), req.Key...)
	default:
		req.Response <- ClientResponse{
			Success: false,
			Error:   fmt.Errorf("unknown operation: %s", req.Operation),
		}
		return
	}

	// Add entry to log
	n.mu.Lock()
	entry := LogEntry{
		Term:    n.currentTerm,
		Index:   len(n.log) + 1,
		Command: command,
	}
	n.log = append(n.log, entry)
	logIndex := len(n.log)
	n.mu.Unlock()

	// Replicate to followers
	success := n.replicateLogEntry(entry, logIndex)

	if success {
		// Apply the entry locally
		n.applyEntry(entry)

		// Send response
		if req.Operation == "get" {
			value, err := n.storage.Get(req.Key)
			req.Response <- ClientResponse{
				Success: true,
				Value:   value,
				Error:   err,
			}
		} else {
			req.Response <- ClientResponse{
				Success: true,
			}
		}
	} else {
		req.Response <- ClientResponse{
			Success: false,
			Error:   fmt.Errorf("failed to replicate to majority"),
		}
	}
}

// replicateLogEntry replicates a log entry to all followers
func (n *RaftNode) replicateLogEntry(entry LogEntry, logIndex int) bool {
	n.mu.RLock()
	term := n.currentTerm
	peers := make(map[string]string)
	for k, v := range n.peers {
		peers[k] = v
	}
	n.mu.RUnlock()

	successCount := 1 // Count self
	totalPeers := len(peers) + 1

	// Send append entries to all peers
	for peerID, peerAddr := range peers {
		go func(id, addr string) {
			req := AppendEntriesRequest{
				Term:         term,
				LeaderID:     n.id,
				PrevLogIndex: logIndex - 1,
				PrevLogTerm:  n.getPrevLogTerm(logIndex - 1),
				Entries:      []LogEntry{entry},
				LeaderCommit: n.commitIndex,
			}

			resp, err := n.sendAppendEntries(addr, req)
			if err != nil {
				log.Printf("Failed to replicate to %s: %v", id, err)
				return
			}

			n.mu.Lock()
			defer n.mu.Unlock()

			if resp.Term > n.currentTerm {
				n.currentTerm = resp.Term
				n.state = Follower
				n.votedFor = ""
				return
			}

			if resp.Success {
				n.matchIndex[id] = logIndex
				n.nextIndex[id] = logIndex + 1
				successCount++

				// Check if we have majority
				if successCount > totalPeers/2 {
					// Update commit index
					n.commitIndex = logIndex
					n.applyCommittedEntries()
				}
			} else {
				// Decrement nextIndex and retry
				if n.nextIndex[id] > 0 {
					n.nextIndex[id]--
				}
			}
		}(peerID, peerAddr)
	}

	// Wait for majority (simplified - in practice, this should be more sophisticated)
	time.Sleep(100 * time.Millisecond)

	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.commitIndex >= logIndex
}

// getPrevLogTerm returns the term of the log entry at the given index
func (n *RaftNode) getPrevLogTerm(index int) int {
	if index == 0 {
		return 0
	}
	if index > len(n.log) {
		return 0
	}
	return n.log[index-1].Term
}

// applyEntry applies a single log entry to the state machine
func (n *RaftNode) applyEntry(entry LogEntry) {
	command := string(entry.Command)

	if len(command) < 4 {
		return
	}

	switch command[:4] {
	case "PUT ":
		// Parse key-value from command
		parts := bytes.Split(entry.Command[4:], []byte{' '})
		if len(parts) >= 2 {
			key := parts[0]
			value := parts[1]
			n.storage.Put(key, value)
		}
	case "DEL ":
		key := entry.Command[4:]
		n.storage.Delete(key)
	}
}

// SubmitRequest submits a client request to the Raft cluster
func (n *RaftNode) SubmitRequest(operation string, key, value []byte) ([]byte, error) {
	req := ClientRequest{
		Operation: operation,
		Key:       key,
		Value:     value,
		Response:  make(chan ClientResponse, 1),
	}

	select {
	case n.clientRequestChan <- req:
		// Request submitted
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout submitting request")
	}

	select {
	case resp := <-req.Response:
		if !resp.Success {
			return nil, resp.Error
		}
		return resp.Value, nil
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("timeout waiting for response")
	}
}

// Get retrieves a value from the cluster
func (n *RaftNode) Get(key []byte) ([]byte, error) {
	return n.SubmitRequest("get", key, nil)
}

// Put stores a key-value pair in the cluster
func (n *RaftNode) Put(key, value []byte) error {
	_, err := n.SubmitRequest("put", key, value)
	return err
}

// Delete removes a key from the cluster
func (n *RaftNode) Delete(key []byte) error {
	_, err := n.SubmitRequest("delete", key, nil)
	return err
}
