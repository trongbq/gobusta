package command

type EngineConfig struct {
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

type Engine struct {
	conf *EngineConfig
}
