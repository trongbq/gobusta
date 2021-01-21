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
)

type Command struct {
	Build      *flag.FlagSet
	BuildClean *bool
	Serve      *flag.FlagSet
	ServePort  *int
}

func NewCommand() *Command {
	buildFlag := flag.NewFlagSet(buildCmd, flag.ExitOnError)
	buildClean := buildFlag.Bool("clean", false, "Clean the dist folder before the build")
	serveFlag := flag.NewFlagSet(serveCmd, flag.ExitOnError)
	servePort := serveFlag.Int("port", 8080, "Port number for serving server")
	c := Command{
		Build:      buildFlag,
		BuildClean: buildClean,
		Serve:      serveFlag,
		ServePort:  servePort,
	}
	c.parse()
	return &c
}

func (c Command) parse() {
	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]
	switch cmd {
	case buildCmd:
		c.Build.Parse(args)
	case serveCmd:
		c.Serve.Parse(args)
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func (c *Command) Execute() error {

	return nil
}
