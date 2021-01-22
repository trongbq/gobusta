package command

import (
	"log"

	"github.com/trongbq/gobusta/engine"
	"github.com/trongbq/gobusta/server"
)

func (c *command) Execute() error {
	c.parse()

	if c.fc.build.Parsed() {
		e, err := engine.New(&engine.Config{
			Content:     content,
			Static:      static,
			Layout:      layouts,
			Publish:     publish,
			FmDelimeter: *c.fc.buildFmDelimeter,
		})
		if err != nil {
			log.Fatal(err)
		}
		err = e.Build()
		if err != nil {
			log.Fatal(err)
		}
	} else if c.fc.serve.Parsed() {
		s := server.New(&server.Config{
			Publish: "dist",
			Port:    *c.fc.servePort,
		})
		err := s.Serve()
		if err != nil {
			log.Fatal()
		}
	}

	return nil
}
