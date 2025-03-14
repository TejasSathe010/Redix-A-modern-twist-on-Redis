# API Documentation

## Supported Commands

### Basic Key-Value Operations

- SET key value
- GET key
- DEL key

### String Operations

- APPEND key value
- INCR key
- INCRBY key increment
- DECR key
- DECRBY key decrement

### List Operations

- LPUSH key value
- RPUSH key value
- LPOP key
- RPOP key
- LRANGE key start end

### Set Operations

- SADD key member
- SREM key member
- SISMEMBER key member
- SMEMBERS key
- SINTER key [key ...]

### Sorted Set Operations

- ZADD key score member
- ZREM key member
- ZRANK key member
- ZREVRANK key member
- ZRANGE key start end
- ZREVRANGE key start end

### Hash Operations

- HSET key field value
- HGET key field
- HDEL key field
- HEXISTS key field
- HGETALL key

## Command Format

All commands follow the Redis protocol format:
- Commands are sent as space-separated strings
- Each command ends with \r\n
- Bulk strings are prefixed with $ and their length
- Arrays are prefixed with * and their length

Example: