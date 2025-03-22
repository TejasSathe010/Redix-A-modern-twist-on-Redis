# Redix A modern twist on Redis in Go (Golang)

A high-performance, Redis-compatible in-memory database implemented in Go (Golang) with support for advanced data structures, persistence, replication, and clustering.

## Table of Contents
1. [Features](#features)
2. [Getting Started](#getting-started)
3. [Usage Examples](#usage-examples)
4. [Project Structure](#project-structure)
5. [API Documentation](#api-documentation)
6. [System Architecture](#system-architecture)
7. [Design Decisions](#design-decisions)
8. [Contributing](#contributing)
9. [License](#license)

## Features

- **In-Memory Storage**: Blazing-fast data access with optional persistence
- **Redis Protocol Support**: Fully compatible with Redis clients
- **Advanced Data Structures**:
  - Strings
  - Lists
  - Sets
  - Sorted Sets
  - Hashes
- **Replication**:
  - Master-slave replication
  - Semi-synchronous replication option
  - Automatic failover
- **Clustering**:
  - Hash slot-based sharding
  - Dynamic rebalancing
  - Quorum-based failure detection
- **Security**:
  - TLS encryption
  - Role-based access control
  - Data-at-rest encryption
- **Monitoring**: Comprehensive metrics collection

## Getting Started

### Prerequisites
- Go 1.20+
- Git

### Installation

```bash
# Clone repository
git clone [https://github.com/yourusername/redis-like-db.git](https://github.com/TejasSathe010/Redix-A-modern-twist-on-Redis)
cd redis-like-db

# Build the server
make build
```

### Running the Server

```bash
# Basic run
make run

# With configuration file
go run cmd/server/main.go --config config.json
```

### Starting a Cluster

```bash
make cluster
```

### Benchmarking

```bash
make benchmark
```

## Usage Examples

### Basic Operations

```bash


# Using redis-cli or telnet localhost 6388
redis-cli SET mykey "Hello, World!"
redis-cli GET mykey
```

## Project Structure

```
redis-like-db/
├── cmd/
│   └── server/          # Server entry point
│       └── main.go
├── internal/
│   ├── storage/         # Storage implementations
│   ├── datastructures/  # Data structure implementations
│   ├── network/         # Network layer
│   ├── replication/     # Replication system
│   ├── cluster/         # Cluster coordination
│   └── security/        # Security components
├── pkg/
│   ├── redisclient/     # Client implementation
│   └── utils/           # Utility packages
├── scripts/             # Helper scripts
├── docs/                # Documentation
└── Makefile             # Build automation
```

## API Documentation

### Supported Commands

- **Key-Value Operations**: SET, GET, DEL - Already added
- **String Operations**: APPEND, INCR, DECR - Already added
- **List Operations**: LPUSH, RPUSH, LPOP, RPOP - Soon
- **Set Operations**: SADD, SREM, SISMEMBER - Soon
- **Sorted Set Operations**: ZADD, ZREM, ZRANK
- **Hash Operations**: HSET, HGET, HDEL

### Command Format

Follows standard Redis protocol (RESP):
```
*3\r\n$3\r\nSET\r\n$4\r\nkey\r\n$5\r\nvalue\r\n
```

### Response Format

- Simple strings: `+OK\r\n`
- Errors: `-ERR Unknown command\r\n`
- Bulk strings: `$5\r\nHello\r\n`
- Arrays: `*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n`
- Integers: `:1000\r\n`

## System Architecture

### High-Level Components

1. **Network Layer**: Handles client connections and protocol parsing
2. **Command Processor**: Validates and routes commands
3. **Data Storage**: Manages in-memory and persistent storage
4. **Replication System**: Implements master-slave replication
5. **Cluster Coordination**: Manages sharding and distributed operations
6. **Security Layer**: Handles authentication and encryption

### Data Flow

1. Client connects to network layer
2. Protocol parsing and command validation
3. Command execution against data storage
4. Response generation and transmission

### Persistence Strategy

- Write-Ahead Logging (WAL)
- Periodic snapshots
- Hybrid persistence model

## Design Decisions

### Storage Engine

- **LSM Tree**: Chosen for better write performance and efficient disk usage
- **Memtable**: In-memory structure for recent writes
- **Compaction**: Regular merging of SSTables to optimize storage

### Network Layer

- **Goroutine-per-connection**: Efficient connection handling
- **Non-blocking I/O**: Scales to handle thousands of connections
- **Redis Protocol (RESP)**: Full compatibility with Redis clients

### Replication

- **Asynchronous by default**: Balances performance and durability
- **Semi-synchronous option**: For critical data scenarios
- **Raft consensus**: For automatic failover

### Security

- **Role-based access**: Fine-grained permission control
- **TLS 1.3**: For secure data transmission
- **AES-256**: For data-at-rest encryption

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/new-command`
3. Commit your changes: `git commit -m "Add new command"`
4. Push to your fork: `git push origin feature/new-command`
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
