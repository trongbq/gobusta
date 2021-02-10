package main

import (
	"log"

	"github.com/trongbq/gobusta/command"
)

func main() {
	c := command.NewCommand()
	err := c.Execute()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("DONE")
}
