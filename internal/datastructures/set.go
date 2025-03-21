package datastructures

import (
	"sync"
)

type Set struct {
	mu       sync.RWMutex
	elements map[interface{}]struct{}
}

func NewSet() *Set {
	return &Set{
		elements: make(map[interface{}]struct{}),
	}
}

func (s *Set) Add(value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.elements[value] = struct{}{}
}

func (s *Set) Remove(value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.elements, value)
}

func (s *Set) Contains(value interface{}) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.elements[value]
	return exists
}

func (s *Set) Members() []interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	members := make([]interface{}, 0, len(s.elements))
	for value := range s.elements {
		members = append(members, value)
	}
	return members
}

func (s *Set) Cardinality() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.elements)
}

func (s *Set) Union(other *Set) *Set {
	result := NewSet()
	s.mu.RLock()
	other.mu.RLock()
	defer s.mu.RUnlock()
	defer other.mu.RUnlock()

	for value := range s.elements {
		result.Add(value)
	}
	for value := range other.elements {
		result.Add(value)
	}

	return result
}

func (s *Set) Intersect(other *Set) *Set {
	result := NewSet()
	s.mu.RLock()
	other.mu.RLock()
	defer s.mu.RUnlock()
	defer other.mu.RUnlock()

	for value := range s.elements {
		if _, exists := other.elements[value]; exists {
			result.Add(value)
		}
	}

	return result
}

func (s *Set) Diff(other *Set) *Set {
	result := NewSet()
	s.mu.RLock()
	other.mu.RLock()
	defer s.mu.RUnlock()
	defer other.mu.RUnlock()

	for value := range s.elements {
		if _, exists := other.elements[value]; !exists {
			result.Add(value)
		}
	}

	return result
}
