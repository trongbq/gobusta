package engine

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
)

func (e *engine) Build() error {
	log.Println("Start building content")
	posts, err := e.collectContent()
	if err != nil {
		return err
	}
	log.Println(posts)
	// Sort all posts by date desc
	sort.SliceStable(posts, func(i, j int) bool {
		return posts[i].PublishedAt.After(posts[j].PublishedAt)
	})
	err = e.renderContent(posts)
	if err != nil {
		return err
	}
	err = e.renderIndex(posts)
	if err != nil {
		return err
	}
	err = e.copyStatic()
	if err != nil {
		return err
	}
	return nil
}

func (e *engine) renderContent(posts []*Post) error {
	for _, p := range posts {
		outPath := filepath.Join(e.cf.Publish, p.URL)
		os.MkdirAll(filepath.Dir(outPath), os.ModePerm)
		f, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("error create content file, %v", err)
		}
		defer f.Close()
		err = e.tpl.ExecuteTemplate(f, "posthtml", *p)
		if err != nil {
			return fmt.Errorf("error render content file, %v", err)
		}
	}
	return nil
}

func (e *engine) renderIndex(posts []*Post) error {
	f, err := os.Create(filepath.Join(e.cf.Publish, "index.html"))
	if err != nil {
		return fmt.Errorf("error create index page: %v", err)
	}
	defer f.Close()
	err = e.tpl.ExecuteTemplate(f, "indexhtml", posts)
	if err != nil {
		return fmt.Errorf("error render index page: %v", err)
	}
	return nil
}

func (e *engine) copyStatic() error {
	return copyDir(e.cf.Static, filepath.Join(e.cf.Publish, e.cf.Static))
}
