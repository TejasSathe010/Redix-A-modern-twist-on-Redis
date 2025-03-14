package replication

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

type Master struct {
	slaves      map[*SlaveConnection]struct{}
	mu          sync.RWMutex
	commandLog  []CommandEntry
	logMutex    sync.RWMutex
	shutdownCtx context.Context
}

type CommandEntry struct {
	Timestamp time.Time
	Command   []string
}

type SlaveConnection struct {
	conn net.Conn
	mu   sync.Mutex
}

func NewMaster() *Master {
	return &Master{
		slaves: make(map[*SlaveConnection]struct{}),
	}
}

func (m *Master) AddSlave(conn net.Conn) {
	sc := &SlaveConnection{conn: conn}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.slaves[sc] = struct{}{}
	go m.handleSlaveConnection(sc)
}

func (m *Master) handleSlaveConnection(sc *SlaveConnection) {
	defer func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		delete(m.slaves, sc)
		sc.conn.Close()
	}()

	for {
		// Send command log to slave
		m.logMutex.RLock()
		defer m.logMutex.RUnlock()
		for _, entry := range m.commandLog {
			// Send each command to the slave
			if err := sc.SendCommand(context.Background(), entry.Command); err != nil {
				return
			}
		}
	}
}

func (m *Master) BroadcastCommand(ctx context.Context, cmd []string) {
	m.logMutex.Lock()
	defer m.logMutex.Unlock()
	m.commandLog = append(m.commandLog, CommandEntry{
		Timestamp: time.Now(),
		Command:   cmd,
	})

	m.mu.RLock()
	defer m.mu.RUnlock()
	for sc := range m.slaves {
		go sc.SendCommand(ctx, cmd)
	}
}

func (m *Master) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for sc := range m.slaves {
		sc.conn.Close()
	}
}

func (sc *SlaveConnection) SendCommand(ctx context.Context, cmd []string) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if sc.conn == nil {
		return fmt.Errorf("connection closed")
	}

	// Format command for transmission using Redis protocol
	writer := bufio.NewWriter(sc.conn)

	// Start array
	writer.WriteString(fmt.Sprintf("*%d\r\n", len(cmd)))

	for _, arg := range cmd {
		writer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg))
	}

	writer.Flush()

	return nil
}
