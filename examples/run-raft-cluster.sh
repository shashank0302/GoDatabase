#!/bin/bash

# Example: Running a Raft GeoCacheGoDB cluster

echo "Starting Raft GeoCacheGoDB cluster..."

# Clean up any existing data
echo "Cleaning up old data..."
rm -rf ./raft-data1 ./raft-data2 ./raft-data3

# Build the raft server
echo "Building raft server..."
go build -o raft-server ./cmd/raft-server

# Start node 1 (leader candidate)
echo "Starting node 1 on port 50051..."
./raft-server -id node1 -addr :50051 -peers "node2:localhost:51052,node3:localhost:51053" -data ./raft-data1 -storage badger &
NODE1_PID=$!

# Start node 2
echo "Starting node 2 on port 50052..."
./raft-server -id node2 -addr :50052 -peers "node1:localhost:51051,node3:localhost:51053" -data ./raft-data2 -storage badger &
NODE2_PID=$!

# Start node 3
echo "Starting node 3 on port 50053..."
./raft-server -id node3 -addr :50053 -peers "node1:localhost:51051,node2:localhost:51052" -data ./raft-data3 -storage badger &
NODE3_PID=$!

echo "Raft cluster started with 3 nodes:"
echo "  - Node 1: localhost:50051 (PID: $NODE1_PID)"
echo "  - Node 2: localhost:50052 (PID: $NODE2_PID)"
echo "  - Node 3: localhost:50053 (PID: $NODE3_PID)"
echo ""
echo "You can connect to any node using:"
echo "  ./client -addr localhost:50051"
echo ""
echo "The cluster will automatically elect a leader and handle failures."
echo "Press Ctrl+C to stop all nodes"

# Wait for interrupt
trap "echo 'Stopping cluster...'; kill $NODE1_PID $NODE2_PID $NODE3_PID; exit" INT

# Keep script running
wait
