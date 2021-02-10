package command

import (
	"flag"
	"fmt"
	"os"
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

	content = "content"
	static  = "static"
	layouts = "layouts"
	publish = "dist"
)

type flagCommand struct {
	build            *flag.FlagSet
	buildClean       *bool
	buildFmDelimeter *string
	serve            *flag.FlagSet
	servePort        *int
}

type command struct {
	fc *flagCommand
}

func NewCommand() *command {
	buildFlag := flag.NewFlagSet(buildCmd, flag.ExitOnError)
	buildClean := buildFlag.Bool("clean", false, "Clean the dist folder before the build")
	buildFmDelimeter := buildFlag.String("delimeter", "+++", "Delimeter for front matter")

	serveFlag := flag.NewFlagSet(serveCmd, flag.ExitOnError)
	servePort := serveFlag.Int("port", 8080, "Port number for serving server")

	c := command{
		fc: &flagCommand{
			build:            buildFlag,
			buildClean:       buildClean,
			buildFmDelimeter: buildFmDelimeter,
			serve:            serveFlag,
			servePort:        servePort,
		},
	}
	c.parse()
	return &c
}

func (c command) parse() {
	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]
	switch cmd {
	case buildCmd:
		c.fc.build.Parse(args)
	case serveCmd:
		c.fc.serve.Parse(args)
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}
