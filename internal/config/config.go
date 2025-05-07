package config

import (
	"fmt"
	"os"
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
	Standalone struct {
		SerialNumber string `yaml:"serial-number"`
	} `yaml:"standalone"`

	MpvPlugin struct {
		SerialNumber string `yaml:"serial-number"`
	} `yaml:"mpv-plugin"`

	Presets []*Preset `yaml:"presets"`

	dir string
}

func New() (*Config, error) {
	dir, found := os.LookupEnv("B8R_CONFIGDIR")
	if !found {
		if home, err := os.UserHomeDir(); err == nil {
			d := filepath.Join(home, ".b8r")
			if st, err := os.Stat(d); err == nil && st.IsDir() {
				dir = d
			} else {
				cdir, err := os.UserConfigDir()
				if err != nil {
					return nil, err
				}
				dir = filepath.Join(cdir, "b8r")
			}
		}
	}
	if dir == "" {
		return nil, fmt.Errorf("config: failed to discover configuration directory")
	}
	if err := os.MkdirAll(dir, 0777); err != nil {
		return nil, err
	}

	f, err := os.Open(filepath.Join(dir, "config.yml"))
	if err != nil {
		return nil, fmt.Errorf("config: failed to open config file: %w", err)
	}
	defer f.Close()

	rv := &Config{
		dir: dir,
	}
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
