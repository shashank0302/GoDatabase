package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"godatabase/internal/raft"
	"godatabase/internal/rpc"
	"godatabase/internal/storage"
)

func main() {
	// Parse command line flags
	addr := flag.String("addr", ":50051", "The server address")
	nodeID := flag.String("id", "node1", "The node ID")
	peers := flag.String("peers", "", "Comma-separated list of peer addresses (id:addr)")
	storageType := flag.String("storage", "badger", "Storage type (badger or btree)")
	dataDir := flag.String("data", "data", "Data directory")
	flag.Parse()

	// Parse peers
	peerMap := make(map[string]string)
	if *peers != "" {
		peerList := splitPeers(*peers)
		for _, peer := range peerList {
			parts := splitPeer(peer)
			if len(parts) == 2 {
				peerMap[parts[0]] = parts[1]
			}
		}
	}

	// Create storage
	var store storage.Storage
	var err error

	switch *storageType {
	case "badger":
		store, err = storage.NewBadgerStorage(*dataDir)
	case "btree":
		store, err = storage.NewStorage(storage.CustomStorage, *dataDir)
	default:
		log.Fatalf("Unknown storage type: %s", *storageType)
	}

	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	// Get global cluster
	globalCluster := raft.GetGlobalCluster()

	// Calculate Raft RPC port (gRPC port + 1000 to avoid conflicts)
	port := parsePort(*addr) + 1000
	raftRPCAddr := ":" + strconv.Itoa(port)
	log.Printf("gRPC address: %s, Raft RPC address: %s", *addr, raftRPCAddr)

	// Create Raft node
	node := raft.NewRaftNode(*nodeID, raftRPCAddr, peerMap, store)

	// Register with global cluster
	err = globalCluster.RegisterNode(node)
	if err != nil {
		log.Fatalf("Failed to register node with global cluster: %v", err)
	}

	// Start the node
	if err := node.Start(); err != nil {
		globalCluster.UnregisterNode(*nodeID)
		log.Fatalf("Failed to start node: %v", err)
	}

	// Start RPC server
	if err := node.StartRPCServer(); err != nil {
		log.Printf("Failed to start RPC server: %v", err)
		// Don't stop the node, let it continue without RPC for now
	}

	// Create Raft storage wrapper
	raftStorage := raft.NewRaftStorage(globalCluster, *nodeID)

	// Create and start gRPC server
	server := rpc.NewServer(raftStorage)
	go func() {
		if err := server.Start(*addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Raft server started:")
	log.Printf("  Node ID: %s", *nodeID)
	log.Printf("  Address: %s", *addr)
	log.Printf("  Peers: %v", peerMap)
	log.Printf("  Storage: %s", *storageType)

	// Start heartbeat monitor
	globalCluster.StartHeartbeatMonitor()

	// Print cluster info periodically
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			info := globalCluster.GetClusterInfo()
			log.Printf("Cluster info: %+v", info)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	log.Println("Shutting down server...")
	server.Stop()
	globalCluster.UnregisterNode(*nodeID)
}

// splitPeers splits a comma-separated list of peers
func splitPeers(peers string) []string {
	if peers == "" {
		return []string{}
	}

	var result []string
	start := 0
	for i, char := range peers {
		if char == ',' {
			if i > start {
				result = append(result, peers[start:i])
			}
			start = i + 1
		}
	}
	if start < len(peers) {
		result = append(result, peers[start:])
	}
	return result
}

// splitPeer splits a peer string into ID and address
func splitPeer(peer string) []string {
	for i, char := range peer {
		if char == ':' {
			return []string{peer[:i], peer[i+1:]}
		}
	}
	return []string{peer}
}

// parsePort extracts the port number from an address string
func parsePort(addr string) int {
	parts := strings.Split(addr, ":")
	if len(parts) < 2 {
		return 50051 // default port
	}
	port, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return 50051 // default port
	}
	return port
}
