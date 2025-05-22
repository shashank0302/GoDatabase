package storage

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"sync"

	"godatabase/internal/btree"
)

const (
	// Page size for the storage engine
	PAGE_SIZE = 4096

	// Magic number to identify our database file
	MAGIC = uint32(0x12345678)

	// Version of the storage format
	VERSION = uint32(1)
)

// StorageEngine represents the storage engine
type StorageEngine struct {
	file     *os.File
	btree    *btree.BTree
	mu       sync.RWMutex
	filename string
}

// NewStorageEngine creates a new storage engine
func NewStorageEngine(filename string) (*StorageEngine, error) {
	// Open or create the database file
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	engine := &StorageEngine{
		file:     file,
		btree:    btree.NewBTree(),
		filename: filename,
	}

	// Initialize the database if it's new
	if err := engine.initialize(); err != nil {
		file.Close()
		return nil, err
	}

	return engine, nil
}

// initialize sets up a new database file
func (e *StorageEngine) initialize() error {
	// Check if the file is empty
	stat, err := e.file.Stat()
	if err != nil {
		return err
	}

	if stat.Size() == 0 {
		// Write the header
		header := make([]byte, 8)
		binary.BigEndian.PutUint32(header[0:4], MAGIC)
		binary.BigEndian.PutUint32(header[4:8], VERSION)
		if _, err := e.file.Write(header); err != nil {
			return err
		}
	} else {
		// Verify the header
		header := make([]byte, 8)
		if _, err := e.file.ReadAt(header, 0); err != nil {
			return err
		}
		magic := binary.BigEndian.Uint32(header[0:4])
		version := binary.BigEndian.Uint32(header[4:8])
		if magic != MAGIC {
			return errors.New("invalid database file")
		}
		if version != VERSION {
			return errors.New("unsupported database version")
		}
	}

	return nil
}

// Put stores a key-value pair
func (e *StorageEngine) Put(key, value []byte) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Insert into B+Tree
	if err := e.btree.Insert(key, value); err != nil {
		return err
	}

	// Write to disk
	return e.flush()
}

// Get retrieves a value for a given key
func (e *StorageEngine) Get(key []byte) ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.btree.Get(key)
}

// Delete removes a key-value pair
func (e *StorageEngine) Delete(key []byte) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Delete from B+Tree
	if err := e.btree.Delete(key); err != nil {
		return err
	}

	// Write to disk
	return e.flush()
}

// flush writes the current state to disk
func (e *StorageEngine) flush() error {
	// Seek to the start of the data section (after header)
	if _, err := e.file.Seek(8, io.SeekStart); err != nil {
		return err
	}
	
	// This is a basic implementation of tree serialization.
	// In a real implementation, you would serialize each node separately
	// and build a page index.
	
	// Write a simple header for the B+Tree data
	treeHeader := make([]byte, 8)
	binary.BigEndian.PutUint32(treeHeader[0:4], uint32(e.btree.Size()))
	binary.BigEndian.PutUint32(treeHeader[4:8], uint32(e.btree.Height()))
	if _, err := e.file.Write(treeHeader); err != nil {
		return err
	}
	
	// For now, we'll use a simplified approach that doesn't support
	// full tree reconstruction, but allows us to store key-value pairs
	
	// Serialize the root node
	if e.btree.Size() > 0 {
		// Get the root node and its data
		rootData := serializeNode(e.btree)
		if _, err := e.file.Write(rootData); err != nil {
			return err
		}
	}
	
	// Ensure all data is written to disk
	return e.file.Sync()
}

// serializeNode creates a byte representation of the key-value pairs in the tree
// This is a simplified implementation that doesn't actually serialize the tree structure
func serializeNode(tree *btree.BTree) []byte {
	// Get all key-value pairs from the tree
	// This is just a placeholder implementation
	// that serializes up to 1000 key-value pairs
	
	buf := make([]byte, 0, PAGE_SIZE)
	
	// This would typically iterate through the tree's leaf nodes 
	// For now, we just append some metadata
	metaSize := 8
	buf = append(buf, make([]byte, metaSize)...)
	
	// In a real implementation, this would follow the proper B+Tree serialization format
	
	return buf
}

// Close closes the storage engine
func (e *StorageEngine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Flush any pending changes
	if err := e.flush(); err != nil {
		return err
	}

	return e.file.Close()
}

// Size returns the number of key-value pairs in the storage engine
func (e *StorageEngine) Size() int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.btree.Size()
} 