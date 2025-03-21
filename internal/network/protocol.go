package network

import (
	"bufio"
	"fmt"
	"strings"
)

type ProtocolParser struct {
	reader *bufio.Reader
}

func NewProtocolParser(reader *bufio.Reader) *ProtocolParser {
	return &ProtocolParser{
		reader: reader,
	}
}

func (p *ProtocolParser) ReadLine() (string, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func (p *ProtocolParser) ParseCommand(line string) ([]string, error) {
	// Implement your command parsing logic here
	// This is a simple example that splits the line by spaces
	return strings.Fields(line), nil
}

func (p *ProtocolParser) WriteResponse(writer *bufio.Writer, response interface{}) error {
	switch res := response.(type) {
	case string:
		writer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(res), res))
	case int64:
		writer.WriteString(fmt.Sprintf(":%d\r\n", res))
	case bool:
		if res {
			writer.WriteString("+OK\r\n")
		} else {
			writer.WriteString("-ERR Operation failed\r\n")
		}
	case []string:
		writer.WriteString(fmt.Sprintf("*%d\r\n", len(res)))
		for _, item := range res {
			writer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(item), item))
		}
	case error:
		writer.WriteString(fmt.Sprintf("-ERR %s\r\n", res.Error()))
	default:
		writer.WriteString("-ERR Unknown response type\r\n")
	}

	writer.Flush()
	return nil
}
