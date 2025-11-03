#!/usr/bin/env bash

# This script will run all Golem base development processes using Overmind
# It will start all processes defined in the Procfile: geth node, SQLite ETL,
# MongoDB, and MongoDB ETL processes

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
PROCFILE_PATH="$SCRIPT_DIR/Procfile"

# Check if Overmind is installed
if ! command -v overmind &> /dev/null; then
    echo "Error: Overmind is not installed."
    echo "Please install Overmind to run this script."
    echo "  - macOS: brew install overmind"
    echo "  - Others: https://github.com/DarthSim/overmind#installation"
    exit 1
fi

# Check if Procfile exists
if [ ! -f "$PROCFILE_PATH" ]; then
    echo "Error: Procfile not found at $PROCFILE_PATH"
    exit 1
fi

# Clean up old WAL files if they exist
if [[ -e "/tmp/golembase.wal" ]]; then
    if [[ -d "/tmp/golembase.wal" ]]; then
        rm -rf "/tmp/golembase.wal"
    elif [[ -f "/tmp/golembase.wal" ]]; then
        rm "/tmp/golembase.wal"
    fi
fi

echo "Starting all Golem base development processes with Overmind..."
echo "Using Procfile: $PROCFILE_PATH"
echo "Press Ctrl+C to stop all processes."

# Change to the golem-base directory where the Procfile is located
cd "$SCRIPT_DIR"

# Run Overmind with the Procfile in the current directory
exec overmind start
