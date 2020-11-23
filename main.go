package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/niklasfasching/go-org/org"
)

const usage = `Gobusta is a tool for building blog
Usage:
	gobusta <command> [arguments]

The commands are:

	build 	build and render entire source
	server	start a local server which serve generated files
	`

const (
	buildCmd = "build"
	serveCmd = "serve"

	sourceDirName   = "source"
	outDirName      = "dist"
	postDirName     = "posts"
	staticDirName   = "static"
	templateDirName = "templates"

	orgFileExt     = ".org"
	orgTitlePrefix = "#+title:"
	orgDatePrefix  = "#+date:"
)

type Post struct {
	Title        string
	Date         string
	URL          string
	Content      template.HTML
	OriginalFile string
}

var (
	buildFlag = flag.NewFlagSet(buildCmd, flag.ExitOnError)
	serveFlag = flag.NewFlagSet(serveCmd, flag.ExitOnError)
)

var (
	baseDir   string
	sourceDir string
	outDir    string
)

func init() {
	var err error
	baseDir, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	sourceDir = join(baseDir, sourceDirName)
	outDir = join(baseDir, outDirName)
}

func main() {
	parseCommand()

	if buildFlag.Parsed() {
		posts, err := collectPosts()
		if err != nil {
			panic(err)
		}
		cleanDir(outDir)
		err = generateIndexPage(posts)
		if err != nil {
			panic(err)
		}
		err = generatePosts(posts)
		if err != nil {
			panic(err)
		}
		err = copyDir(join(sourceDir, staticDirName), join(outDir, staticDirName))
		if err != nil {
			panic(err)
		}
	} else if serveFlag.Parsed() {
		// TODO: Add reload
		http.Handle("/", http.FileServer(http.Dir(outDir)))
		http.ListenAndServe(":8080", nil)
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

func generateIndexPage(ps []Post) error {
	t, err := newTemplate(join(sourceDir, templateDirName, "index.html"))
	if err != nil {
		return err
	}
	f, err := os.Create(join(outDir, "index.html"))
	if err != nil {
		return err
	}
	defer f.Close()
	return t.Execute(f, ps)
}

func generatePosts(ps []Post) error {
	t, err := newTemplate(join(sourceDir, templateDirName, "post.html"))
	if err != nil {
		return err
	}
	outPostDir := join(outDir, postDirName)
	err = os.Mkdir(outPostDir, 0755)
	if err != nil {
		return err
	}
	for _, p := range ps {
		f, err := os.Create(join(outPostDir, getFileName(p.OriginalFile)+".html"))
		if err != nil {
			return err
		}
		defer f.Close()

		err = t.Execute(f, p)
		if err != nil {
			return err
		}
	}
	return nil
}

func collectPosts() ([]Post, error) {
	var posts []Post
	err := filepath.Walk(join(sourceDir, postDirName), func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() && strings.HasSuffix(path, orgFileExt) {
			c, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			post := Post{}
			// Extract metadata
			// FIXME: Improve with flexibility
			cs := string(c)
			lines := strings.Split(cs, "\n")
			if len(lines) < 2 ||
				!strings.HasPrefix(lines[0], orgTitlePrefix) ||
				!strings.HasPrefix(lines[1], orgDatePrefix) {
				return err
			}
			post.Title = strings.Trim(lines[0][len(orgTitlePrefix):], " ")
			post.Date = strings.Trim(lines[1][len(orgDatePrefix):], " <>")

			// Render to HTML
			html, err := convertToHTML(strings.Join(lines[2:], "\n"))
			if err != nil {
				return err
			}
			post.Content = template.HTML(html)

			// Set URL
			post.URL = fmt.Sprintf("/%s/%s.html", postDirName, getFileName(path))

			post.OriginalFile = path

			posts = append(posts, post)

		}
		return nil
	})

	return posts, err
}

func convertToHTML(c string) (string, error) {
	writer := org.NewHTMLWriter()
	writer.HighlightCodeBlock = highlightCodeBlock
	orgConf := org.New()
	return orgConf.Parse(bytes.NewReader([]byte(c)), "").Write(writer)
}

func join(paths ...string) string {
	return strings.Join(paths, string(os.PathSeparator))
}

func newTemplate(file string) (*template.Template, error) {
	c, err := ioutil.ReadFile(file)
	if err != nil {
		return &template.Template{}, err
	}
	cs := string(c)
	return template.New(getFileName(file)).Parse(cs)
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
				os.RemoveAll(join(dir, info.Name()))
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
			err := copyDir(path, join(dest, info.Name()))
			if err != nil {
				return err
			}
		} else {
			err := copyFile(path, join(dest, info.Name()))
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

func getFileName(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func highlightCodeBlock(source, lang string, inline bool) string {
	var w strings.Builder
	l := lexers.Get(lang)
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)
	it, _ := l.Tokenise(nil, source)
	_ = html.New(html.WithLineNumbers(true)).Format(&w, styles.Get("github"), it)
	if inline {
		return `<div class="highlight-inline">` + "\n" + w.String() + "\n" + `</div>`
	}
	return `<div class="highlight">` + "\n" + w.String() + "\n" + `</div>`
}
