package main

import (
	"flag"
	"fmt"
	"os"
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
)

var (
	buildFlag = flag.NewFlagSet(buildCmd, flag.ExitOnError)
	serveFlag = flag.NewFlagSet(serveCmd, flag.ExitOnError)
)

func main() {
	parseCommand()
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
