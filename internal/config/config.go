package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type Preset struct {
	Name      string   `yaml:"name"`
	Source    string   `yaml:"source"`
	Include   *string  `yaml:"include"`
	Exclude   *string  `yaml:"exclude"`
	Entries   []string `yaml:"entries"`
	Mute      *bool    `yaml:"mute"`
	Random    *bool    `yaml:"random"`
	Recursive *bool    `yaml:"recursive"`
	Start     *bool    `yaml:"start"`
}

type Config struct {
	Presets []*Preset `yaml:"presets"`
}

func New() (*Config, error) {
	fn, ok := os.LookupEnv("B8R_CONFIG")
	if !ok {
		u, err := user.Current()
		if err != nil {
			return nil, err
		}

		fn = filepath.Join(u.HomeDir, ".b8r.yml")
	}

	f, err := os.Open(fn)
	if err != nil {
		return nil, fmt.Errorf("config: failed to open config file: %w", err)
	}
	defer f.Close()

	rv := &Config{}
	if err := yaml.NewDecoder(f).Decode(rv); err != nil {
		return nil, err
	}
	return rv, nil
}

func (c *Config) GetPreset(name string) *Preset {
	for _, pr := range c.Presets {
		if name == pr.Name {
			return pr
		}
	}
	return nil
}

func (c *Config) ListPresets() []string {
	rv := []string{}
	for _, pr := range c.Presets {
		rv = append(rv, pr.Name)
	}
	return rv
}
