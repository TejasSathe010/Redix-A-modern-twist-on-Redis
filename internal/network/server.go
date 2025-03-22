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

type Server struct {
	addr          string
	handler       CommandHandler
	shutdownCtx   context.Context
	cancelFunc    context.CancelFunc
	shutdownMutex sync.Mutex
}

type CommandHandler interface {
	HandleCommand(ctx context.Context, args []string) (interface{}, error)
}

func NewServer(addr string, handler CommandHandler) *Server {
	return &Server{
		addr:    addr,
		handler: handler,
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	shutdownCtx, cancelFunc := context.WithCancel(context.Background())
	s.shutdownCtx = shutdownCtx
	s.cancelFunc = cancelFunc

	fmt.Printf("Server listening on %s\n", s.addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.shutdownCtx.Done():
				fmt.Println("Server shutting down")
				return nil
			default:
				fmt.Printf("Error accepting connection: %v\n", err)
				continue
			}
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) Shutdown() {
	s.shutdownMutex.Lock()
	defer s.shutdownMutex.Unlock()

	if s.cancelFunc != nil {
		s.cancelFunc()
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		select {
		case <-s.shutdownCtx.Done():
			return
		default:
			// Read command using Redis protocol
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Error reading command: %v\n", err)
				}
				return
			}

			// Parse Redis protocol
			args, err := parseRedisCommand(line)
			if err != nil {
				writeError(writer, "Invalid command format")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			response, err := s.handler.HandleCommand(ctx, args)
			if err != nil {
				writeError(writer, err.Error())
			}

			writeResponse(writer, response)
		}
	}
}

func parseRedisCommand(data string) ([]string, error) {
	data = strings.TrimSuffix(data, "\r\n")

	// Split by spaces, but handle quotes
	var args []string
	var currentArg strings.Builder
	insideQuotes := false
	escapeNext := false

	for _, c := range data {
		if escapeNext {
			currentArg.WriteRune(c)
			escapeNext = false
			continue
		}

		switch c {
		case '\\':
			escapeNext = true
		case '"':
			insideQuotes = !insideQuotes
		case ' ':
			if !insideQuotes {
				args = append(args, currentArg.String())
				currentArg.Reset()
				continue
			}
			currentArg.WriteRune(c)
		default:
			currentArg.WriteRune(c)
		}
	}

	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	return args, nil
}

func writeBulkString(writer *bufio.Writer, value string) {
	writer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(value), value))
}

func writeError(writer *bufio.Writer, message string) {
	writer.WriteString(fmt.Sprintf("-ERR %s\r\n", message))
}

func writeSimpleString(writer *bufio.Writer, message string) {
	writer.WriteString(fmt.Sprintf("+%s\r\n", message))
}

func writeInteger(writer *bufio.Writer, value int64) {
	writer.WriteString(fmt.Sprintf(":%d\r\n", value))
}

func writeArrayStart(writer *bufio.Writer, length int) {
	writer.WriteString(fmt.Sprintf("*%d\r\n", length))
}

func writeResponse(writer *bufio.Writer, response interface{}) {
	switch res := response.(type) {
	case string:
		writeBulkString(writer, res)
	case int64:
		writeInteger(writer, res)
	case bool:
		if res {
			writeSimpleString(writer, "OK")
		} else {
			writeError(writer, "Operation failed")
		}
	case []string:
		writeArrayStart(writer, len(res))
		for _, item := range res {
			writeBulkString(writer, item)
		}
	case error:
		writeError(writer, res.Error())
	default:
		writeError(writer, "Unknown response type")
	}

	writer.Flush()
}
