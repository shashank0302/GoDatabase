package raft

import (
	"log"
	"net"
	"net/rpc"
	"time"
)

// RaftRPC represents the RPC server for Raft communication
type RaftRPC struct {
	node *RaftNode
}

// RequestVote handles vote requests from candidates
func (r *RaftRPC) RequestVote(req RequestVoteRequest, resp *RequestVoteResponse) error {
	r.node.mu.Lock()
	defer r.node.mu.Unlock()

	log.Printf("Node %s received vote request from %s for term %d", r.node.id, req.CandidateID, req.Term)

	// Reply false if term < currentTerm
	if req.Term < r.node.currentTerm {
		resp.Term = r.node.currentTerm
		resp.VoteGranted = false
		return nil
	}

	// If RPC request or response contains term T > currentTerm: set currentTerm = T, convert to follower
	if req.Term > r.node.currentTerm {
		r.node.currentTerm = req.Term
		r.node.state = Follower
		r.node.votedFor = ""
	}

	// If votedFor is null or candidateId, and candidate's log is at least as up-to-date as receiver's log, grant vote
	if (r.node.votedFor == "" || r.node.votedFor == req.CandidateID) && r.isLogUpToDate(req.LastLogIndex, req.LastLogTerm) {
		r.node.votedFor = req.CandidateID
		r.node.lastHeartbeat = time.Now()
		resp.Term = r.node.currentTerm
		resp.VoteGranted = true
		log.Printf("Node %s granted vote to %s", r.node.id, req.CandidateID)
	} else {
		resp.Term = r.node.currentTerm
		resp.VoteGranted = false
		log.Printf("Node %s denied vote to %s", r.node.id, req.CandidateID)
	}

	return nil
}

// AppendEntries handles append entries requests from leaders
func (r *RaftRPC) AppendEntries(req AppendEntriesRequest, resp *AppendEntriesResponse) error {
	r.node.mu.Lock()
	defer r.node.mu.Unlock()

	log.Printf("Node %s received append entries from %s for term %d", r.node.id, req.LeaderID, req.Term)

	// Reply false if term < currentTerm
	if req.Term < r.node.currentTerm {
		resp.Term = r.node.currentTerm
		resp.Success = false
		return nil
	}

	// If RPC request or response contains term T > currentTerm: set currentTerm = T, convert to follower
	if req.Term > r.node.currentTerm {
		r.node.currentTerm = req.Term
		r.node.state = Follower
		r.node.votedFor = ""
	}

	// Update last heartbeat
	r.node.lastHeartbeat = time.Now()

	// If this is a heartbeat (no entries), just return success
	if len(req.Entries) == 0 {
		resp.Term = r.node.currentTerm
		resp.Success = true
		return nil
	}

	// Reply false if log doesn't contain an entry at prevLogIndex whose term matches prevLogTerm
	if !r.logContainsEntry(req.PrevLogIndex, req.PrevLogTerm) {
		resp.Term = r.node.currentTerm
		resp.Success = false
		return nil
	}

	// If an existing entry conflicts with a new one (same index but different terms), delete the existing entry and all that follow it
	conflictIndex := -1
	for i, entry := range req.Entries {
		logIndex := req.PrevLogIndex + 1 + i
		if logIndex < len(r.node.log) && r.node.log[logIndex].Term != entry.Term {
			conflictIndex = logIndex
			break
		}
	}

	if conflictIndex != -1 {
		// Truncate log from conflict index
		r.node.log = r.node.log[:conflictIndex]
	}

	// Append any new entries not already in the log
	for _, entry := range req.Entries {
		entry.Index = len(r.node.log) + 1
		r.node.log = append(r.node.log, entry)
	}

	// If leaderCommit > commitIndex, set commitIndex = min(leaderCommit, index of last new entry)
	if req.LeaderCommit > r.node.commitIndex {
		lastNewEntryIndex := len(r.node.log)
		if req.LeaderCommit < lastNewEntryIndex {
			r.node.commitIndex = req.LeaderCommit
		} else {
			r.node.commitIndex = lastNewEntryIndex
		}
	}

	// Apply committed entries
	r.applyCommittedEntries()

	resp.Term = r.node.currentTerm
	resp.Success = true
	return nil
}

// isLogUpToDate checks if the candidate's log is at least as up-to-date as this node's log
func (r *RaftRPC) isLogUpToDate(candidateLastIndex, candidateLastTerm int) bool {
	lastIndex := len(r.node.log)
	lastTerm := r.node.getLastLogTerm()

	// Raft determines which of two logs is more up-to-date by comparing the index and term of the last entries in the logs.
	// If the logs have last entries with different terms, then the log with the later term is more up-to-date.
	// If the logs end with the same term, then whichever log is longer is more up-to-date.
	return candidateLastTerm > lastTerm || (candidateLastTerm == lastTerm && candidateLastIndex >= lastIndex)
}

// logContainsEntry checks if the log contains an entry at the given index with the given term
func (r *RaftRPC) logContainsEntry(index, term int) bool {
	if index == 0 {
		return true // Special case for empty log
	}
	if index > len(r.node.log) {
		return false
	}
	return r.node.log[index-1].Term == term
}

// applyCommittedEntries applies all committed entries to the state machine
func (r *RaftRPC) applyCommittedEntries() {
	for r.node.lastApplied < r.node.commitIndex {
		r.node.lastApplied++
		entry := r.node.log[r.node.lastApplied-1]

		// Apply the command to the storage
		switch string(entry.Command[:4]) { // First 4 bytes indicate operation
		case "PUT ":
			// Parse key-value from command
			keyValue := entry.Command[4:]
			// Find the separator (assuming it's a space)
			spaceIndex := -1
			for i, b := range keyValue {
				if b == ' ' {
					spaceIndex = i
					break
				}
			}
			if spaceIndex > 0 {
				key := keyValue[:spaceIndex]
				value := keyValue[spaceIndex+1:]
				r.node.storage.Put(key, value)
			}
		case "DEL ":
			key := entry.Command[4:]
			r.node.storage.Delete(key)
		}
	}
}

// StartRPCServer starts the RPC server for this node
func (n *RaftNode) StartRPCServer() error {
	rpcServer := rpc.NewServer()
	raftRPC := &RaftRPC{node: n}

	err := rpcServer.Register(raftRPC)
	if err != nil {
		return err
	}

	// Use a simpler address format
	address := "localhost" + n.address
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	log.Printf("Raft RPC server listening on %s", address)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-n.ctx.Done():
					return
				default:
					log.Printf("Failed to accept connection: %v", err)
					continue
				}
			}
			go rpcServer.ServeConn(conn)
		}
	}()

	return nil
}

// sendRequestVote sends a vote request to a peer
func (n *RaftNode) sendRequestVote(peerAddr string, req RequestVoteRequest) (*RequestVoteResponse, error) {
	client, err := rpc.Dial("tcp", peerAddr)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	var resp RequestVoteResponse
	err = client.Call("RaftRPC.RequestVote", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// sendAppendEntries sends an append entries request to a peer
func (n *RaftNode) sendAppendEntries(peerAddr string, req AppendEntriesRequest) (*AppendEntriesResponse, error) {
	client, err := rpc.Dial("tcp", peerAddr)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	var resp AppendEntriesResponse
	err = client.Call("RaftRPC.AppendEntries", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
