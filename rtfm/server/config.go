package server

import (
	"fmt"
	"time"
)

type Config struct {
	Host         string        `default:"localhost" envconfig:"HTTP_HOST"`
	Port         int           `default:"8080"            envconfig:"HTTP_PORT"`
	ReadTimeout  time.Duration `default:"5s" envconfig:"HTTP_READ_TIMEOUT"`
	WriteTimeout time.Duration `default:"10s" envconfig:"HTTP_WRITE_TIMEOUT"`
}

// Address returns the full address string for the server
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
