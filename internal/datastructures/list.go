package datastructures

import (
	"sync"
)

type List struct {
	mu     sync.RWMutex
	elements []interface{}
}

func NewList() *List {
	return &List{
		elements: make([]interface{}, 0),
	}
}

func (l *List) PushFront(value interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.elements = append([]interface{}{value}, l.elements...)
}

func (l *List) PushBack(value interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.elements = append(l.elements, value)
}

func (l *List) PopFront() (interface{}, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.elements) == 0 {
		return nil, false
	}
	value := l.elements[0]
	l.elements = l.elements[1:]
	return value, true
}

func (l *List) PopBack() (interface{}, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.elements) == 0 {
		return nil, false
	}
	index := len(l.elements) - 1
	value := l.elements[index]
	l.elements = l.elements[:index]
	return value, true
}

func (l *List) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.elements)
}

func (l *List) Range(start, stop int) []interface{} {
	l.mu.RLockdefer l.mu.RUnlock()
	if start < 0 || stop > len(l.elements) || start > stop {
		return nil
	}
	return l.elements[start:stop]
}