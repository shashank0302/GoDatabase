#!/bin/bash

# Test script to check cluster health

echo "Testing Raft Cluster Health"
echo "=========================="
echo ""

# Function to test a node
test_node() {
    local port=$1
    local node=$2
    echo "Testing Node $node on port $port..."
    
    # Try to connect
    timeout 2 bash -c "echo 'help' | ./client -addr localhost:$port 2>&1" > /tmp/test_output.txt
    
    if grep -q "GeoCacheGoDB Client" /tmp/test_output.txt; then
        echo "✓ Node $node is responding"
        
        # Try a simple operation
        echo "put test_key test_value" | timeout 5 ./client -addr localhost:$port 2>&1 > /tmp/test_operation.txt
        
        if grep -q "OK" /tmp/test_operation.txt; then
            echo "✓ Node $node accepted write (is leader)"
        elif grep -q "not the leader" /tmp/test_operation.txt; then
            echo "✓ Node $node is follower"
        elif grep -q "no leader available" /tmp/test_operation.txt; then
            echo "✗ Node $node reports no leader available"
        else
            echo "✗ Node $node failed operation"
            cat /tmp/test_operation.txt
        fi
    else
        echo "✗ Node $node is not responding"
    fi
    echo ""
}

# Test all three nodes
test_node 50051 1
test_node 50052 2
test_node 50053 3

echo "Checking processes..."
ps aux | grep raft-server | grep -v grep

echo ""
echo "Checking ports..."
netstat -tlnp 2>/dev/null | grep -E "50051|50052|50053|51051|51052|51053" || \
lsof -i :50051 -i :50052 -i :50053 -i :51051 -i :51052 -i :51053 2>/dev/null
