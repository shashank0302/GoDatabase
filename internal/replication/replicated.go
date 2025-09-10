package replication

import (
	"errors"
	"log"
	"sync"
	
	"godatabase/internal/storage"
	"godatabase/pkg/client"
)

// ReplicatedStorage implements storage with replication to multiple nodes
type ReplicatedStorage struct {
	primary   storage.Storage
	replicas  []storage.Storage
	mu        sync.RWMutex
	asyncMode bool // If true, replicate asynchronously
}

// NewReplicatedStorage creates a new replicated storage
func NewReplicatedStorage(primary storage.Storage, replicaAddrs []string, asyncMode bool) (*ReplicatedStorage, error) {
	rs := &ReplicatedStorage{
		primary:   primary,
		replicas:  make([]storage.Storage, 0, len(replicaAddrs)),
		asyncMode: asyncMode,
	}
	
	// Connect to replicas
	for _, addr := range replicaAddrs {
		replica, err := client.New(addr)
		if err != nil {
			log.Printf("Failed to connect to replica %s: %v", addr, err)
			// Continue with other replicas
			continue
		}
		rs.replicas = append(rs.replicas, replica)
	}
	
	if len(rs.replicas) == 0 && len(replicaAddrs) > 0 {
		return nil, errors.New("failed to connect to any replica")
	}
	
	return rs, nil
}

// Put stores a key-value pair in primary and replicates to backups
func (rs *ReplicatedStorage) Put(key, value []byte) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	
	// Write to primary first
	if err := rs.primary.Put(key, value); err != nil {
		return err
	}
	
	// Replicate to backups
	if rs.asyncMode {
		// Asynchronous replication
		for _, replica := range rs.replicas {
			go func(r storage.Storage) {
				if err := r.Put(key, value); err != nil {
					log.Printf("Failed to replicate PUT to backup: %v", err)
				}
			}(replica)
		}
	} else {
		// Synchronous replication
		var wg sync.WaitGroup
		errChan := make(chan error, len(rs.replicas))
		
		for _, replica := range rs.replicas {
			wg.Add(1)
			go func(r storage.Storage) {
				defer wg.Done()
				if err := r.Put(key, value); err != nil {
					errChan <- err
				}
			}(replica)
		}
		
		wg.Wait()
		close(errChan)
		
		// Check for errors
		for err := range errChan {
			if err != nil {
				// At least one replica failed
				// In production, you might want to handle this differently
				log.Printf("Replication error: %v", err)
			}
		}
	}
	
	return nil
}

// Get retrieves a value from the primary
func (rs *ReplicatedStorage) Get(key []byte) ([]byte, error) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	
	// Read from primary
	value, err := rs.primary.Get(key)
	if err == nil {
		return value, nil
	}
	
	// If primary fails, try replicas (read repair)
	for _, replica := range rs.replicas {
		if value, err := replica.Get(key); err == nil {
			// Found in replica, repair primary
			go rs.primary.Put(key, value)
			return value, nil
		}
	}
	
	return nil, errors.New("key not found")
}

// Delete removes a key from primary and replicas
func (rs *ReplicatedStorage) Delete(key []byte) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	
	// Delete from primary first
	if err := rs.primary.Delete(key); err != nil {
		return err
	}
	
	// Delete from replicas
	if rs.asyncMode {
		// Asynchronous deletion
		for _, replica := range rs.replicas {
			go func(r storage.Storage) {
				if err := r.Delete(key); err != nil {
					log.Printf("Failed to replicate DELETE to backup: %v", err)
				}
			}(replica)
		}
	} else {
		// Synchronous deletion
		var wg sync.WaitGroup
		for _, replica := range rs.replicas {
			wg.Add(1)
			go func(r storage.Storage) {
				defer wg.Done()
				if err := r.Delete(key); err != nil {
					log.Printf("Failed to delete from replica: %v", err)
				}
			}(replica)
		}
		wg.Wait()
	}
	
	return nil
}

// Close closes all connections
func (rs *ReplicatedStorage) Close() error {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	
	// Close primary
	if err := rs.primary.Close(); err != nil {
		log.Printf("Error closing primary: %v", err)
	}
	
	// Close replicas
	for _, replica := range rs.replicas {
		if err := replica.Close(); err != nil {
			log.Printf("Error closing replica: %v", err)
		}
	}
	
	return nil
}

// Size returns the size from the primary
func (rs *ReplicatedStorage) Size() int {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	
	return rs.primary.Size()
} 