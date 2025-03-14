package network

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

type Connection struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	parser *ProtocolParser
	active bool
	mu     sync.RWMutex
}

func NewConnection(conn net.Conn) *Connection {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	return &Connection{
		conn:   conn,
		reader: reader,
		writer: writer,
		parser: NewProtocolParser(reader),
		active: true,
	}
}

func (c *Connection) Start(handler CommandHandler) {
	go c.handleCommands(handler)
}

func (c *Connection) handleCommands(handler CommandHandler) {
	for c.active {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading command: %v\n", err)
			}
			c.Close()
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		args, err := c.parser.ParseCommand()
		if err != nil {
			c.writeError(err.Error())
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		response, err := handler.HandleCommand(ctx, args)
		if err != nil {
			c.writeError(err.Error())
			continue
		}

		c.writeResponse(response)
	}
}

func (c *Connection) writeResponse(response interface{}) {
	err := c.parser.WriteResponse(c.writer, response)
	if err != nil {
		fmt.Printf("Error writing response: %v\n", err)
		c.Close()
		return
	}

	err = c.writer.Flush()
	if err != nil {
		fmt.Printf("Error flushing writer: %v\n", err)
		c.Close()
		return
	}
}

func (c *Connection) writeError(message string) {
	_, err := fmt.Fprintf(c.writer, "-ERR %s\r\n", message)
	if err != nil {
		fmt.Printf("Error writing error response: %v\n", err)
		c.Close()
		return
	}

	err = c.writer.Flush()
	if err != nil {
		fmt.Printf("Error flushing writer: %v\n", err)
		c.Close()
		return
	}
}

func (c *Connection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.active {
		return
	}

	c.active = false
	c.conn.Close()
}
