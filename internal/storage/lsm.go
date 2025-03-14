package storage

import (
	"os"
	"sync"
)

type LSMTree struct {
	memTable   *MemTable
	immutables []*MemTable
	disk       *DiskStorage
	mutex      sync.RWMutex
}

type DiskStorage struct {
	dir     string
	levels  []*Level
	current *MemTable
}

type Level struct {
	files []string
}

func NewLSMTree(dir string) (*LSMTree, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	disk := &DiskStorage{
		dir:    dir,
		levels: make([]*Level, 7),
	}

	for i := 0; i < 7; i++ {
		disk.levels[i] = &Level{}
	}

	return &LSMTree{
		memTable: NewMemTable(),
		disk:     disk,
	}, nil
}

func (l *LSMTree) Set(key string, value interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.memTable.Set(key, value)

	if l.memTable.Size() > 1000 {
		l.flushMemTable()
	}
}

func (l *LSMTree) Get(key string) (interface{}, bool) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	if value, exists := l.memTable.Get(key); exists {
		return value, exists
	}

	for _, immutable := range l.immutables {
		if value, exists := immutable.Get(key); exists {
			return value, exists
		}
	}

	// Check disk storage
	return l.disk.Get(key)
}

func (l *LSMTree) Delete(key string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.memTable.Delete(key)
}

func (l *LSMTree) flushMemTable() {
	immutable := l.memTable
	l.memTable = NewMemTable()
	l.immutables = append(l.immutables, immutable)

	go l.mergeLevels()
}

func (l *LSMTree) mergeLevels() {
	// Implement level merging logic
	// This is a simplified version
	for i := 0; i < len(l.disk.levels)-1; i++ {
		if len(l.disk.levels[i].files) > 10 {
			// Merge to next level
			l.mergeLevel(i, i+1)
		}
	}
}

func (l *LSMTree) mergeLevel(srcLevel, destLevel int) {
	// Implementation would merge files from srcLevel to destLevel
	// and then remove them from srcLevel
}

type MemTable struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

func NewMemTable() *MemTable {
	return &MemTable{
		data: make(map[string]interface{}),
	}
}

func (m *MemTable) Set(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

func (m *MemTable) Get(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.data[key]
	return value, exists
}

func (m *MemTable) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

func (m *MemTable) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data)
}
