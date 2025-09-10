package main

import (
	"fmt"
	"log"
	"time"
	
	"godatabase/internal/replication"
	"godatabase/pkg/client"
)

func main() {
	fmt.Println("GeoCacheGoDB Distributed Demo")
	fmt.Println("==============================")
	
	// Connect to primary node
	primary, err := client.New("localhost:8080")
	if err != nil {
		log.Fatalf("Failed to connect to primary: %v", err)
	}
	defer primary.Close()
	
	// Create replicated storage with multiple replicas
	replicas := []string{"localhost:8081", "localhost:8082"}
	storage, err := replication.NewReplicatedStorage(primary, replicas, false) // Synchronous mode
	if err != nil {
		log.Fatalf("Failed to create replicated storage: %v", err)
	}
	defer storage.Close()
	
	fmt.Println("\n1. Writing data to distributed storage...")
	
	// Write some data
	data := map[string]string{
		"user:1":    "Alice",
		"user:2":    "Bob",
		"user:3":    "Charlie",
		"config:db": "production",
		"cache:key": "cached_value",
	}
	
	for key, value := range data {
		if err := storage.Put([]byte(key), []byte(value)); err != nil {
			log.Printf("Failed to put %s: %v", key, err)
		} else {
			fmt.Printf("  ✓ Stored: %s = %s\n", key, value)
		}
		time.Sleep(100 * time.Millisecond) // Small delay for demo
	}
	
	fmt.Println("\n2. Reading data from distributed storage...")
	
	// Read data back
	for key := range data {
		value, err := storage.Get([]byte(key))
		if err != nil {
			log.Printf("Failed to get %s: %v", key, err)
		} else {
			fmt.Printf("  ✓ Retrieved: %s = %s\n", key, string(value))
		}
	}
	
	fmt.Println("\n3. Testing direct replica access...")
	
	// Connect directly to a replica
	replica1, err := client.New("localhost:8081")
	if err != nil {
		log.Printf("Failed to connect to replica: %v", err)
	} else {
		defer replica1.Close()
		
		// Read from replica directly
		value, err := replica1.Get([]byte("user:1"))
		if err != nil {
			log.Printf("Failed to read from replica: %v", err)
		} else {
			fmt.Printf("  ✓ Direct replica read: user:1 = %s\n", string(value))
		}
	}
	
	fmt.Println("\n4. Demonstrating consistency...")
	
	// Update a value
	newValue := "Alice (Updated)"
	if err := storage.Put([]byte("user:1"), []byte(newValue)); err != nil {
		log.Printf("Failed to update: %v", err)
	} else {
		fmt.Printf("  ✓ Updated: user:1 = %s\n", newValue)
	}
	
	// Wait a bit for replication
	time.Sleep(500 * time.Millisecond)
	
	// Check all replicas
	fmt.Println("\n5. Verifying replication across all nodes...")
	nodes := map[string]string{
		"Primary (8080)": "localhost:8080",
		"Replica 1 (8081)": "localhost:8081",
		"Replica 2 (8082)": "localhost:8082",
	}
	
	for name, addr := range nodes {
		node, err := client.New(addr)
		if err != nil {
			log.Printf("  ✗ Failed to connect to %s: %v", name, err)
			continue
		}
		
		value, err := node.Get([]byte("user:1"))
		if err != nil {
			log.Printf("  ✗ %s: Failed to read: %v", name, err)
		} else {
			fmt.Printf("  ✓ %s: user:1 = %s\n", name, string(value))
		}
		node.Close()
	}
	
	fmt.Println("\nDemo completed!")
} 