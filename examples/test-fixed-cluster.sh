#!/bin/bash

# Test the fixed cluster implementation

echo "Testing Fixed Raft Cluster"
echo "========================="

# Kill any existing processes
echo "Killing existing processes..."
pkill -f raft-server

# Clean up
echo "Cleaning up..."
rm -rf ./raft-data1 ./raft-data2 ./raft-data3
rm -f node*.log cluster.pids

# Build
echo "Building..."
go build -o raft-server ./cmd/raft-server
go build -o client ./cmd/client

echo ""
echo "Starting cluster with fixes..."
echo "============================="

# Start nodes with a small delay between them
./raft-server -id node1 -addr :50051 -peers "node2:localhost:51052,node3:localhost:51053" -data ./raft-data1 -storage badger > node1.log 2>&1 &
NODE1_PID=$!
echo "Started Node 1 (PID: $NODE1_PID)"

sleep 1

./raft-server -id node2 -addr :50052 -peers "node1:localhost:51051,node3:localhost:51053" -data ./raft-data2 -storage badger > node2.log 2>&1 &
NODE2_PID=$!
echo "Started Node 2 (PID: $NODE2_PID)"

sleep 1

./raft-server -id node3 -addr :50053 -peers "node1:localhost:51051,node2:localhost:51052" -data ./raft-data3 -storage badger > node3.log 2>&1 &
NODE3_PID=$!
echo "Started Node 3 (PID: $NODE3_PID)"

echo ""
echo "Waiting for cluster to stabilize..."
sleep 10

echo ""
echo "Checking cluster status..."
echo "========================"

# Check for leaders
echo "Leaders found:"
grep -h "became leader" node*.log | sort

echo ""
echo "Current cluster state:"
tail -n 5 node1.log | grep -E "(Cluster info|became leader|stepping down)"
tail -n 5 node2.log | grep -E "(Cluster info|became leader|stepping down)"
tail -n 5 node3.log | grep -E "(Cluster info|became leader|stepping down)"

echo ""
echo "Testing client operations..."
echo "==========================="

# Test each node
for port in 50051 50052 50053; do
    echo ""
    echo "Testing node on port $port..."
    
    # Try to connect and do a simple operation
    timeout 5 bash -c "echo -e 'put test_key test_value\nget test_key\nquit' | ./client -addr localhost:$port" 2>&1 | head -n 10
    
    if [ $? -eq 0 ]; then
        echo "✓ Node $port is working"
    else
        echo "✗ Node $port failed"
    fi
done

echo ""
echo "Cluster PIDs: $NODE1_PID $NODE2_PID $NODE3_PID"
echo "To stop: kill $NODE1_PID $NODE2_PID $NODE3_PID"
echo "To monitor: tail -f node*.log"
