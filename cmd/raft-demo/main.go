package main

import (
	"fmt"
	"log"
	"time"

	"godatabase/pkg/client"
)

func main() {
	fmt.Println("GeoCacheGoDB Raft Demo")
	fmt.Println("======================")

	// Try to connect to different nodes to find the leader
	nodes := []string{
		"localhost:8081",
		"localhost:8082",
		"localhost:8083",
	}

	var c *client.Client
	var err error

	// Find a working node
	for _, addr := range nodes {
		fmt.Printf("Trying to connect to %s...\n", addr)
		c, err = client.New(addr)
		if err == nil {
			fmt.Printf("✓ Connected to %s\n", addr)
			break
		}
		fmt.Printf("✗ Failed to connect to %s: %v\n", addr, err)
	}

	if c == nil {
		log.Fatalf("Failed to connect to any node")
	}
	defer c.Close()

	fmt.Println("\n1. Writing data to Raft cluster...")

	// Write some data
	data := map[string]string{
		"user:1":    "Alice",
		"user:2":    "Bob",
		"user:3":    "Charlie",
		"config:db": "production",
		"cache:key": "cached_value",
	}

	for key, value := range data {
		if err := c.Put([]byte(key), []byte(value)); err != nil {
			log.Printf("Failed to put %s: %v", key, err)
		} else {
			fmt.Printf("  ✓ Stored: %s = %s\n", key, value)
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("\n2. Reading data from Raft cluster...")

	// Read data back
	for key := range data {
		value, err := c.Get([]byte(key))
		if err != nil {
			log.Printf("Failed to get %s: %v", key, err)
		} else {
			fmt.Printf("  ✓ Retrieved: %s = %s\n", key, string(value))
		}
	}

	fmt.Println("\n3. Testing consistency across nodes...")

	// Test reading from different nodes
	for i, addr := range nodes {
		fmt.Printf("\nTesting node %d (%s):\n", i+1, addr)

		nodeClient, err := client.New(addr)
		if err != nil {
			fmt.Printf("  ✗ Failed to connect: %v\n", err)
			continue
		}
		defer nodeClient.Close()

		// Read a few keys
		testKeys := []string{"user:1", "config:db", "cache:key"}
		for _, key := range testKeys {
			value, err := nodeClient.Get([]byte(key))
			if err != nil {
				fmt.Printf("  ✗ Failed to read %s: %v\n", key, err)
			} else {
				fmt.Printf("  ✓ %s = %s\n", key, string(value))
			}
		}
	}

	fmt.Println("\n4. Testing leader failover...")
	fmt.Println("   (In a real scenario, you would kill the leader process)")
	fmt.Println("   The cluster should automatically elect a new leader")

	fmt.Println("\n5. Demonstrating strong consistency...")

	// Update a value
	newValue := "Alice (Updated via Raft)"
	if err := c.Put([]byte("user:1"), []byte(newValue)); err != nil {
		log.Printf("Failed to update: %v", err)
	} else {
		fmt.Printf("  ✓ Updated: user:1 = %s\n", newValue)
	}

	// Wait a bit for replication
	time.Sleep(500 * time.Millisecond)

	// Verify the update is consistent across all nodes
	fmt.Println("\n6. Verifying consistency after update...")
	for i, addr := range nodes {
		fmt.Printf("\nNode %d (%s):\n", i+1, addr)

		nodeClient, err := client.New(addr)
		if err != nil {
			fmt.Printf("  ✗ Failed to connect: %v\n", err)
			continue
		}
		defer nodeClient.Close()

		value, err := nodeClient.Get([]byte("user:1"))
		if err != nil {
			fmt.Printf("  ✗ Failed to read: %v\n", err)
		} else {
			fmt.Printf("  ✓ user:1 = %s\n", string(value))
		}
	}

	fmt.Println("\nRaft demo completed!")
	fmt.Println("\nKey benefits of Raft consensus:")
	fmt.Println("  ✓ Strong consistency - all nodes have the same data")
	fmt.Println("  ✓ Automatic leader election - no single point of failure")
	fmt.Println("  ✓ Split-brain protection - only one leader at a time")
	fmt.Println("  ✓ Fault tolerance - cluster continues with majority")
}
