# Running OP-Geth Locally with Docker Compose

This guide explains how to run OP-Geth locally in [dev mode](https://geth.ethereum.org/docs/developers/dapp-developer/dev-mode) with MongoDB and SQLite ETL services using Docker Compose.

> **Tip:** We also have a [Golem Base CLI](./cmd/golembase/README.md) available for interacting with the Golem Base storage system. 

## Services Overview

### Core Services

The docker-compose file defines the following services:

- **op-geth**: The main Ethereum node service running in [dev mode](https://geth.ethereum.org/docs/developers/dapp-developer/dev-mode)
- **rpcplorer**: Web interface for exploring blockchain data

### Supporting Services

Additionally the following supporting service is included:

- **setup**: Initializes MongoDB keyfile for replica set authentication

## Prerequisites

- Docker
- Docker Compose

## Getting Started

1. Clone the repository
2. Run the services:
   ```bash
   docker-compose up
   ```

   Or optionally provide the -d option to run in detached state:
   ```bash
   docker-compose up -d
   ```

## Service Details

### OP-Geth
- Runs in dev mode with HTTP and WebSocket APIs enabled
- Exposes port 8545 for RPC connections
- Supports various APIs: eth, web3, net, debug, golembase
- Uses write-ahead logging for data persistence

### RPC Explorer
- Web interface for exploring blockchain data
- Exposes port 8080 (Note: If port 8080 is already in use on your machine, change the host port number, which is the first in the pair.)
- Connects to op-geth RPC endpoint

**Tip:** RPC Explorer is a realtime explorer that connects directly to your local node. You can open RPC Explorer right in the browser. After the services have all started up, simply go to http://localhost:8080/. Here are some features:

- **Block Explorer:** View detailed information about blocks, including transactions, gas usage, miner data, and more.
- **Transaction Explorer:** Examine transaction details, including status, gas costs, input data, and involved addresses.
- **Address Explorer:** View basic address information including balance, transaction count (nonce), and contract status.
- **Smart Search:** Instant navigation using the search bar that automatically detects and redirects to addresses, transactions, or blocks.
- **Real-time Updates:** Auto-refreshing display of latest blocks and transactions.

You can learn more [here](https://github.com/Golem-Base/rpcplorer).

## Volumes

The following volumes are created by the docker-compose file and managed by Docker Compose:
- `geth_data`: OP-Geth data storage

## Ports

- 8545: OP-Geth RPC
- 8080: RPC Explorer

## Building with Latest Changes

When pulling the latest changes from the repository, you may need to rebuild the services to ensure you have the latest code:

```bash
# Stop all services
docker-compose down

# Remove existing images (optional, but recommended when pulling latest changes)
docker-compose rm -f

# Rebuild and start services
docker-compose up --build
```

## Development

To modify or extend the setup:

1. Edit `docker-compose.yml` for service configuration
2. Modify the Dockerfile for service-specific build requirements
3. Update environment variables as needed

## Troubleshooting

1. Check service logs:
   ```bash
   docker-compose logs [service-name]
   ```

2. Verify service health:
   ```bash
   docker-compose ps
   ```

3. Restart services:
   ```bash
   docker-compose restart [service-name]
   ```
