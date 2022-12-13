package local

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/rafaelmartins/b8r/internal/dataset"
	"github.com/rafaelmartins/b8r/internal/mime"
)

type LocalSource struct {
	mtx   sync.Mutex
	path  string
	cache *dataset.DataSet
}

func (f *LocalSource) Name() string {
	return "local"
}

func (f *LocalSource) SetParameter(key string, value string) error {
	if key == "path" {
		f.mtx.Lock()
		defer f.mtx.Unlock()

		info, err := os.Stat(value)
		if err != nil {
			return err
		}

		f.path = value
		f.cache = nil
		if info.IsDir() {
			f.cache = dataset.New()
		}

		return nil
	}
	return errors.New("local: invalid parameter")
}

func (f *LocalSource) List() ([]string, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	if f.path == "" {
		return nil, errors.New("local: missing path")
	}

	if f.cache == nil {
		return []string{path.Base(f.path)}, nil
	}

	entries, err := os.ReadDir(f.path)
	if err != nil {
		return nil, err
	}

	f.cache = dataset.New()
	for _, entry := range entries {
		if !entry.IsDir() {
			f.cache.Add(entry.Name())
		}
	}
	return f.cache.Slice(), nil
}

func (f *LocalSource) GetFile(key string) (string, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	if f.path == "" {
		return "", errors.New("local: missing path")
	}

	if f.cache == nil {
		return f.path, nil
	}

	if !f.cache.Contains(key) {
		return "", errors.New("local: invalid key")
	}

	return filepath.Join(f.path, key), nil
}

func (f *LocalSource) GetMimeType(key string) (string, error) {
	filename, err := f.GetFile(key)
	if err != nil {
		return "", err
	}
	return mime.Detect(filename)
}
