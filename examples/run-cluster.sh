#!/bin/bash

# Example: Running a distributed GeoCacheGoDB cluster

echo "Starting GeoCacheGoDB cluster..."

# Start node 1 (primary)
echo "Starting node 1 on port 8080..."
./server -addr :8080 -data ./data1 -storage badger &
NODE1_PID=$!

# Start node 2 (backup)
echo "Starting node 2 on port 8081..."
./server -addr :8081 -data ./data2 -storage badger &
NODE2_PID=$!

# Start node 3 (backup)
echo "Starting node 3 on port 8082..."
./server -addr :8082 -data ./data3 -storage badger &
NODE3_PID=$!

echo "Cluster started with 3 nodes:"
echo "  - Node 1: localhost:8080 (PID: $NODE1_PID)"
echo "  - Node 2: localhost:8081 (PID: $NODE2_PID)"
echo "  - Node 3: localhost:8082 (PID: $NODE3_PID)"
echo ""
echo "You can connect to any node using:"
echo "  ./client -addr localhost:8080"
echo ""
echo "Press Ctrl+C to stop all nodes"

# Wait for interrupt
trap "echo 'Stopping cluster...'; kill $NODE1_PID $NODE2_PID $NODE3_PID; exit" INT

# Keep script running
wait 