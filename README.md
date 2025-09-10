# GeoCacheGoDB - Distributed Key-Value Store

A distributed, ACID-compliant key-value store in Go, designed to learn and implement distributed systems concepts with geo-distributed caching capabilities.

## Project Status

### ✅ Single-Node Implementation (COMPLETE)
- **B+Tree Storage**: Custom implementation with 4KB pages
- **Storage Interface**: Clean API for key-value operations
- **BadgerDB Integration**: Production-ready storage alternative
- **Concurrency**: Mutex-based thread safety
- **Unit Tests**: 90% coverage

### ✅ Distributed System Implementation (COMPLETE)
- **Raft Consensus**: Full Raft algorithm implementation
- **Leader Election**: Automatic leader election with timeouts
- **Log Replication**: Strong consistency through replicated logs
- **Split-brain Protection**: Only one leader at a time
- **Fault Tolerance**: Cluster continues with majority of nodes
- **gRPC Communication**: High-performance RPC layer
- **Client Libraries**: Both gRPC and TCP clients

## 🚀 Distributed System Implementation Plan

### Phase 1: Basic Distributed System (Start Here!)
1. **Simple TCP Network Layer**
   ```go
   // internal/network/
   ├── server.go      // TCP server
   ├── client.go      // TCP client
   └── protocol.go    // Message format
   ```

2. **Basic Replication**
   - Primary-backup replication
   - Simple leader election
   - Basic failover

3. **Client Library**
   ```go
   // pkg/client/
   └── client.go      // Client SDK
   ```

### Phase 2: Production-Ready Distribution
1. **gRPC Integration**
   ```go
   // internal/rpc/
   ├── proto/
   │   └── storage.proto
   ├── server.go
   └── client.go
   ```

2. **Raft Consensus**
   ```go
   // internal/raft/
   ├── node.go
   ├── log.go
   └── state.go
   ```

3. **Geo-Distribution Features**
   - Regional replicas
   - Geo-aware routing
   - Cross-region consistency

### Phase 3: Advanced Features
1. **Sharding**
   - Consistent hashing
   - Dynamic resharding
   - Shard migration

2. **Transactions**
   - Two-phase commit
   - Distributed transactions
   - MVCC (optional)

## Quick Start

### Raft Cluster (Recommended)

```bash
# Build the system
go build ./cmd/raft-server
go build ./cmd/client

# Start a 3-node Raft cluster
./examples/run-raft-cluster.sh

# In another terminal, test the cluster
./client -addr localhost:50051
> put hello world
> get hello
```

### Basic Replication (Legacy)

```bash
# Build the system
go build ./cmd/server
go build ./cmd/client

# Start a 3-node cluster with basic replication
./examples/run-cluster.sh

# In another terminal, test the cluster
./client -addr localhost:8080
> put hello world
> get hello
```

## Quick Start for Distributed Development

### Step 1: Create Network Layer
```go
// Example: Simple TCP server
package network

import (
    "net"
    "godatabase/internal/storage"
)

type Server struct {
    storage storage.Storage
    addr    string
}

func (s *Server) Start() error {
    listener, err := net.Listen("tcp", s.addr)
    if err != nil {
        return err
    }
    
    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }
        go s.handleConnection(conn)
    }
}
```

### Step 2: Add Replication
```go
// Example: Primary-backup replication
type ReplicatedStorage struct {
    primary storage.Storage
    backups []storage.Storage
}

func (r *ReplicatedStorage) Put(key, value []byte) error {
    // Write to primary
    if err := r.primary.Put(key, value); err != nil {
        return err
    }
    
    // Replicate to backups
    for _, backup := range r.backups {
        go backup.Put(key, value) // Async replication
    }
    
    return nil
}
```

## Current Architecture

```
godatabase/
├── internal/
│   ├── btree/         # Custom B+Tree (working)
│   ├── storage/       # Storage interface (working)
│   └── network/       # TODO: Network layer
├── pkg/
│   └── client/        # TODO: Client library
└── cmd/
    ├── server/        # TODO: Server binary
    └── client/        # TODO: CLI client
```

## Why This Approach Works

1. **Storage Layer is Abstracted**: Your `Storage` interface hides implementation details
2. **Focus on Distributed Concepts**: Use BadgerDB and focus on consensus, replication, and networking
3. **Iterative Development**: Start simple (TCP + primary-backup) then add complexity (gRPC + Raft)
4. **Learning-Oriented**: Each phase teaches specific distributed systems concepts

## Next Steps

1. **Create a simple TCP server** that wraps your storage
2. **Implement basic replication** (primary-backup)
3. **Add failure detection** and failover
4. **Then move to Raft** for proper consensus

This approach lets you learn distributed systems concepts without getting stuck on database internals!

## Code Organization

```
godatabase/
├── cmd/
│   └── test/           # Test runner
├── internal/
│   ├── btree/          # Original B+Tree implementation
│   │   ├── btree.go    # Tree operations
│   │   ├── node.go     # Node structure
│   │   └── btree_test.go
│   └── storage/        # Custom storage engine
│       ├── interface.go
│       ├── badger.go
│       ├── engine.go
│       └── errors.go
├── main.go
└── README.md
```

## Future Directory Structure

```
godatabase/
├── cmd/
│   ├── server/        # Server binary
│   └── client/        # Client binary
├── internal/
│   ├── btree/         # B+Tree implementation
│   ├── storage/       # Storage engine
│   ├── rpc/           # gRPC implementation
│   ├── raft/          # Raft consensus
│   ├── transaction/   # Transaction management
│   └── sharding/      # Sharding implementation
├── pkg/
│   └── client/        # Client library
└── deployments/       # Docker and Kubernetes configs
```

## Dependencies

Current:
```go
require (
    github.com/dgraph-io/badger/v3 v3.2103.5
)
```

Future:
```go
require (
    google.golang.org/grpc v1.58.0
    google.golang.org/protobuf v1.31.0
    go.etcd.io/etcd/raft/v3 v3.5.9
    github.com/hashicorp/raft v1.5.0
)
```

## Live Example: Running the Server and Clients

### Step 1: Start the Server

Open a terminal and run:
```bash
go run cmd/server/main.go
```
You should see:
```
2025/06/11 15:32:38 Starting gRPC server on :50051
```
This means your server is running and listening on port 50051.

### Step 2: Open Two Client Terminals

#### **Terminal 1 (Client 1)**
Run:
```bash
go run cmd/client/main.go
```
You'll see:
```
GeoCacheGoDB Client (type 'help' for commands)
>
```

#### **Terminal 2 (Client 2)**
Run the same command:
```bash
go run cmd/client/main.go
```
You'll see the same prompt:
```
GeoCacheGoDB Client (type 'help' for commands)
>
```

### Step 3: Interact with the Server

#### **In Terminal 1 (Client 1)**
- **Put a key-value pair:**
  ```
  > put hello world
  OK
  ```
- **Get the value:**
  ```
  > get hello
  world
  ```

#### **In Terminal 2 (Client 2)**
- **Get the same key:**
  ```
  > get hello
  world
  ```
  You'll see the same value because both clients are connected to the same server.

- **Put another key-value pair:**
  ```
  > put foo bar
  OK
  ```

#### **Back in Terminal 1 (Client 1)**
- **Get the new key:**
  ```
  > get foo
  bar
  ```
  You'll see the value set by Client 2.

### Step 4: Delete a Key

#### **In Terminal 1 (Client 1)**
- **Delete a key:**
  ```
  > delete hello
  OK
  ```

#### **In Terminal 2 (Client 2)**
- **Try to get the deleted key:**
  ```
  > get hello
  Error: key not found: ...
  ```
  The key is gone for both clients.

---

### **What's Happening Behind the Scenes**

1. **Server:**  
   - Runs a gRPC server on port 50051.
   - Handles requests from multiple clients simultaneously.

2. **Clients:**  
   - Connect to the server using gRPC.
   - Each client operates independently but interacts with the same data.

3. **Concurrency:**  
   - gRPC handles multiple connections automatically.
   - If two clients try to write to the same key simultaneously, the server processes them in order.

---

### **Try It Yourself!**

1. Start the server.
2. Open two client terminals.
3. Use `put`, `get`, and `delete` commands in both clients.
4. Observe how changes made by one client are visible to the other.




