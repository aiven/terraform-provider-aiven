package provision

type Config struct {
	TemplatesDir string `default:"./rtfm/exporter/provision/templates" envconfig:"TEMPLATES_DIR"`
}
