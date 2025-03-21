package datastructures

import (
	"sync"
)

type Hash struct {
	mu     sync.RWMutex
	fields map[string]interface{}
}

func NewHash() *Hash {
	return &Hash{
		fields: make(map[string]interface{}),
	}
}

func (h *Hash) HSet(field string, value interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.fields[field] = value
}

func (h *Hash) HGet(field string) (interface{}, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	value, exists := h.fields[field]
	return value, exists
}

func (h *Hash) HDel(field string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.fields, field)
}

func (h *Hash) HExists(field string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, exists := h.fields[field]
	return exists
}

func (h *Hash) HGetAll() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make(map[string]interface{})
	for k, v := range h.fields {
		result[k] = v
	}
	return result
}

func (h *Hash) HKeys() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	keys := make([]string, 0, len(h.fields))
	for k := range h.fields {
		keys = append(keys, k)
	}
	return keys
}

func (h *Hash) HVals() []interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()
	values := make([]interface{}, 0, len(h.fields))
	for _, v := range h.fields {
		values = append(values, v)
	}
	return values
}

func (h *Hash) HLen() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.fields)
}
