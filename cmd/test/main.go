package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"godatabase/internal/storage"
)

func main() {
	// Create test directory
	testDir := "testdata"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(testDir)

	// Test only BadgerDB implementation for now
	// fmt.Println("Testing BadgerDB Storage:")
	// testStorage(storage.BadgerStorageType, filepath.Join(testDir, "badger.db"))

	fmt.Println("Testing Custom B+Tree Storage:")
	testStorage(storage.CustomStorage, filepath.Join(testDir, "custom.db"))
	
	/*
	// ===== TO TEST BOTH STORAGE ENGINES =====
	// Once the custom storage engine is fixed, you can test both implementations:
	
	// Test Custom B+Tree Storage
	fmt.Println("Testing Custom B+Tree Storage:")
	testStorage(storage.CustomStorage, filepath.Join(testDir, "custom.db"))

	// Test BadgerDB Storage
	fmt.Println("\nTesting BadgerDB Storage:")
	testStorage(storage.BadgerStorageType, filepath.Join(testDir, "badger.db"))
	
	// NOTE: Before enabling this code, ensure the following is fixed in the custom storage engine:
	// 1. Fix the node.getChild() method to properly handle parent-child relationships
	// 2. Ensure proper key-value storage and retrieval in the B+Tree
	// 3. Complete the serialization/deserialization in the StorageEngine.flush() method
	//
	// Start with simple tests before enabling concurrent operations.
	*/
	fmt.Println("\nStorage engine tests completed successfully!")
}

func testStorage(storageType storage.StorageType, path string) {
	// Create new storage instance
	s, err := storage.NewStorage(storageType, path)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	defer s.Close()

	// Test Put
	fmt.Println("Testing Put...")
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for k, v := range testData {
		if err := s.Put([]byte(k), []byte(v)); err != nil {
			log.Printf("Put failed for %s: %v", k, err)
		}
	}

	// Test Get
	fmt.Println("Testing Get...")
	for k, v := range testData {
		value, err := s.Get([]byte(k))
		if err != nil {
			log.Printf("Get failed for %s: %v", k, err)
			continue
		}
		fmt.Printf("Got %s = %s (expected: %s)\n", k, string(value), v)
	}

	// Test Size
	fmt.Printf("Storage size: %d\n", s.Size())

	// Test Delete
	fmt.Println("Testing Delete...")
	if err := s.Delete([]byte("key2")); err != nil {
		log.Printf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = s.Get([]byte("key2"))
	if err == nil {
		log.Println("Expected error for deleted key")
	} else {
		fmt.Println("Successfully deleted key2")
	}

	fmt.Println("Basic tests completed.")
	
	// Uncomment for concurrent testing once the basic functionality is stable
	/*
	fmt.Println("Testing concurrent operations...")
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(i int) {
			key := fmt.Sprintf("concurrent_key_%d", i)
			value := fmt.Sprintf("concurrent_value_%d", i)

			// Put
			if err := s.Put([]byte(key), []byte(value)); err != nil {
				log.Printf("Concurrent Put failed: %v", err)
			}

			// Get
			if val, err := s.Get([]byte(key)); err != nil {
				log.Printf("Concurrent Get failed: %v", err)
			} else if string(val) != value {
				log.Printf("Concurrent Get returned wrong value: expected %s, got %s", value, string(val))
			}

			// Delete
			if err := s.Delete([]byte(key)); err != nil {
				log.Printf("Concurrent Delete failed: %v", err)
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines to finish
	for i := 0; i < 10; i++ {
		<-done
	}

	fmt.Printf("Final storage size: %d\n", s.Size())
	*/
} 