package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/imdario/mergo"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

var (
	Levels = map[string]zapcore.Level{
		"debug": zapcore.DebugLevel,
		"info":  zapcore.InfoLevel,
		"warn":  zapcore.WarnLevel,
		"error": zapcore.ErrorLevel,
	}

	defaultConfigYAML = []byte(`
log:
  level: info
admin:
  listeners:
  - address: ":3000"
    protocol: http
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
	if filename == "" {
		return defaultConfig, nil
	}
	content, err := ioutil.ReadFile(filepath.Clean(filename))
	if err != nil {
		return Config{}, fmt.Errorf("reading file '%v': %w", filename, err)
	}

	result, err := parse(content)
	if err != nil {
		return Config{}, fmt.Errorf("parsing file '%v': %w", filename, err)
	}

	err = mergo.Merge(&result, defaultConfig)
	if err != nil {
		return Config{}, fmt.Errorf("merging defaults: %w", err)
	}
	return result, nil
}

func parse(content []byte) (Config, error) {
	contentAsString := string(content)

	var result Config
	err := yaml.Unmarshal([]byte(contentAsString), &result)
	if err != nil {
		return Config{}, err
	}
	return result, nil
}
