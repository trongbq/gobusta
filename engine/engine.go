package engine

import "text/template"

type Config struct {
	// Source of articles
	Content string
	// Static files
	Static string
	// Layout template for rendering
	Layout string
	// Publish destination
	Publish string
	// Delimeter for Front Matter
	FmDelimeter string
}

type engine struct {
	cf  *Config
	tpl *template.Template
}

func New(cf *Config) (*engine, error) {
	tpl, err := collectLayouts(cf.Layout)
	if err != nil {
		return nil, err
	}
	e := &engine{
		cf:  cf,
		tpl: tpl,
	}
	return e, nil
}
