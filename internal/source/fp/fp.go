package fp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/danwakefield/fnmatch"
	"github.com/goccy/go-yaml"
)

type config struct {
	Handlers map[string][]string `yaml:"handlers"`
	Aliases  map[string]string   `yaml:"aliases"`
}

type filedata struct {
	Mimetype string `json:"mimetype"`
}

type FpSource struct{}

func (f *FpSource) Name() string {
	return "fp"
}

func (f *FpSource) Remote() bool {
	return true
}

func (f *FpSource) getConfig() (*config, error) {
	fn, ok := os.LookupEnv("FP_CONFIG")
	if !ok {
		u, err := user.Current()
		if err != nil {
			return nil, err
		}

		fn = filepath.Join(u.HomeDir, ".fp.yml")
	}

	fp, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	cfg := &config{}
	if err := yaml.NewDecoder(fp).Decode(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (f *FpSource) List(entries []string, recursive bool) ([]string, bool, error) {
	cfg, err := f.getConfig()
	if err != nil {
		return nil, false, err
	}

	all := []string{}
	for k := range cfg.Aliases {
		all = append(all, k)
	}

	if len(entries) == 0 {
		return all, false, nil
	}

	rv := []string{}
	for _, entry := range entries {
		found := false
		for _, a := range all {
			if fnmatch.Match(entry, a, 0) {
				found = true
				rv = append(rv, a)
			}
		}
		if !found {
			return nil, false, fmt.Errorf("fp: invalid entry: %s", entry)
		}
	}
	return rv, len(entries) == 1 && len(rv) == 1 && entries[0] == rv[0], nil
}

func (f *FpSource) GetFile(key string) (string, error) {
	cfg, err := f.getConfig()
	if err != nil {
		return "", err
	}

	if url, ok := cfg.Aliases[key]; ok {
		return url, nil
	}
	return "", fmt.Errorf("fp: invalid entry: %s", key)
}

func (f *FpSource) GetMimeType(key string) (string, error) {
	url, err := f.GetFile(key)
	if err != nil {
		return "", err
	}

	c, err := http.Get(url + ".json")
	if err != nil {
		return "", err
	}
	defer c.Body.Close()

	fd := &filedata{}
	if err := json.NewDecoder(c.Body).Decode(fd); err != nil {
		return "", err
	}

	return fd.Mimetype, nil
}

func (f *FpSource) CompletionHandler(prev string, cur string) []string {
	cfg, err := f.getConfig()
	if err != nil {
		return nil
	}

	rv := []string{}
	for e := range cfg.Aliases {
		if strings.HasPrefix(e, cur) {
			rv = append(rv, e)
		}
	}
	return rv
}
