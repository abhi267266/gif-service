#!/bin/bash

# Test script for file upload API

echo "Creating test file..."
echo "Hello from Cloudflare R2!" > test.txt

echo "Starting API server in background..."
go run . &
SERVER_PID=$!

# Wait for server to start
echo "Waiting for server to start..."
sleep 3

echo -e "\n=== Testing GET / endpoint ==="
curl -s http://localhost:8080/ | jq .

echo -e "\n=== Testing POST /upload endpoint ==="
curl -s -X POST http://localhost:8080/upload \
  -F "file=@test.txt" | jq .

# Cleanup
echo -e "\n=== Cleaning up ==="
kill $SERVER_PID 2>/dev/null
rm -f test.txt

echo "Test completed!"
