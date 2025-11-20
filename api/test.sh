#!/bin/bash

# Test script for file upload API

echo "Creating test files..."
echo "fake video content" > test.mp4
echo "text content" > test.txt

echo "Building API..."
go build -o api .
if [ $? -ne 0 ]; then
    echo "Build failed"
    exit 1
fi

echo "Starting API server in background..."
export PORT=5002
./api > server.log 2>&1 &
SERVER_PID=$!

# Wait for server to start
echo "Waiting for server to start..."
sleep 5

echo -e "\n=== Testing GET / endpoint ==="
curl -s http://localhost:5002/ | jq .

echo -e "\n=== Testing POST /upload endpoint (Valid Video) ==="
curl -s -X POST http://localhost:5002/upload \
  -H "Content-Type: multipart/form-data" \
  -F "file=@test.mp4;type=video/mp4" | jq .

echo -e "\n=== Testing POST /upload endpoint (Invalid File) ==="
curl -s -X POST http://localhost:5002/upload \
  -F "file=@test.txt" | jq .

# Cleanup
echo -e "\n=== Cleaning up ==="
kill $SERVER_PID 2>/dev/null
rm -f test.mp4 test.txt api server.log

echo "Test completed!"
