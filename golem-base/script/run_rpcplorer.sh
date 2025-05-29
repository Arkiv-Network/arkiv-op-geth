#!/usr/bin/env bash

# Default node URL
NODE_URL=${NODE_URL:-"http://localhost:8545"}

# Function to check if node is available
check_node() {
  curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' "$NODE_URL" >/dev/null
  return $?
}

# Wait for node to be available
echo "Waiting for node to be available at $NODE_URL..."
while ! check_node; do
  echo "Node not available yet. Retrying in 2 seconds..."
  sleep 2
done
echo "Node is available!"

# Start rpcplorer
echo "Starting rpcplorer..."
rpcplorer
