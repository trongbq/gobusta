package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Depado/bfchroma"
	"github.com/alecthomas/chroma/styles"
	"github.com/russross/blackfriday/v2"
	"gopkg.in/yaml.v2"
)

const usage = `Gobusta is a handy tool for building minimal blog
Usage:
	gobusta <command> [arguments]

The commands are:

	build 	build and render entire source
	serve	start a local server which serve generated files

And few options
	build --clean to clean the build directory before build content
	serve --port [int] to specify the running port number
	`

const (
	buildCmd = "build"
	serveCmd = "serve"

	contentDirName         = "content"
	staticDirName          = "static"
	templateDirName        = "templates"
	templatePartialDirName = "templates/partials"

	outDirName = "dist"

	frontMatterDelimiter = "+++\n"
)

var (
	baseDir   string
	outDir    string
	templates *template.Template

	ErrInvalidPostFormat   = errors.New("Invalid post format")
	ErrContentDirNotExist  = errors.New("Content directory does not exist")
	ErrTemplateDirNotExist = errors.New("Template directory does not exist")

	chromaRenderer = bfchroma.NewRenderer(bfchroma.ChromaStyle(styles.GitHub))
)

type Post struct {
	Title   string
	Date    string
	Content string
	URL     string
}

func (p Post) RenderContent() template.HTML {
	return template.HTML(blackfriday.Run([]byte(p.Content), blackfriday.WithRenderer(chromaRenderer)))
}

var (
	buildFlag  = flag.NewFlagSet(buildCmd, flag.ExitOnError)
	buildClean = buildFlag.Bool("clean", false, "Clean the dist folder before the build")
	serveFlag  = flag.NewFlagSet(serveCmd, flag.ExitOnError)
	servePort  = serveFlag.Int("port", 8080, "Port number for serving server")
)

func init() {
	log.Println("Init the program: collect metadata information")
	var err error
	baseDir, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	outDir = filepath.Join(baseDir, outDirName)
}

func main() {
	parseCommand()

	if buildFlag.Parsed() {
		var err error
		if *buildClean {
			log.Println("Clean the dist before the build")
			err = cleanDir(outDir)
			if err != nil {
				panic(err)
			}
		}
		templates, err = prepareTemplates(filepath.Join(baseDir, templateDirName))
		if err != nil {
			panic(err)
		}
		posts, err := renderContentToHTML(filepath.Join(baseDir, contentDirName))
		if err != nil {
			panic(err)
		}
		err = renderIndexPage(posts)
		if err != nil {
			panic(err)
		}
		err = copyDir(filepath.Join(baseDir, staticDirName), filepath.Join(outDir, staticDirName))
		if err != nil {
			panic(err)
		}
	} else if serveFlag.Parsed() {
		// TODO: Add reload
		http.Handle("/", http.FileServer(http.Dir(outDir)))
		log.Printf("Server will be running on port %v", *servePort)
		err := http.ListenAndServe(fmt.Sprintf(":%v", *servePort), nil)
		if err != nil {
			log.Fatalf("Can not start server: %v", err)
		}
	}
}

func parseCommand() {
	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]
	switch cmd {
	case buildCmd:
		buildFlag.Parse(args)
	case serveCmd:
		serveFlag.Parse(args)
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func renderContentToHTML(dir string) ([]Post, error) {
	log.Println("Render the content to HTML")
	exist, err := exists(dir)
	if err != nil {
		log.Printf("Something wrong with the content directory %v", err)
		return nil, err
	}
	if !exist {
		return nil, ErrTemplateDirNotExist
	}

	var posts []Post
	type FrontMatter struct {
		Title string
		Date  string
	}
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		log.Printf("Processing content file %v", path)
		file, err := os.Open(path)
		if err != nil {
			return err
		}

		// Read the content file
		fmData, content, err := splitPostContent(file)
		if err != nil {
			return err
		}
		var fm FrontMatter
		err = yaml.Unmarshal([]byte(fmData), &fm)
		if err != nil {
			return err
		}
		p := Post{
			Title:   fm.Title,
			Date:    fm.Date,
			Content: content,
			URL:     path[len(dir):len(path)-len(".md")] + ".html",
		}
		posts = append(posts, p)

		// Render the content to HTML file
		outPath := filepath.Join(outDir, p.URL)
		os.MkdirAll(filepath.Dir(outPath), os.ModePerm)
		f, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer f.Close()
		err = templates.ExecuteTemplate(f, "posthtml", p)
		if err != nil {
			return err
		}
		return nil
	})
	return posts, err
}

func splitPostContent(input io.Reader) (string, string, error) {
	splitFunc := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		fmd := []byte(frontMatterDelimiter)
		dl := len(frontMatterDelimiter)
		if i := bytes.Index(data, fmd); i >= 0 {
			// Found the first delimiter, find second one
			if j := bytes.Index(data[i+dl:], fmd); j >= 0 {
				return i + dl + j + dl, bytes.TrimSpace(data[i+dl : j+dl]), nil
			}
			// Can not find second delimiter, return error
			return 0, nil, ErrInvalidPostFormat
		}
		if atEOF {
			return len(data), bytes.TrimSpace(data), nil
		}
		return 0, nil, nil
	}
	s := bufio.NewScanner(input)
	s.Split(splitFunc)

	s.Scan()
	front := s.Text()
	s.Scan()
	content := s.Text()

	if err := s.Err(); err != nil {
		return "", "", err
	}
	return front, content, nil
}

func renderIndexPage(ps []Post) error {
	f, err := os.Create(filepath.Join(outDir, "index.html"))
	if err != nil {
		return err
	}
	defer f.Close()
	return templates.ExecuteTemplate(f, "indexhtml", ps)
}

func cleanDir(dir string) error {
	d, err := os.Stat(dir)
	// Check if dir exists, then clean it
	if err == nil {
		if d.IsDir() {
			// Clean output content
			infos, err := ioutil.ReadDir(dir)
			if err != nil {
				return err
			}
			for _, info := range infos {
				os.RemoveAll(filepath.Join(dir, info.Name()))
			}
			return nil
		} else {
			// If `out` is not a dir, then simply delete it
			err := os.Remove(dir)
			if err != nil {
				return err
			}
		}
	}
	return os.Mkdir(dir, 0755)
}

func copyDir(src, dest string) error {
	// Make sure source directory exists
	srcStat, err := os.Stat(src)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err != nil {
		return errors.New(fmt.Sprintf("Can not copy from %v to %v, source does not exist", src, dest, src))
	}

	// Check destination is valid, create new if it does not exist
	_, err = os.Stat(dest)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err != nil {
		// Destination does not exists, create it
		err = os.Mkdir(dest, srcStat.Mode())
		if err != nil {
			return err
		}
	}
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		log.Printf("Copy file %v", path[len(src):])
		return copyFile(path, filepath.Join(dest, path[len(src):]))
	})
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		log.Printf("Can not open source file %v", src)
		return err
	}

	// Make sure to create parent directory first before creating the file
	os.MkdirAll(filepath.Dir(dest), os.ModePerm)

	out, err := os.Create(dest)
	if err != nil {
		log.Printf("Can not create destination file %v", dest)
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		log.Printf("Can not copy file from %v to %v", src, dest)
		return err
	}

	err = in.Close()
	if err != nil {
		return err
	}
	return out.Close()
}

func prepareTemplates(dir string) (*template.Template, error) {
	log.Printf("Preparing templates in directory %v\n", dir)
	exist, err := exists(dir)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, ErrTemplateDirNotExist
	}

	var templates []string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}
		log.Printf("Got template file: %v", path)
		templates = append(templates, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return template.New("").ParseFiles(templates...)
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	if err == nil {
		return true, nil
	}
	return false, nil
}
