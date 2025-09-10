package raft

// RequestVoteRequest represents a request vote RPC
type RequestVoteRequest struct {
	Term         int    // candidate's term
	CandidateID  string // candidate requesting vote
	LastLogIndex int    // index of candidate's last log entry
	LastLogTerm  int    // term of candidate's last log entry
}

// RequestVoteResponse represents a request vote RPC response
type RequestVoteResponse struct {
	Term        int  // currentTerm, for candidate to update itself
	VoteGranted bool // true means candidate received vote
}

// AppendEntriesRequest represents an append entries RPC
type AppendEntriesRequest struct {
	Term         int        // leader's term
	LeaderID     string     // so follower can redirect clients
	PrevLogIndex int        // index of log entry immediately preceding new ones
	PrevLogTerm  int        // term of prevLogIndex entry
	Entries      []LogEntry // log entries to store (empty for heartbeat)
	LeaderCommit int        // leader's commitIndex
}

// AppendEntriesResponse represents an append entries RPC response
type AppendEntriesResponse struct {
	Term    int  // currentTerm, for leader to update itself
	Success bool // true if follower contained entry matching prevLogIndex and prevLogTerm
}

// ClientRequest represents a client request to the Raft cluster
type ClientRequest struct {
	Operation string // "put", "get", "delete"
	Key       []byte
	Value     []byte
	Response  chan ClientResponse
}

// ClientResponse represents a response to a client request
type ClientResponse struct {
	Success bool
	Value   []byte
	Error   error
}
