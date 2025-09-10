package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"godatabase/pkg/client"
)

func main() {
	// Parse command line flags
	addr := flag.String("addr", "localhost:50051", "The server address")
	flag.Parse()

	// Create client
	c, err := client.NewClient(*addr)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	// Create scanner for reading input
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("GeoCacheGoDB Client (type 'help' for commands)")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "help":
			fmt.Println("Commands:")
			fmt.Println("  put <key> <value>    Store a key-value pair")
			fmt.Println("  get <key>            Retrieve a value")
			fmt.Println("  delete <key>         Remove a key-value pair")
			fmt.Println("  quit                 Exit the client")

		case "put":
			if len(parts) != 3 {
				fmt.Println("Usage: put <key> <value>")
				continue
			}
			err := c.Put([]byte(parts[1]), []byte(parts[2]))
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("OK")
			}

		case "get":
			if len(parts) != 2 {
				fmt.Println("Usage: get <key>")
				continue
			}
			value, err := c.Get([]byte(parts[1]))
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("%s\n", value)
			}

		case "delete":
			if len(parts) != 2 {
				fmt.Println("Usage: delete <key>")
				continue
			}
			err := c.Delete([]byte(parts[1]))
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("OK")
			}

		case "quit":
			return

		default:
			fmt.Println("Unknown command. Type 'help' for available commands.")
		}
	}
} 