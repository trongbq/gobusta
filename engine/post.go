package engine

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/trongbq/gobusta/engine/parser"
)

type PostPublishedDate time.Time

type Post struct {
	Title       string
	PublishedAt time.Time
	Tags        []tag
	Content     string
	URL         string
}

type tag struct {
	name string
}

func (e *engine) collectContent() ([]*Post, error) {
	var posts []*Post
	exist, err := pathExists(e.cf.Content)
	if err != nil {
		return nil, fmt.Errorf("accessing %v folder error %v", e.cf.Content, err)
	}
	if !exist {
		return nil, fmt.Errorf("%v folder does not exist", e.cf.Content)
	}
	p := parser.NewParser(e.cf.FmDelimeter)
	err = filepath.Walk(e.cf.Content, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Only markdown is supported
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		// Separate front matter and content
		fm, content, err := p.Parse(f)
		if err != nil {
			return fmt.Errorf("%v: %v", path, err)
		}
		log.Println(fm)
		log.Println(content)
		return nil
	})
	return posts, err
}
