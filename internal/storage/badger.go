package storage

import (
	"github.com/dgraph-io/badger/v3"
)

// BadgerStorage implements the Storage interface using BadgerDB.
// BadgerDB is an embeddable, persistent, and fast key-value (KV) database.
// It's designed with a single point in mind: to provide a simple, 
// efficient, and embeddable key-value store for Go projects.
type BadgerStorage struct {
	db *badger.DB // The underlying BadgerDB instance
}

// NewBadgerStorage creates a new BadgerDB storage instance.
// It opens a BadgerDB database at the specified path.
// If the database doesn't exist, it will be created.
//
// Parameters:
//   - path: The directory where BadgerDB will store its data files
//
// Returns:
//   - A pointer to a BadgerStorage instance
//   - An error if the database couldn't be opened
func NewBadgerStorage(path string) (*BadgerStorage, error) {
	// Configure BadgerDB options
	opts := badger.DefaultOptions(path)
	opts.Logger = nil // Disable Badger's default logging
	
	// Open the database
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	
	return &BadgerStorage{db: db}, nil
}

// Put implements Storage.Put by storing a key-value pair in BadgerDB.
// It uses BadgerDB's transactional API to ensure atomicity.
//
// Parameters:
//   - key: The key as a byte slice
//   - value: The value as a byte slice
//
// Returns:
//   - An error if the operation fails
func (s *BadgerStorage) Put(key, value []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

// Get implements Storage.Get by retrieving a value for a given key.
// It uses BadgerDB's read-only transaction to retrieve the value.
//
// Parameters:
//   - key: The key to look up
//
// Returns:
//   - The value as a byte slice
//   - An error if the key doesn't exist or the operation fails
func (s *BadgerStorage) Get(key []byte) ([]byte, error) {
	var value []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		
		value, err = item.ValueCopy(nil)
		return err
	})
	
	return value, err
}

// Delete implements Storage.Delete by removing a key-value pair.
// It uses BadgerDB's transactional API to ensure atomicity.
//
// Parameters:
//   - key: The key to delete
//
// Returns:
//   - An error if the key doesn't exist or the operation fails
func (s *BadgerStorage) Delete(key []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// Close implements Storage.Close by properly closing the BadgerDB database.
// This ensures all pending writes are flushed to disk and resources are released.
//
// Returns:
//   - An error if the close operation fails
func (s *BadgerStorage) Close() error {
	return s.db.Close()
}

// Size implements Storage.Size by counting the number of keys in BadgerDB.
// Since BadgerDB doesn't provide a direct way to get the number of keys,
// this method iterates through all keys to count them.
//
// Returns:
//   - The number of key-value pairs in the database
func (s *BadgerStorage) Size() int {
	// BadgerDB doesn't provide a direct way to get the number of keys
	// This is a simple implementation that counts the keys
	var count int
	s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		
		for it.Rewind(); it.Valid(); it.Next() {
			count++
		}
		return nil
	})
	return count
} 