package storage

import (
	"context"
	"strings"
	"sync"

	"github.com/go-redis/redis/v8"
)

type InMemoryStore struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]interface{}),
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
}

type CommandHandler struct {
	store *InMemoryStore
}

func NewCommandHandler(store *InMemoryStore) *CommandHandler {
	return &CommandHandler{store: store}
}

func (h *CommandHandler) HandleCommand(ctx context.Context, args []string) (interface{}, error) {
	if len(args) == 0 {
		return nil, redis.ErrWrongNumberOfArgs
	}

	command := strings.ToUpper(args[0])

	switch command {
	case "SET":
		if len(args) < 3 {
			return nil, redis.ErrWrongNumberOfArgs
		}
		h.store.Set(args[1], strings.Join(args[2:], " "))
		return "OK", nil
	case "GET":
		if len(args) != 2 {
			return nil, redis.ErrWrongNumberOfArgs
		}
		value, exists := h.store.Get(args[1])
		if !exists {
			return nil, redis.Nil
		}
		return value, nil
	case "DEL":
		if len(args) < 2 {
			return nil, redis.ErrWrongNumberOfArgs
		}
		h.store.Delete(args[1])
		return 1, nil
	default:
		return nil, redis.ErrUnknownCommand
	}
}
