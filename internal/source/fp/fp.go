package fp

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rafaelmartins/b8r/internal/dataset"
	"gopkg.in/yaml.v3"
)

type config struct {
	Handlers map[string][]string `yaml:"handlers"`
	Aliases  map[string]string   `yaml:"aliases"`
}

type filedata struct {
	Mimetype string `json:"mimetype"`
}

type FpSource struct {
	mtx        sync.Mutex
	entry      string
	config     *config
	cache      *dataset.DataSet
	singleItem bool
}

func (f *FpSource) Name() string {
	return "fp"
}

func (f *FpSource) PreFilterMimeType() bool {
	return false
}

func (f *FpSource) IsSingleItem() bool {
	return f.singleItem
}

func (f *FpSource) SetParameter(key string, value interface{}) error {
	switch key {
	case "entry":
		v, ok := value.(string)
		if !ok {
			return errors.New("fp: entry must be a string")
		}

		cfg, err := f.getConfig()
		if err != nil {
			return err
		}

		f.mtx.Lock()
		defer f.mtx.Unlock()

		if v != "" {
			_, f.singleItem = cfg.Aliases[v]
		}
		f.entry = v
	}

	return nil
}

func (f *FpSource) getConfig() (*config, error) {
	if f.config != nil {
		return f.config, nil
	}

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

	f.mtx.Lock()
	defer f.mtx.Unlock()

	f.config = cfg

	return cfg, nil
}

func (f *FpSource) List() ([]string, error) {
	cfg, err := f.getConfig()
	if err != nil {
		return nil, err
	}

	f.mtx.Lock()
	defer f.mtx.Unlock()

	f.cache = dataset.New()

	if f.entry == "" {
		for entry := range cfg.Aliases {
			f.cache.Add(entry)
		}
	} else if _, ok := cfg.Aliases[f.entry]; ok {
		f.cache.Add(f.entry)
	}

	return f.cache.Slice(), nil
}

func (f *FpSource) GetFile(key string) (string, error) {
	cfg, err := f.getConfig()
	if err != nil {
		return "", err
	}

	f.mtx.Lock()
	defer f.mtx.Unlock()

	if !f.cache.Contains(key) {
		return "", errors.New("fp: invalid alias")
	}

	if url, ok := cfg.Aliases[key]; ok {
		return url, nil
	}
	return "", errors.New("fp: invalid alias")
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
