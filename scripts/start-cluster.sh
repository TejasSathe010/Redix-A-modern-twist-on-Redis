#!/bin/bash

# This script starts multiple instances of the server in cluster mode

# Configuration
NUM_NODES=3
BASE_PORT=6379
DATA_DIR="./data"
CONFIG_DIR="./config"

# Create data and config directories if they don't exist
mkdir -p $DATA_DIR
mkdir -p $CONFIG_DIR

# Generate configuration files for each node
echo "Generating configuration files..."
for i in $(seq 0 $(($NUM_NODES - 1))); do
    port=$(($BASE_PORT + $i))
    cat > "$CONFIG_DIR/node$i.json" << EOF
{
    "server": {
        "addr": ":$port",
        "mode": "cluster",
        "cluster_nodes": [
            "localhost:$BASE_PORT",
            "localhost:$((BASE_PORT + 1))",
            "localhost:$((BASE_PORT + 2))"
        ]
    },
    "storage": {
        "dir": "$DATA_DIR/node$i"
    }
}
EOF
done

# Start each node
echo "Starting cluster nodes..."
for i in $(seq 0 $(($NUM_NODES - 1))); do
    port=$(($BASE_PORT + $i))
    echo "Starting node $i on port $port..."
    go run cmd/server/main.go --config "$CONFIG_DIR/node$i.json" &
done

echo "Cluster started with $NUM_NODES nodes."