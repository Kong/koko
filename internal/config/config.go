package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/ghodss/yaml"
	"github.com/ilyakaznacheev/cleanenv"
)

var (
	defaultConfigYAML = []byte(`
log:
  level: info
  format: json
admin:
  listeners:
  - address: ":3000"
    protocol: http
database:
  query_timeout: 5s
metrics:
  enabled: true
disable_anonymous_reports: false
`)
	defaultConfig Config
)

func init() {
	err := yaml.Unmarshal(defaultConfigYAML, &defaultConfig)
	if err != nil {
		panic(fmt.Errorf("failed to decode default config: %v", err))
	}
}

// Get constructs the Config using the filename, env vars and defaults.
func Get(filename string) (Config, error) {
	var c Config
	if filename != "" {
		if _, err := os.Stat(filename); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				filename = ""
			}
		}
	}
	var err error
	if filename == "" {
		err = cleanenv.ReadEnv(&c)
	} else {
		err = cleanenv.ReadConfig(filename, &c)
	}
	if err != nil {
		return Config{}, fmt.Errorf("unable to read config: %w", err)
	}
	return c, nil
}
