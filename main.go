package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday/v2"
	"gopkg.in/yaml.v2"
)

const usage = `Gobusta is a handy tool for building minimal blog
Usage:
	gobusta <command> [arguments]

The commands are:

	build 	build and render entire source
	server	start a local server which serve generated files
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

type Post struct {
	Title   string
	Date    string
	Content string
	URL     string
}

func (p Post) RenderContent() template.HTML {
	return template.HTML(blackfriday.Run([]byte(p.Content)))
}

var (
	buildFlag = flag.NewFlagSet(buildCmd, flag.ExitOnError)
	serveFlag = flag.NewFlagSet(serveCmd, flag.ExitOnError)
)

var (
	baseDir    string
	contentDir string
	outDir     string
	templates  *template.Template

	ErrInvalidPostFormat = errors.New("Invalid post format")
)

func init() {
	log.Println("Init the program: collect metadata information")
	var err error
	baseDir, err = os.Getwd()
	if err != nil {
		panic(err)
	}
	contentDir = filepath.Join(baseDir, contentDirName)
	outDir = filepath.Join(baseDir, outDirName)
}

func main() {
	parseCommand()

	if buildFlag.Parsed() {
		// Get all templates in template directory
		var err error
		templates, err = prepareTemplates(filepath.Join(baseDir, templateDirName))
		if err != nil {
			panic(err)
		}
		posts, err := renderContentToHTML()
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
		err := http.ListenAndServe(":8081", nil)
		if err != nil {
			log.Fatalf("Can not start server: %v", err)
		}
		log.Println("Server is running on port 8080")
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

func renderContentToHTML() ([]Post, error) {
	log.Println("Render the content to HTML")
	var posts []Post
	type FrontMatter struct {
		Title string
		Date  string
	}
	err := filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
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
			URL:     path[len(contentDir):len(path)-len(".md")] + ".html",
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

// func cleanDir(dir string) error {
//     d, err := os.Stat(dir)
//     // Check if dir exists, then clean it
//     if err == nil {
//         if d.IsDir() {
//             // Clean output content
//             infos, err := ioutil.ReadDir(dir)
//             if err != nil {
//                 return err
//             }
//             for _, info := range infos {
//                 os.RemoveAll(filepath.Join(dir, info.Name()))
//             }
//             return nil
//         } else {
//             // If `out` is not a dir, then simply delete it
//             err := os.Remove(dir)
//             if err != nil {
//                 return err
//             }
//         }
//     }
//     return os.Mkdir(dir, 0755)
// }

func copyDir(src, dest string) error {
	srcStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	_, err = os.Stat(dest)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		return fmt.Errorf("destination dir already exists: %v", dest)
	}

	err = os.Mkdir(dest, srcStat.Mode())
	if err != nil {
		return err
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if path == src {
			return nil
		}

		if info.IsDir() {
			err := copyDir(path, filepath.Join(dest, info.Name()))
			if err != nil {
				return err
			}
		} else {
			err := copyFile(path, filepath.Join(dest, info.Name()))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	err = in.Close()
	if err != nil {
		return err
	}
	return out.Close()
}

// func highlightCodeBlock(source, lang string, inline bool) string {
//     var w strings.Builder
//     l := lexers.Get(lang)
//     if l == nil {
//         l = lexers.Fallback
//     }
//     l = chroma.Coalesce(l)
//     it, _ := l.Tokenise(nil, source)
//     _ = html.New(html.WithLineNumbers(true)).Format(&w, styles.Get("github"), it)
//     if inline {
//         return `<div class="highlight-inline">` + "\n" + w.String() + "\n" + `</div>`
//     }
//     return `<div class="highlight">` + "\n" + w.String() + "\n" + `</div>`
// }

func prepareTemplates(dir string) (*template.Template, error) {
	log.Printf("Preparing templates in directory %v\n", dir)
	var templates []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
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
