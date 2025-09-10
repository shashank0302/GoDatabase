package raft

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"godatabase/internal/storage"
)

// NodeState represents the state of a Raft node
type NodeState int

const (
	Follower NodeState = iota
	Candidate
	Leader
)

func (s NodeState) String() string {
	switch s {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	default:
		return "Unknown"
	}
}

// LogEntry represents a single entry in the Raft log
type LogEntry struct {
	Term    int
	Index   int
	Command []byte
}

// RaftNode represents a single Raft node
type RaftNode struct {
	// Node identification
	id      string
	address string

	// Persistent state (updated on stable storage before responding to RPCs)
	currentTerm int
	votedFor    string
	log         []LogEntry

	// Volatile state on all servers
	commitIndex int
	lastApplied int

	// Volatile state on leaders (reinitialized after election)
	nextIndex  map[string]int
	matchIndex map[string]int

	// Node state
	state NodeState

	// Cluster configuration
	peers map[string]string // peer_id -> address

	// Storage interface
	storage storage.Storage

	// Channels for communication
	requestVoteChan   chan RequestVoteRequest
	appendEntriesChan chan AppendEntriesRequest
	clientRequestChan chan ClientRequest
	stopChan          chan struct{}

	// Mutex for thread safety
	mu sync.RWMutex

	// Election timeout
	electionTimeout time.Duration
	lastHeartbeat   time.Time

	// Heartbeat interval for leaders
	heartbeatInterval time.Duration

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

// NewRaftNode creates a new Raft node
func NewRaftNode(id, address string, peers map[string]string, storage storage.Storage) *RaftNode {
	ctx, cancel := context.WithCancel(context.Background())

	return &RaftNode{
		id:                id,
		address:           address,
		peers:             peers,
		storage:           storage,
		state:             Follower,
		currentTerm:       0,
		votedFor:          "",
		log:               make([]LogEntry, 0),
		commitIndex:       0,
		lastApplied:       0,
		nextIndex:         make(map[string]int),
		matchIndex:        make(map[string]int),
		requestVoteChan:   make(chan RequestVoteRequest, 100),
		appendEntriesChan: make(chan AppendEntriesRequest, 100),
		clientRequestChan: make(chan ClientRequest, 100),
		stopChan:          make(chan struct{}),
		electionTimeout:   time.Duration(150+rand.Intn(150)) * time.Millisecond, // 150-300ms
		heartbeatInterval: 50 * time.Millisecond,
		ctx:               ctx,
		cancel:            cancel,
	}
}

// Start starts the Raft node
func (n *RaftNode) Start() error {
	log.Printf("Starting Raft node %s on %s", n.id, n.address)

	// Start the main event loop
	go n.run()

	// Start the election timer
	go n.electionTimer()

	// Start heartbeat if leader
	go n.heartbeatTimer()

	return nil
}

// Stop stops the Raft node
func (n *RaftNode) Stop() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.ctx.Err() != nil {
		return // Already stopped
	}

	log.Printf("Stopping Raft node %s", n.id)
	n.cancel()

	select {
	case <-n.stopChan:
		// Channel already closed
	default:
		close(n.stopChan)
	}
}

// run is the main event loop
func (n *RaftNode) run() {
	for {
		select {
		case <-n.ctx.Done():
			return
		case req := <-n.requestVoteChan:
			n.handleRequestVote(req)
		case req := <-n.appendEntriesChan:
			n.handleAppendEntries(req)
		case req := <-n.clientRequestChan:
			n.handleClientRequest(req)
		}
	}
}

// electionTimer handles election timeouts
func (n *RaftNode) electionTimer() {
	for {
		select {
		case <-n.ctx.Done():
			return
		default:
			n.mu.Lock()
			state := n.state
			lastHeartbeat := n.lastHeartbeat
			n.mu.Unlock()

			if state != Leader {
				timeout := n.electionTimeout
				if time.Since(lastHeartbeat) > timeout {
					n.startElection()
				}
			}

			time.Sleep(50 * time.Millisecond)
		}
	}
}

// heartbeatTimer sends heartbeats if this node is the leader
func (n *RaftNode) heartbeatTimer() {
	ticker := time.NewTicker(n.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.mu.RLock()
			state := n.state
			n.mu.RUnlock()

			if state == Leader {
				n.sendHeartbeats()
			}
		}
	}
}

// startElection starts a new election
func (n *RaftNode) startElection() {
	n.mu.Lock()
	defer n.mu.Unlock()

	log.Printf("Node %s starting election for term %d", n.id, n.currentTerm+1)

	// Transition to candidate
	n.state = Candidate
	n.currentTerm++
	n.votedFor = n.id
	n.lastHeartbeat = time.Now()

	// Reset election timeout
	n.electionTimeout = time.Duration(150+rand.Intn(150)) * time.Millisecond

	// Request votes from all peers
	votes := 1 // Vote for self
	totalVotes := len(n.peers) + 1

	for peerID, peerAddr := range n.peers {
		go func(id, addr string) {
			req := RequestVoteRequest{
				Term:         n.currentTerm,
				CandidateID:  n.id,
				LastLogIndex: len(n.log),
				LastLogTerm:  n.getLastLogTerm(),
			}

			resp, err := n.sendRequestVote(addr, req)
			if err != nil {
				log.Printf("Failed to send vote request to %s: %v", id, err)
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

			if resp.VoteGranted {
				votes++
				if votes > totalVotes/2 {
					n.becomeLeader()
				}
			}
		}(peerID, peerAddr)
	}
}

// becomeLeader transitions this node to leader state
func (n *RaftNode) becomeLeader() {
	log.Printf("Node %s became leader for term %d", n.id, n.currentTerm)

	n.state = Leader
	n.lastHeartbeat = time.Now()

	// Initialize nextIndex and matchIndex for all peers
	for peerID := range n.peers {
		n.nextIndex[peerID] = len(n.log) + 1
		n.matchIndex[peerID] = 0
	}

	// Send initial heartbeat
	n.sendHeartbeats()
}

// StepDown forces this node to step down from leader role
func (n *RaftNode) StepDown() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.state == Leader {
		log.Printf("Node %s stepping down from leader role", n.id)
		n.state = Follower
		n.votedFor = ""
		n.lastHeartbeat = time.Now()
	}
}

// sendHeartbeats sends heartbeat messages to all peers
func (n *RaftNode) sendHeartbeats() {
	n.mu.RLock()
	term := n.currentTerm
	peers := make(map[string]string)
	for k, v := range n.peers {
		peers[k] = v
	}
	n.mu.RUnlock()

	for peerID, peerAddr := range peers {
		go func(id, addr string) {
			req := AppendEntriesRequest{
				Term:         term,
				LeaderID:     n.id,
				PrevLogIndex: 0,
				PrevLogTerm:  0,
				Entries:      []LogEntry{},
				LeaderCommit: n.commitIndex,
			}

			resp, err := n.sendAppendEntries(addr, req)
			if err != nil {
				log.Printf("Failed to send heartbeat to %s: %v", id, err)
				return
			}

			n.mu.Lock()
			defer n.mu.Unlock()

			if resp.Term > n.currentTerm {
				n.currentTerm = resp.Term
				n.state = Follower
				n.votedFor = ""
			}
		}(peerID, peerAddr)
	}
}

// getLastLogTerm returns the term of the last log entry
func (n *RaftNode) getLastLogTerm() int {
	if len(n.log) == 0 {
		return 0
	}
	return n.log[len(n.log)-1].Term
}

// GetState returns the current state of the node
func (n *RaftNode) GetState() (NodeState, int) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.state, n.currentTerm
}

// IsLeader returns true if this node is the leader
func (n *RaftNode) IsLeader() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.state == Leader
}

// GetAddress returns the address of this node
func (n *RaftNode) GetAddress() string {
	return n.address
}

// GetID returns the ID of this node
func (n *RaftNode) GetID() string {
	return n.id
}

// GetContext returns the context for this node
func (n *RaftNode) GetContext() context.Context {
	return n.ctx
}

// handleRequestVote handles vote requests
func (n *RaftNode) handleRequestVote(req RequestVoteRequest) {
	// This will be handled by the RPC server
}

// handleAppendEntries handles append entries requests
func (n *RaftNode) handleAppendEntries(req AppendEntriesRequest) {
	// This will be handled by the RPC server
}

// applyCommittedEntries applies all committed entries to the state machine
func (n *RaftNode) applyCommittedEntries() {
	for n.lastApplied < n.commitIndex {
		n.lastApplied++
		entry := n.log[n.lastApplied-1]

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
				n.storage.Put(key, value)
			}
		case "DEL ":
			key := entry.Command[4:]
			n.storage.Delete(key)
		}
	}
}
