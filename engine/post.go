package engine

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/trongbq/gobusta/engine/parser"

	"github.com/Depado/bfchroma"
	"github.com/alecthomas/chroma/styles"
	"github.com/russross/blackfriday/v2"
)

var chromaRenderer = bfchroma.NewRenderer(bfchroma.ChromaStyle(styles.GitHub))

type Post struct {
	Title       string
	PublishedAt time.Time
	Tags        []string
	Content     string
	URL         string
}

func (p Post) RenderContent() template.HTML {
	return template.HTML(blackfriday.Run([]byte(p.Content), blackfriday.WithRenderer(chromaRenderer)))
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
		date, err := time.Parse("2006-01-02", fm.Date)
		if err != nil {
			return fmt.Errorf("%v: parse date error %v", path, err)
		}
		p := Post{
			Title:       fm.Title,
			PublishedAt: date,
			Tags:        fm.Tags,
			Content:     content,
			URL:         path[len(e.cf.Content):len(path)-len(".md")] + ".html",
		}
		posts = append(posts, &p)
		return nil
	})
	return posts, err
}
