package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

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
