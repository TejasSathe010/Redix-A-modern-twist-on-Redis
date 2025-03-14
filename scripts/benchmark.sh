#!/bin/bash

# Configuration
SERVER_ADDR="localhost:6379"
NUM_KEYS=10000
NUM_ITERATIONS=10

# Generate test data
echo "Generating test data..."
for i in $(seq 1 $NUM_KEYS); do
    echo "SET key$i value$i" >> /tmp/benchmark_commands.txt
done

# Benchmark
echo "Starting benchmark..."
time {
    for i in $(seq 1 $NUM_ITERATIONS); do
        echo "Iteration $i..."
        cat /tmp/benchmark_commands.txt | nc $SERVER_ADDR
    done
}

# Clean up
rm /tmp/benchmark_commands.txt

echo "Benchmark completed."