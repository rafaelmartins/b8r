package presets

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Preset struct {
	Name      string `yaml:"name"`
	Source    string `yaml:"source"`
	Include   string `yaml:"include"`
	Exclude   string `yaml:"exclude"`
	Entry     string `yaml:"entry"`
	Mute      bool   `yaml:"mute"`
	Random    bool   `yaml:"random"`
	Recursive bool   `yaml:"recursive"`
	Start     bool   `yaml:"start"`
}

type Presets []*Preset

func New() (Presets, error) {
	fn, ok := os.LookupEnv("B8R_PRESETS")
	if !ok {
		u, err := user.Current()
		if err != nil {
			return nil, err
		}

		fn = filepath.Join(u.HomeDir, ".b8r-presets.yml")
	}

	f, err := os.Open(fn)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && !ok {
			return Presets{}, nil
		}
		return nil, err
	}
	defer f.Close()

	rv := Presets{}
	if err := yaml.NewDecoder(f).Decode(&rv); err != nil {
		return nil, err
	}
	return rv, nil
}

func (p Presets) Get(name string) *Preset {
	for _, pr := range p {
		if name == pr.Name {
			return pr
		}
	}
	return nil
}

func (p Presets) List() []string {
	rv := []string{}
	for _, pr := range p {
		rv = append(rv, pr.Name)
	}
	return rv
}
