#!/usr/bin/env bash
set -e

# This script waits for the RPC endpoint to become available and then starts the SQLite ETL process.
# It handles:
#   - Checking for RPC endpoint availability with retry logic
#   - Cleaning up any existing SQLite database files
#   - Starting the SQLite ETL process that transfers data from the blockchain to SQLite
#   - Configuring connection parameters for the RPC endpoint and SQLite database
#   - Setting up proper error handling

# Default values
RPC_ENDPOINT="http://localhost:8545"
WAL_PATH="/tmp/golembase.wal"
DB_PATH="/tmp/golembase-sqlite"
MAX_ATTEMPTS=30
SLEEP_SECONDS=2

# Get the script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Get the golem base directory (parent of script directory)
GOLEM_BASE_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Clean up any existing SQLite files
rm -f ${DB_PATH}*

echo "Waiting for RPC endpoint to become available at ${RPC_ENDPOINT}..."

# Function to check if RPC is available
check_rpc() {
  curl -s -o /dev/null -w "%{http_code}" "${RPC_ENDPOINT}" | grep -q "200"
  return $?
}

# Wait for RPC to become available
attempt=1
while [ $attempt -le $MAX_ATTEMPTS ]; do
  echo "Attempt $attempt of $MAX_ATTEMPTS: Checking RPC availability..."

  if check_rpc; then
    echo "RPC endpoint is available! Starting ETL process..."
    break
  fi

  echo "RPC endpoint not available yet. Waiting ${SLEEP_SECONDS} seconds..."
  sleep $SLEEP_SECONDS
  attempt=$((attempt + 1))
done

if [ $attempt -gt $MAX_ATTEMPTS ]; then
  echo "Error: RPC endpoint did not become available after $MAX_ATTEMPTS attempts."
  exit 1
fi

# Start the ETL process
echo "Starting SQLite ETL process..."
exec go run "${GOLEM_BASE_DIR}/etl/sqlite/" --wal "${WAL_PATH}" --db "${DB_PATH}" --rpc-endpoint "${RPC_ENDPOINT}"
