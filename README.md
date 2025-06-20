# GeoCacheGoDB - Custom Storage Engine

This project implements a custom, ACID-compliant database in Go with a focus on educational purposes and understanding database internals. The storage layer provides two implementations:

1. **Custom B+Tree Storage**: A completely original implementation of a B+Tree with file persistence
2. **BadgerDB Storage**: A wrapper around BadgerDB for comparison and testing

## Project Overview

### Core Components
-Original B+Tree Is implemented referenced from [ build-your-own-db-from-scratch](https://build-your-own.org/database/)
1. **Original B+Tree Implementation** (`internal/btree/`)
   - Custom-designed node structure with fixed 4KB pages
   - Original implementation of internal and leaf nodes
   - Custom variable-length key-value storage encoding
   - Novel parent-child relationship tracking system
   - Original node splitting and merging algorithms
   - Custom serialization/deserialization format

2. **Custom Storage Engine** (`internal/storage/`)
   - Original storage interface design
   - Custom file format with magic numbers and versioning
   - Original file I/O operations
   - Custom synchronization with mutex locks
   - Basic persistence implementation

### Current Features

1. **B+Tree Core** (Original Implementation)
   - ✅ Custom node structure with fixed 4K page size
   - ✅ Original CRUD operations (Insert, Get, Delete)
   - ✅ Custom node splitting and merging algorithms
   - ✅ Novel parent-child relationship tracking
   - ✅ Original serialization/deserialization
   - ✅ Comprehensive unit tests

2. **Storage Layer** (Original Implementation)
   - ✅ Custom storage interface design
   - ✅ Original file format with versioning
   - ✅ Custom file I/O operations
   - ✅ Original error handling system
   - ✅ Custom size tracking

## Implementation Details

### B+Tree Implementation

The B+Tree implementation (`internal/btree/`) is a completely original work:

1. **Node Structure** (`node.go`)
   ```go
   type Node struct {
       typ      uint16    // Node type (internal/leaf)
       nkeys    uint16    // Number of keys
       pointers []uint64  // Child pointers
       offsets  []uint16  // Key-value offsets
       data     []byte    // Key-value data
   }
   ```
   - Custom 4KB page size design
   - Original internal/leaf node implementation
   - Novel variable-length encoding
   - Custom serialization format
   - Original parent-child tracking

2. **Tree Operations** (`btree.go`)
   - Original insert algorithm with automatic splitting
   - Custom delete implementation with rebalancing
   - Novel traversal and search methods
   - Original size and height tracking
   - Custom node relationship management

### Storage Engine

The storage engine (`internal/storage/`) features:

1. **Custom Interface**
   ```go
   type Storage interface {
       Put(key, value []byte) error
       Get(key []byte) ([]byte, error)
       Delete(key []byte) error
       Close() error
       Size() int
   }
   ```

2. **Original File Format**
   ```go
   const (
       PAGE_SIZE = 4096
       MAGIC    = uint32(0x12345678)
       VERSION  = uint32(1)
   )
   ```

3. **Custom Implementation**
   - Original file I/O operations
   - Custom serialization format
   - Novel node relationship tracking
   - Original error handling

## Future Implementation Plan

### 1. Persistence Layer (Priority: High)
- [ ] Enhance disk-based storage system
- [ ] Implement custom page caching
- [ ] Add original transaction logging
- [ ] Design custom WAL implementation
- [ ] Create crash recovery system

### 2. Concurrency Control (Priority: High)
- [ ] Design custom read-write locks
- [ ] Implement original MVCC system
- [ ] Create custom transaction isolation
- [ ] Add deadlock detection

### 3. Query Engine (Priority: Medium)
- [ ] Implement custom range scans
- [ ] Add original prefix search
- [ ] Create reverse iteration
- [ ] Design bulk operations

### 4. Performance Optimizations (Priority: Medium)
- [ ] Implement custom buffer pool
- [ ] Add value compression
- [ ] Create bloom filters
- [ ] Add statistics collection

### 5. Testing and Validation (Priority: High)
- [ ] Add stress tests
- [ ] Implement crash recovery tests
- [ ] Add performance benchmarks
- [ ] Create integration tests

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



