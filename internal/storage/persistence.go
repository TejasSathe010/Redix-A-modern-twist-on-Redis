package storage

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type PersistenceLayer struct {
	walFile     *os.File
	snapshotDir string
	memTable    *MemTable
	mutex       sync.Mutex
}

type MemTable struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

func NewPersistenceLayer(baseDir string) (*PersistenceLayer, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}

	walPath := filepath.Join(baseDir, "wal.log")
	walFile, err := os.OpenFile(walPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	memTable := &MemTable{
		data: make(map[string]interface{}),
	}

	return &PersistenceLayer{
		walFile:     walFile,
		snapshotDir: baseDir,
		memTable:    memTable,
	}, nil
}

func (p *PersistenceLayer) Set(key string, value interface{}) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Write to WAL
	err := p.writeToWAL(key, value)
	if err != nil {
		return err
	}

	// Update memtable
	p.memTable.data[key] = value

	// Check if we need to flush to disk
	if len(p.memTable.data) > 1000 {
		go p.flushToDisk()
	}

	return nil
}

func (p *PersistenceLayer) Get(key string) (interface{}, bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	value, exists := p.memTable.data[key]
	return value, exists
}

func (p *PersistenceLayer) Delete(key string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Write delete marker to WAL
	err := p.writeDeleteToWAL(key)
	if err != nil {
		return err
	}

	// Remove from memtable
	delete(p.memTable.data, key)

	return nil
}

func (p *PersistenceLayer) writeToWAL(key string, value interface{}) error {
	// Create a record with timestamp and operation type
	var buf []byte
	// Add timestamp
	ts := uint64(time.Now().UnixNano())
	tsBytes := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(tsBytes, ts)
	buf = append(buf, tsBytes[:n]...)

	// Add operation type (SET)
	buf = append(buf, 'S')

	// Add key
	keyBytes := []byte(key)
	keyLenBytes := make([]byte, binary.MaxVarintLen64)
	n = binary.PutUvarint(keyLenBytes, uint64(len(keyBytes)))
	buf = append(buf, keyLenBytes[:n]...)
	buf = append(buf, keyBytes...)

	// Add value
	valueBytes, ok := value.([]byte)
	if !ok {
		valueBytes = []byte(fmt.Sprintf("%v", value))
	}
	valueLenBytes := make([]byte, binary.MaxVarintLen64)
	n = binary.PutUvarint(valueLenBytes, uint64(len(valueBytes)))
	buf = append(buf, valueLenBytes[:n]...)
	buf = append(buf, valueBytes...)

	// Write to WAL file
	_, err := p.walFile.Write(buf)
	if err != nil {
		return err
	}

	return p.walFile.Sync()
}

func (p *PersistenceLayer) writeDeleteToWAL(key string) error {
	// Create a record with timestamp and operation type
	var buf []byte
	// Add timestamp
	ts := uint64(time.Now().UnixNano())
	tsBytes := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(tsBytes, ts)
	buf = append(buf, tsBytes[:n]...)

	// Add operation type (DELETE)
	buf = append(buf, 'D')

	// Add key
	keyBytes := []byte(key)
	keyLenBytes := make([]byte, binary.MaxVarintLen64)
	n = binary.PutUvarint(keyLenBytes, uint64(len(keyBytes)))
	buf = append(buf, keyLenBytes[:n]...)
	buf = append(buf, keyBytes...)

	// Write to WAL file
	_, err := p.walFile.Write(buf)
	if err != nil {
		return err
	}

	return p.walFile.Sync()
}

func (p *PersistenceLayer) flushToDisk() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Create a new snapshot file
	snapshotPath := filepath.Join(p.snapshotDir, fmt.Sprintf("snapshot-%d.sst", time.Now().UnixNano()))
	snapshotFile, err := os.Create(snapshotPath)
	if err != nil {
		return err
	}
	defer snapshotFile.Close()

	// Write all key-value pairs to the snapshot
	for key, value := range p.memTable.data {
		// Format: key length (varint) + key + value length (varint) + value
		keyBytes := []byte(key)
		keyLenBytes := make([]byte, binary.MaxVarintLen64)
		n := binary.PutUvarint(keyLenBytes, uint64(len(keyBytes)))
		snapshotFile.Write(keyLenBytes[:n])
		snapshotFile.Write(keyBytes)

		valueBytes, ok := value.([]byte)
		if !ok {
			valueBytes = []byte(fmt.Sprintf("%v", value))
		}
		valueLenBytes := make([]byte, binary.MaxVarintLen64)
		n = binary.PutUvarint(valueLenBytes, uint64(len(valueBytes)))
		snapshotFile.Write(valueLenBytes[:n])
		snapshotFile.Write(valueBytes)
	}

	// After successful snapshot, clear the memtable
	p.memTable.data = make(map[string]interface{})

	return nil
}

func (p *PersistenceLayer) Recover() error {
	// Recover from previous WAL if needed
	return recoverFromWAL(p.walFile.Name(), p.memTable)
}

func recoverFromWAL(walPath string, memTable *MemTable) error {
	file, err := os.Open(walPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No WAL file yet
		}
		return err
	}
	defer file.Close()

	var buf [65536]byte
	for {
		n, err := file.Read(buf[:])
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		pos := 0
		for pos < n {
			// Read timestamp
			var ts uint64
			ts, pos = binary.Uvarint(buf[pos:])
			if pos == 0 {
				return fmt.Errorf("invalid WAL record", ts)
			}

			// Read operation type
			if pos >= n {
				break
			}
			op := buf[pos]
			pos++

			// Process SET operation
			if op == 'S' {
				// Find key length
				var keyLen uint64
				keyLen, pos = binary.Uvarint(buf[pos:])
				if pos == 0 {
					return fmt.Errorf("invalid WAL record")
				}

				// Read key
				if pos+int(keyLen) > n {
					break
				}
				key := string(buf[pos : pos+int(keyLen)])
				pos += int(keyLen)

				// Find value length
				var valueLen uint64
				valueLen, pos = binary.Uvarint(buf[pos:])
				if pos == 0 {
					return fmt.Errorf("invalid WAL record")
				}

				// Read value
				if pos+int(valueLen) > n {
					break
				}
				value := buf[pos : pos+int(valueLen)]
				pos += int(valueLen)

				memTable.data[key] = value
			} else if op == 'D' { // DELETE operation
				// Find key length
				var keyLen uint64
				keyLen, pos = binary.Uvarint(buf[pos:])
				if pos == 0 {
					return fmt.Errorf("invalid WAL record")
				}

				// Read key
				if pos+int(keyLen) > n {
					break
				}
				key := string(buf[pos : pos+int(keyLen)])
				pos += int(keyLen)

				delete(memTable.data, key)
			}
		}
	}

	return nil
}

// NewMemTable creates a new instance of MemTable
func NewMemTable() *MemTable {
	return &MemTable{
		data: make(map[string]interface{}),
	}
}

// Set adds or updates a key-value pair in the MemTable
func (m *MemTable) Set(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

// Get retrieves a value by key from the MemTable
func (m *MemTable) Get(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.data[key]
	return value, exists
}

// Delete removes a key-value pair from the MemTable
func (m *MemTable) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

// Size returns the number of key-value pairs in the MemTable
func (m *MemTable) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data)
}
