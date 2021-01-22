package engine

import "log"

func (e *engine) Build() error {
	log.Println("Start building content")
	posts, err := e.collectContent()
	if err != nil {
		return err
	}
	log.Println(posts)
	return nil
}
