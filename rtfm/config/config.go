package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/aiven/terraform-provider-aiven/rtfm/exporter/extract"
	"github.com/aiven/terraform-provider-aiven/rtfm/exporter/provision"
	"github.com/aiven/terraform-provider-aiven/rtfm/server"
)

type Config struct {
	ShutdownTimeout time.Duration `default:"5s" envconfig:"SHUTDOWN_TIMEOUT"`

	ServerConfig    *server.Config
	ExportConfig    *extract.Config
	ProvisionConfig *provision.Config
}

func LoadConfig[T any]() (*T, error) {
	var cfg T
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &cfg, nil
}
