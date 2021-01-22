package engine

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

func collectLayouts(l string) (*template.Template, error) {
	log.Printf("Collecting layout files in %v folder\n", l)
	exist, err := pathExists(l)
	if err != nil {
		return nil, fmt.Errorf("accessing %v folder error %v", l, err)
	}
	if !exist {
		return nil, fmt.Errorf("%v folder does not exist", l)
	}
	var layouts []string
	err = filepath.Walk(l, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		layouts = append(layouts, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error collecting layout files %v", err)
	}
	tpl, err := template.New("").ParseFiles(layouts...)
	if err != nil {
		return nil, fmt.Errorf("Error parsing layout files: %v", err)
	}
	log.Println("Collecting layout files done!")
	return tpl, nil
}
