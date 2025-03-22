# Build and test automation

# Default target
all: build

# Build the server binary
build:
	go build -o bin/redis-like-server cmd/server/main.go

# Run the server
run:
	go run cmd/server/main.go --config config/default.json

# Run tests
test:
	go test ./...

# Generate documentation
docs:
	@echo "Generating documentation..."
	@# You would typically use a documentation generator here
	@echo "Documentation generated."

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf ./data/
	rm -rf ./config/

# Benchmark the server
benchmark:
	@echo "Running benchmark..."
	@bash scripts/benchmark.sh

# Start a cluster
cluster:
	@echo "Starting cluster..."
	@bash scripts/start-cluster.sh

# Help message
help:
	@echo "Available targets:"
	@echo "  all       - Build the project"
	@echo "  build     - Build the server binary"
	@echo "  run       - Run the server"
	@echo "  test      - Run all tests"
	@echo "  docs      - Generate documentation"
	@echo "  clean     - Clean build artifacts"
	@echo "  benchmark - Run benchmark"
	@echo "  cluster   - Start a cluster"
	@echo "  help      - Show this help message"