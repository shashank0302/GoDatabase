// Package storage provides storage engine implementations for the database.
// It includes a custom B+Tree implementation and a BadgerDB wrapper.
package storage

// Storage defines the interface for storage operations
// Any storage engine implementation must provide these methods.
type Storage interface {
	// Put stores a key-value pair in the storage engine.
	// Returns an error if the operation fails.
	Put(key, value []byte) error
	
	// Get retrieves a value for a given key from the storage engine.
	// Returns the value and an error (which will be non-nil if the key doesn't exist).
	Get(key []byte) ([]byte, error)
	
	// Delete removes a key-value pair from the storage engine.
	// Returns an error if the operation fails or the key doesn't exist.
	Delete(key []byte) error
	
	// Close closes the storage engine, flushing any pending changes to disk
	// and releasing any resources. Returns an error if the operation fails.
	Close() error
	
	// Size returns the number of key-value pairs in the storage engine.
	Size() int
}

// StorageType represents the type of storage to use.
// It's used to select between different storage engine implementations.
type StorageType string

const (
	// CustomStorage uses our own B+Tree implementation for storage.
	// This implementation is in the internal/btree package.
	CustomStorage StorageType = "custom"
	
	// BadgerStorageType uses BadgerDB, a third-party key-value store.
	// This is a wrapper around github.com/dgraph-io/badger/v3.
	BadgerStorageType StorageType = "badger"
)

// NewStorage creates a new storage instance of the specified type.
// This factory function returns the appropriate storage implementation based on the type.
// Parameters:
//   - storageType: The type of storage to create (CustomStorage or BadgerStorageType)
//   - path: The path to the storage file/directory
//
// Returns:
//   - A Storage instance
//   - An error if the creation fails
func NewStorage(storageType StorageType, path string) (Storage, error) {
	switch storageType {
	case CustomStorage:
		return NewStorageEngine(path)
	case BadgerStorageType:
		return NewBadgerStorage(path)
	default:
		return nil, ErrInvalidStorageType
	}
} 