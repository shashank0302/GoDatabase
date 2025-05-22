package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTest(t *testing.T) (string, func()) {
	// Create temporary directory for test files
	testDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatal(err)
	}

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(testDir)
	}

	return testDir, cleanup
}

func testStorageImplementation(t *testing.T, storageType StorageType, path string) {
	// Create new storage instance
	s, err := NewStorage(storageType, path)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer s.Close()

	// Test Put
	t.Run("Put", func(t *testing.T) {
		testData := map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}

		for k, v := range testData {
			if err := s.Put([]byte(k), []byte(v)); err != nil {
				t.Errorf("Put failed for %s: %v", k, err)
			}
		}

		// Verify size
		if size := s.Size(); size != len(testData) {
			t.Errorf("Expected size %d, got %d", len(testData), size)
		}
	})

	// Test Get
	t.Run("Get", func(t *testing.T) {
		testData := map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}

		for k, v := range testData {
			value, err := s.Get([]byte(k))
			if err != nil {
				t.Errorf("Get failed for %s: %v", k, err)
				continue
			}
			if string(value) != v {
				t.Errorf("Expected value %s for key %s, got %s", v, k, string(value))
			}
		}

		// Test non-existent key
		_, err := s.Get([]byte("nonexistent"))
		if err == nil {
			t.Error("Expected error for non-existent key")
		}
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		// Delete a key
		if err := s.Delete([]byte("key2")); err != nil {
			t.Errorf("Delete failed: %v", err)
		}

		// Verify deletion
		_, err := s.Get([]byte("key2"))
		if err == nil {
			t.Error("Expected error for deleted key")
		}

		// Verify size
		if size := s.Size(); size != 2 {
			t.Errorf("Expected size 2 after deletion, got %d", size)
		}
	})

	// Test concurrent operations
	t.Run("Concurrent", func(t *testing.T) {
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(i int) {
				key := []byte{byte(i)}
				value := []byte{byte(i + 1)}

				// Put
				if err := s.Put(key, value); err != nil {
					t.Errorf("Concurrent Put failed: %v", err)
				}

				// Get
				if val, err := s.Get(key); err != nil {
					t.Errorf("Concurrent Get failed: %v", err)
				} else if val[0] != value[0] {
					t.Errorf("Concurrent Get returned wrong value: expected %d, got %d", value[0], val[0])
				}

				// Delete
				if err := s.Delete(key); err != nil {
					t.Errorf("Concurrent Delete failed: %v", err)
				}

				done <- true
			}(i)
		}

		// Wait for all goroutines to finish
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func TestCustomStorage(t *testing.T) {
	testDir, cleanup := setupTest(t)
	defer cleanup()

	path := filepath.Join(testDir, "custom.db")
	testStorageImplementation(t, CustomStorage, path)
}

func TestBadgerStorage(t *testing.T) {
	testDir, cleanup := setupTest(t)
	defer cleanup()

	path := filepath.Join(testDir, "badger.db")
	testStorageImplementation(t, BadgerStorageType, path)
} 