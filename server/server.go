package server

import (
	"fmt"
	"log"
	"net/http"
)

type Config struct {
	Publish string
	Port    int
}

type server struct {
	cf *Config
}

func New(cf *Config) *server {
	return &server{
		cf: cf,
	}
}

func (s *server) Serve() error {
	http.Handle("/", http.FileServer(http.Dir(s.cf.Publish)))
	log.Printf("Server will be running on port %v\n", s.cf.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%v", s.cf.Port), nil)
	return err
}
