package storage

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type InMemoryStore struct {
	data map[string]interface{}
	ttls map[string]time.Time
	mu   sync.RWMutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]interface{}),
		ttls: make(map[string]time.Time),
	}
}

func (s *InMemoryStore) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *InMemoryStore) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.data[key]
	return value, exists
}

func (s *InMemoryStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	delete(s.ttls, key)
}

func (s *InMemoryStore) Incr(key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if value, exists := s.data[key]; exists {
		if num, ok := value.(int64); ok {
			num++
			s.data[key] = num
			return num, nil
		}
		return 0, fmt.Errorf("ERR value is not an integer or out of range")
	}

	s.data[key] = int64(1)
	return int64(1), nil
}

func (s *InMemoryStore) IncrBy(key string, increment int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if value, exists := s.data[key]; exists {
		if num, ok := value.(int64); ok {
			num += increment
			s.data[key] = num
			return num, nil
		}
		return 0, fmt.Errorf("ERR value is not an integer or out of range")
	}

	s.data[key] = increment
	return increment, nil
}

func (s *InMemoryStore) Expire(key string, seconds int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data[key]; exists {
		s.ttls[key] = time.Now().Add(time.Duration(seconds) * time.Second)
		return true
	}
	return false
}

func (s *InMemoryStore) TTL(key string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if expireTime, exists := s.ttls[key]; exists {
		remaining := expireTime.Sub(time.Now())
		return int64(remaining.Seconds()), true
	}
	return -1, false
}

func (s *InMemoryStore) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.data[key]
	return exists
}

func (s *InMemoryStore) Keys(pattern string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var matches []string
	for key := range s.data {
		if match, _ := filepath.Match(pattern, key); match {
			matches = append(matches, key)
		}
	}
	return matches
}

func (s *InMemoryStore) Type(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if value, exists := s.data[key]; exists {
		switch value.(type) {
		case string:
			return "string"
		case int64:
			return "integer"
		default:
			return "unknown"
		}
	}
	return ""
}

func (s *InMemoryStore) FlushAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]interface{})
	s.ttls = make(map[string]time.Time)
}

func (s *InMemoryStore) MSet(keysValues ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := 0; i < len(keysValues); i += 2 {
		if i+1 < len(keysValues) {
			s.data[keysValues[i]] = keysValues[i+1]
		}
	}
}

func (s *InMemoryStore) MGet(keys ...string) []interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var values []interface{}
	for _, key := range keys {
		if value, exists := s.data[key]; exists {
			values = append(values, value)
		} else {
			values = append(values, nil)
		}
	}
	return values
}

type CommandHandler struct {
	store *InMemoryStore
}

func NewCommandHandler(store *InMemoryStore) *CommandHandler {
	return &CommandHandler{store: store}
}

func (h *CommandHandler) HandleCommand(ctx context.Context, args []string) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("unknown command")
	}

	command := strings.ToUpper(args[0])

	switch command {
	case "SET":
		if len(args) < 3 {
			return nil, fmt.Errorf("wrong number of arguments for SET")
		}
		h.store.Set(args[1], strings.Join(args[2:], " "))
		return "OK", nil

	case "GET":
		if len(args) != 2 {
			return nil, fmt.Errorf("wrong number of arguments for GET")
		}
		value, exists := h.store.Get(args[1])
		if !exists {
			return nil, nil // Return nil to indicate key doesn't exist (no error)
		}
		return value, nil

	case "DEL":
		if len(args) < 2 {
			return nil, fmt.Errorf("wrong number of arguments for DEL")
		}
		key := args[1]
		if _, exists := h.store.Get(key); exists {
			h.store.Delete(key)
			return int64(1), nil
		}
		return int64(0), nil

	case "INCR":
		if len(args) != 2 {
			return nil, fmt.Errorf("wrong number of arguments for INCR")
		}
		result, err := h.store.Incr(args[1])
		if err != nil {
			return nil, err
		}
		return result, nil

	case "INCRBY":
		if len(args) != 3 {
			return nil, fmt.Errorf("wrong number of arguments for INCRBY")
		}
		incr, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("value is not an integer or out of range")
		}
		result, err := h.store.IncrBy(args[1], incr)
		if err != nil {
			return nil, err
		}
		return result, nil

	case "EXPIRE":
		if len(args) != 3 {
			return nil, fmt.Errorf("wrong number of arguments for EXPIRE")
		}
		seconds, err := strconv.Atoi(args[2])
		if err != nil {
			return nil, fmt.Errorf("invalid expire time")
		}
		result := h.store.Expire(args[1], seconds)
		if result {
			return int64(1), nil
		}
		return int64(0), nil

	case "TTL":
		if len(args) != 2 {
			return nil, fmt.Errorf("wrong number of arguments for TTL")
		}
		ttl, exists := h.store.TTL(args[1])
		if !exists {
			return nil, nil
		}
		return ttl, nil

	case "EXISTS":
		if len(args) != 2 {
			return nil, fmt.Errorf("wrong number of arguments for EXISTS")
		}
		exists := h.store.Exists(args[1])
		if exists {
			return int64(1), nil
		}
		return int64(0), nil

	case "KEYS":
		if len(args) != 2 {
			return nil, fmt.Errorf("wrong number of arguments for KEYS")
		}
		return h.store.Keys(args[1]), nil

	case "TYPE":
		if len(args) != 2 {
			return nil, fmt.Errorf("wrong number of arguments for TYPE")
		}
		return h.store.Type(args[1]), nil

	case "FLUSHALL":
		if len(args) != 1 {
			return nil, fmt.Errorf("wrong number of arguments for FLUSHALL")
		}
		h.store.FlushAll()
		return "OK", nil

	case "MSET":
		if len(args) < 3 || (len(args)-1)%2 != 0 {
			return nil, fmt.Errorf("wrong number of arguments for MSET")
		}
		h.store.MSet(args[1:]...)
		return "OK", nil

	case "MGET":
		if len(args) < 2 {
			return nil, fmt.Errorf("wrong number of arguments for MGET")
		}
		return h.store.MGet(args[1:]...), nil

	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}
