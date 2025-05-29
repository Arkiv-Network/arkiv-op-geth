#!/usr/bin/env bash


# This script sets up and runs a MongoDB instance in a Docker container.
# It configures MongoDB with authentication and replica set capabilities.
# The script handles:
#   - Creating necessary keyfiles for authentication
#   - Removing any existing MongoDB containers
#   - Starting a new MongoDB container with proper configuration
#   - Setting up proper signal handling for clean container removal on exit
#   - Exposing MongoDB on port 27017 with admin/password credentials



# Don't exit immediately on error, we want to handle errors gracefully
set -u -o pipefail

# Detect if docker is actually podman
isPodman=false
if docker version | grep -q Podman; then
  isPodman=true
fi

# Global flag to track if we're in cleanup
CLEANING_UP=false

# Function to clean up MongoDB container
cleanup() {
    # Prevent multiple cleanups
    if [ "$CLEANING_UP" = true ]; then
        return
    fi
    CLEANING_UP=true

    echo -e "\nCleaning up MongoDB container..."
    docker rm -f mongodb 2>/dev/null || true
    echo "MongoDB container removed."
    exit 0
}

# Trap SIGINT (Ctrl+C) and SIGTERM signals, but handle EXIT separately
trap cleanup SIGINT SIGTERM
trap 'if [ "$CLEANING_UP" = false ]; then cleanup; fi' EXIT

# Ensure keyfile directory exists
mkdir -p /tmp/mongodb-keyfile

# Create keyfile if it doesn't exist
if [ ! -f /tmp/mongodb-keyfile/mongodb-keyfile ]; then
    openssl rand -base64 756 > /tmp/mongodb-keyfile/mongodb-keyfile
    chmod 400 /tmp/mongodb-keyfile/mongodb-keyfile
fi

# Remove any existing MongoDB container
echo "Removing any existing MongoDB container..."
docker rm -f mongodb 2>/dev/null || true

# Run MongoDB container in foreground
echo "Starting MongoDB container..."
if [ "${isPodman}" == "true" ]; then
    # With the default networking mode (pasta) we lose connectivity to mongo
    # shortly after startup. There might be some other daemon (systemd-networkd?)
    # that doesn't like the extra network config.
    # Using slirp4netns avoids this issue.
    # Besides that, we add the U option to the bind mount to automatically chown
    # all the contents to the containers user, and we set the user to be mongodb.
    docker run --name mongodb \
        -p 27017:27017 \
        -v /tmp/mongodb-keyfile:/keyfile:U,ro \
        -e MONGO_INITDB_ROOT_USERNAME=admin \
        -e MONGO_INITDB_ROOT_PASSWORD=password \
        --user=mongodb:mongodb \
        --network=slirp4netns \
        -d docker.io/library/mongo:latest \
        mongod --replSet rs0 --bind_ip_all --keyFile /keyfile/mongodb-keyfile --auth
else
    docker run --name mongodb \
        -p 27017:27017 \
        -v /tmp/mongodb-keyfile:/keyfile:ro \
        -e MONGO_INITDB_ROOT_USERNAME=admin \
        -e MONGO_INITDB_ROOT_PASSWORD=password \
        -d docker.io/library/mongo:latest \
        mongod --replSet rs0 --bind_ip_all --keyFile /keyfile/mongodb-keyfile --auth
fi

# Check if container started successfully
if ! docker ps | grep -q mongodb; then
    echo "Failed to start MongoDB container"
    exit 1
fi

echo "MongoDB container started. Container ID: $(docker ps -q -f name=mongodb)"
echo "Waiting for MongoDB service to start..."

# Wait for MongoDB to become available with a longer timeout
COUNTER=0
MAX_WAIT=60 # Wait up to 60 seconds
while [ $COUNTER -lt $MAX_WAIT ]; do
    if docker exec mongodb mongosh --quiet --eval "db.adminCommand('ping')" 2>/dev/null; then
        echo "MongoDB is ready!"
        break
    fi

    (( COUNTER++ )) || true
    echo "Waiting... ($COUNTER/$MAX_WAIT seconds)"
    sleep 1

    # Check if container is still running
    if ! docker ps | grep -q mongodb; then
        echo "MongoDB container stopped unexpectedly. Checking logs:"
        docker logs mongodb
        exit 1
    fi
done

if [ $COUNTER -ge $MAX_WAIT ]; then
    echo "MongoDB failed to start within $MAX_WAIT seconds. Checking logs:"
    docker logs mongodb
    exit 1
fi

echo "MongoDB started, checking replica set status..."
sleep 2

# Check if replica set is already initialized
RS_STATUS=$(docker exec mongodb mongosh -u admin -p password --authenticationDatabase admin --quiet --eval "try { rs.status(); } catch(e) { e.codeName }" 2>/dev/null)

if [[ "$RS_STATUS" == *"NotYetInitialized"* ]]; then
    echo "Replica set not initialized. Initializing now..."
    # Initialize replica set
    docker exec mongodb mongosh -u admin -p password --authenticationDatabase admin --eval "rs.initiate({_id: 'rs0', members: [{_id: 0, host: 'localhost:27017'}]})"

    echo "Waiting for replica set initialization..."
    sleep 5
else
    echo "Replica set already initialized."
fi

# Verify replica set status
echo "Checking replica set status..."
docker exec mongodb mongosh -u admin -p password --authenticationDatabase admin --eval "rs.status()"

echo "MongoDB replica set is running successfully"
echo "Connect using: mongodb://admin:password@localhost:27017/?authSource=admin&replicaSet=rs0"
echo ""
echo "MongoDB is running. Press Ctrl+C to stop and remove the container."

# Watch container logs until terminated
echo "Showing MongoDB logs (Ctrl+C to stop):"
docker logs -f mongodb || true

# If we get here, we've exited the logs command but container might still be running
# This prevents premature cleanup and ensures we wait for user to press Ctrl+C
if docker ps | grep -q mongodb; then
    echo "MongoDB is still running. Press Ctrl+C to stop and remove the container."
    # Wait indefinitely until the container stops or we're interrupted
    while docker ps | grep -q mongodb; do
        sleep 1
    done
fi
