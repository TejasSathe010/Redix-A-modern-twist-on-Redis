package replication

import (
	"context"
	"io"
	"net"
	"sync"

	"bufio"
	"fmt"

	"github.com/TejasSathe010/Redix-A-modern-twist-on-Redis/internal/network"
)

type Slave struct {
	masterAddr string
	conn       net.Conn
	mu         sync.Mutex
}

func NewSlave(masterAddr string) *Slave {
	return &Slave{
		masterAddr: masterAddr,
	}
}

func (s *Slave) Connect() error {
	var err error
	s.mu.Lock()
	defer s.mu.Unlock()
	s.conn, err = net.Dial("tcp", s.masterAddr)
	return err
}

func (s *Slave) SyncWithMaster(ctx context.Context) error {
	if err := s.Connect(); err != nil {
		return err
	}

	// Send SYNC command to master
	// In a real implementation, this would be more complex
	// and follow the actual Redis replication protocol

	// Create a parser for reading responses
	parser := network.NewProtocolParser(bufio.NewReader(s.conn))

	// Send command to master
	if err := s.SendCommand(ctx, []string{"SYNC"}); err != nil {
		return err
	}

	// Read and process the command log from master
	for {
		line, err := parser.Reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}

		// Process the command from master
		// This is a simplified implementation
		// In a real scenario, you would parse and execute the command
		fmt.Printf("Received command from master: %s", line)
	}
}

func (s *Slave) SendCommand(ctx context.Context, cmd []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn == nil {
		return fmt.Errorf("not connected to master")
	}

	// Format command for transmission
	// Using Redis protocol format
	writer := bufio.NewWriter(s.conn)

	// Start array
	writer.WriteString(fmt.Sprintf("*%d\r\n", len(cmd)))

	for _, arg := range cmd {
		writer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg))
	}

	writer.Flush()

	return nil
}

func (s *Slave) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
}
