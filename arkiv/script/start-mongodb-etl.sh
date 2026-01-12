#!/usr/bin/env bash
set -e

# This script starts the MongoDB ETL process for Golem Base.
# It handles:
#   - Waiting for MongoDB to become available in replica set mode
#   - Waiting for the RPC endpoint to become available
#   - Starting the MongoDB ETL process that transfers data from the blockchain to MongoDB
#   - Configuring connection parameters for MongoDB and the RPC endpoint
#   - Setting up proper error handling and retry logic


# Configuration
MONGO_URI="mongodb://admin:password@localhost:27017"
WAL_PATH="/tmp/golembase.wal"
RPC_ENDPOINT="http://localhost:8545"
DB_NAME="golembase"
MAX_ATTEMPTS=30
SLEEP_INTERVAL=5

# Get the script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Get the golem base directory (parent of script directory)
GOLEM_BASE_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Waiting for MongoDB to be available in replica mode..."

# Function to check MongoDB status
check_mongo_status() {
  echo "Attempting to connect to MongoDB at $MONGO_URI..."
  if ! mongosh "$MONGO_URI" --quiet --eval 'rs.status().ok' 2>&1; then
    echo "MongoDB connection failed"
    return 1
  fi
  echo "MongoDB connection successful"
  return 0
}

# Wait for MongoDB to become available in replica mode
attempt=0
while [ $attempt -lt $MAX_ATTEMPTS ]; do
  ((attempt++)) || true

  echo "Attempt $attempt/$MAX_ATTEMPTS: Checking MongoDB status..."
  if check_mongo_status; then
    echo "MongoDB is available and running in replica mode!"
    break
  fi

  if [ $attempt -eq $MAX_ATTEMPTS ]; then
    echo "Failed to connect to MongoDB in replica mode after $MAX_ATTEMPTS attempts."
    exit 1
  fi

  echo "MongoDB not ready yet or not in replica mode. Waiting $SLEEP_INTERVAL seconds..."
  sleep $SLEEP_INTERVAL
done

# Function to check RPC endpoint status
check_rpc_status() {
  curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"net_version","params":[],"id":1}' "$RPC_ENDPOINT" | grep -q "result" && echo 1 || echo 0
}

# Wait for RPC endpoint to become available
echo "Waiting for RPC endpoint to be available..."
attempt=0
while [ $attempt -lt $MAX_ATTEMPTS ]; do
  ((attempt++)) || true

  echo "Attempt $attempt/$MAX_ATTEMPTS: Checking RPC endpoint status..."
  STATUS=$(check_rpc_status)

  if [ "$STATUS" = "1" ]; then
    echo "RPC endpoint is available!"
    break
  fi

  if [ $attempt -eq $MAX_ATTEMPTS ]; then
    echo "Failed to connect to RPC endpoint after $MAX_ATTEMPTS attempts."
    exit 1
  fi

  echo "RPC endpoint not ready yet. Waiting $SLEEP_INTERVAL seconds..."
  sleep $SLEEP_INTERVAL
done

# Start the MongoDB ETL process
echo "Starting MongoDB ETL process..."
exec go run "${GOLEM_BASE_DIR}/etl/mongodb/" --wal "$WAL_PATH" --mongo-uri "$MONGO_URI" --rpc-endpoint "$RPC_ENDPOINT" --db-name "$DB_NAME"
