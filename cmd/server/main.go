package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	
	"godatabase/internal/rpc"
	"godatabase/internal/storage"
)

func main() {
	// Parse command line flags
	addr := flag.String("addr", ":50051", "The server address")
	storageType := flag.String("storage", "badger", "Storage type (badger or btree)")
	flag.Parse()
	
	// Create storage
	var store storage.Storage
	var err error
	
	switch *storageType {
	case "badger":
		store, err = storage.NewBadgerStorage("data")
	case "btree":
		store, err = storage.NewStorage(storage.CustomStorage, "data")
	default:
		log.Fatalf("Unknown storage type: %s", *storageType)
	}
	
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()
	
	// Create and start gRPC server
	server := rpc.NewServer(store)
	go func() {
		if err := server.Start(*addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	// Graceful shutdown
	log.Println("Shutting down server...")
	server.Stop()
} 