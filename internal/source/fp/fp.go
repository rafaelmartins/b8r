package fp

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"sync"

	"github.com/rafaelmartins/b8r/internal/dataset"
	"gopkg.in/yaml.v2"
)

type config struct {
	Handlers map[string][]string `yaml:"handlers"`
	Aliases  map[string]string   `yaml:"aliases"`
}

type filedata struct {
	Mimetype string `json:"mimetype"`
}

type FpSource struct {
	mtx    sync.Mutex
	config *config
	cache  *dataset.DataSet
}

func (f *FpSource) Name() string {
	return "fp"
}

func (f *FpSource) SetParameter(key string, value string) error {
	return errors.New("fp: invalid parameter")
}

func (f *FpSource) List() ([]string, error) {
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
	f.cache = dataset.New()
	for entry := range cfg.Aliases {
		f.cache.Add(entry)
	}
	return f.cache.Slice(), nil
}

func (f *FpSource) GetFile(key string) (string, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	if !f.cache.Contains(key) {
		return "", errors.New("fp: invalid alias")
	}

	if url, ok := f.config.Aliases[key]; ok {
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
