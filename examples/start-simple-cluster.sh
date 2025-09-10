#!/bin/bash

# Simple script to start a 3-node Raft cluster
# This ensures all nodes start quickly and can find each other

echo "Starting Simple Raft Cluster"
echo "==========================="

# Kill any existing processes
echo "Killing existing processes..."
pkill -f raft-server

# Clean up old data
echo "Cleaning up old data..."
rm -rf ./raft-data1 ./raft-data2 ./raft-data3

# Build if needed
if [ ! -f "./raft-server" ]; then
    echo "Building raft-server..."
    go build -o raft-server ./cmd/raft-server
fi

if [ ! -f "./client" ]; then
    echo "Building client..."
    go build -o client ./cmd/client
fi

echo ""
echo "Starting nodes..."
echo "================="

# Start all nodes in quick succession
./raft-server -id node1 -addr :50051 -peers "node2:localhost:51052,node3:localhost:51053" -data ./raft-data1 -storage badger > node1.log 2>&1 &
NODE1_PID=$!
echo "Started Node 1 (PID: $NODE1_PID)"

./raft-server -id node2 -addr :50052 -peers "node1:localhost:51051,node3:localhost:51053" -data ./raft-data2 -storage badger > node2.log 2>&1 &
NODE2_PID=$!
echo "Started Node 2 (PID: $NODE2_PID)"

./raft-server -id node3 -addr :50053 -peers "node1:localhost:51051,node2:localhost:51052" -data ./raft-data3 -storage badger > node3.log 2>&1 &
NODE3_PID=$!
echo "Started Node 3 (PID: $NODE3_PID)"

echo ""
echo "Waiting for cluster to stabilize..."
sleep 5

echo ""
echo "Checking cluster status..."
echo "========================"

# Check logs for leader election
echo ""
echo "Leader election status:"
grep -h "became leader" node*.log || echo "No leader elected yet"

echo ""
echo "Current node states:"
tail -n 20 node1.log | grep -E "(starting election|became leader|granted vote)" | tail -n 5
tail -n 20 node2.log | grep -E "(starting election|became leader|granted vote)" | tail -n 5
tail -n 20 node3.log | grep -E "(starting election|became leader|granted vote)" | tail -n 5

echo ""
echo "Cluster is running!"
echo "=================="
echo "Node 1: localhost:50051 (log: node1.log)"
echo "Node 2: localhost:50052 (log: node2.log)"
echo "Node 3: localhost:50053 (log: node3.log)"
echo ""
echo "To test: ./client -addr localhost:50051"
echo "To stop: kill $NODE1_PID $NODE2_PID $NODE3_PID"
echo ""
echo "Logs are in: node1.log, node2.log, node3.log"

# Save PIDs to file for easy cleanup
echo "$NODE1_PID $NODE2_PID $NODE3_PID" > cluster.pids

echo ""
echo "To monitor logs: tail -f node*.log"
