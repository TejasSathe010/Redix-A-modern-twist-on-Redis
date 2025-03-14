package network

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type ProtocolParser struct {
	r *bufio.Reader
}

func NewProtocolParser(r *bufio.Reader) *ProtocolParser {
	return &ProtocolParser{r: r}
}

func (p *ProtocolParser) ParseCommand() ([]string, error) {
	line, err := p.r.ReadString('\n')
	if err != nil {
		return nil, err
	}

	line = strings.TrimSuffix(line, "\r\n")

	if line == "" {
		return nil, fmt.Errorf("empty command")
	}

	return parseCommandString(line)
}

func parseCommandString(line string) ([]string, error) {
	var args []string
	var currentArg strings.Builder
	insideQuotes := false
	escapeNext := false

	for _, c := range line {
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

func (p *ProtocolParser) WriteResponse(w io.Writer, response interface{}) error {
	switch res := response.(type) {
	case string:
		_, err := fmt.Fprintf(w, "$%d\r\n%s\r\n", len(res), res)
		return err
	case int64:
		_, err := fmt.Fprintf(w, ":%d\r\n", res)
		return err
	case bool:
		if res {
			_, err := fmt.Fprintf(w, "+OK\r\n")
			return err
		}
		_, err := fmt.Fprintf(w, "-ERR Operation failed\r\n")
		return err
	case []string:
		_, err := fmt.Fprintf(w, "*%d\r\n", len(res))
		if err != nil {
			return err
		}
		for _, item := range res {
			_, err = fmt.Fprintf(w, "$%d\r\n%s\r\n", len(item), item)
			if err != nil {
				return err
			}
		}
		return nil
	case error:
		_, err := fmt.Fprintf(w, "-ERR %s\r\n", res.Error())
		return err
	default:
		_, err := fmt.Fprintf(w, "-ERR Unknown response type\r\n")
		return err
	}
}
