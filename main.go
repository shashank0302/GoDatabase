// main.go
package main

import (
	"fmt"
	"log"
	"os"
	"godatabase/internal/storage"
)

func main() {
	// Example using BadgerDB storage
	fmt.Println("Using BadgerDB storage:")
	useBadgerDB()
	
	// Example using Custom B+Tree storage
	fmt.Println("\nUsing Custom B+Tree storage:")
	useCustomStorage()
}

func useBadgerDB() {
	// BadgerDB stores data in a directory
	dbPath := "test.db"
	defer os.RemoveAll(dbPath) // Clean up after test
	
	// Create a new BadgerDB storage instance
	s, err := storage.NewStorage(storage.BadgerStorageType, dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	// Example operations
	basicOperations(s)
}

func useCustomStorage() {
	// Custom storage engine stores data in a file
	dbPath := "custom_test.db"
	defer os.Remove(dbPath) // Clean up after test
	
	// Create a new Custom B+Tree storage instance
	s, err := storage.NewStorage(storage.CustomStorage, dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	// Example operations
	basicOperations(s)
}

func basicOperations(s storage.Storage) {
	key := []byte("hello")
	val := []byte("world")
	
	// Put a key-value pair
	if err := s.Put(key, val); err != nil {
		log.Printf("Put error: %v", err)
	}
	
	// Get the value for a key
	got, err := s.Get(key)
	if err != nil {
		log.Printf("Get error: %v", err)
	} else {
		fmt.Printf("Got: %s\n", got)
	}
	
	// Delete a key-value pair
	if err := s.Delete(key); err != nil {
		log.Printf("Delete error: %v", err)
	} else {
		fmt.Println("Successfully deleted key")
	}
	
	// Verify the key is gone
	_, err = s.Get(key)
	if err != nil {
		fmt.Println("Key no longer exists (expected)")
	}
}
