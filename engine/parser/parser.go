package parser

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type FrontMatter struct {
	Title string
	Date  string
	Tags  []string
}

type Parser struct {
	// Front matter delimiter
	fd string
}

func NewParser(fd string) *Parser {
	return &Parser{fd}
}

// Parse parses input into front matter and content string
func (p *Parser) Parse(in io.Reader) (*FrontMatter, string, error) {
	splitFunc := func(data []byte, eof bool) (advance int, token []byte, err error) {
		if eof && len(data) == 0 {
			return 0, nil, nil
		}
		fmd := []byte(p.fd)
		fmdl := len(p.fd)
		// log.Println("hello", eof, string(data))
		if i := bytes.Index(data, fmd); i >= 0 {
			return i + fmdl, bytes.TrimSpace(data[:i]), nil
		}
		if eof {
			return 0, nil, nil
		}
		return len(data), bytes.TrimSpace(data), nil
	}

	s := bufio.NewScanner(in)
	s.Split(splitFunc)
	s.Scan()
	fmContent := s.Text()
	s.Scan()
	content := s.Text()
	if err := s.Err(); err != nil {
		return nil, "", fmt.Errorf("can not parse content: %v", err)
	}
	var fm FrontMatter
	if err := json.Unmarshal([]byte(fmContent), &fm); err != nil {
		return nil, "", fmt.Errorf("can not parse front matter: %v", err)
	}
	return &fm, content, nil
}
