package local

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/rafaelmartins/b8r/internal/dataset"
	"github.com/rafaelmartins/b8r/internal/mime"
)

type LocalSource struct {
	mtx       sync.Mutex
	path      string
	isDir     bool
	recursive bool
	cache     *dataset.DataSet
}

func (f *LocalSource) Name() string {
	return "local"
}

func (f *LocalSource) SetParameter(key string, value interface{}) error {
	switch key {
	case "path":
		v, ok := value.(string)
		if !ok {
			return errors.New("local: path must be a string")
		}

		f.mtx.Lock()
		defer f.mtx.Unlock()

		info, err := os.Stat(v)
		if err != nil {
			return err
		}

		f.path = v
		f.isDir = info.IsDir()
		f.cache = nil

		return nil

	case "recursive":
		v, ok := value.(bool)
		if !ok {
			return errors.New("local: recursive must be a bool")
		}

		f.mtx.Lock()
		defer f.mtx.Unlock()

		f.recursive = v

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

	if !f.isDir {
		return []string{path.Base(f.path)}, nil
	}

	f.cache = dataset.New()
	root := true

	if err := filepath.WalkDir(f.path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			if root {
				root = false
				return nil
			}
			if !f.recursive {
				return fs.SkipDir
			}
			return nil
		}

		entry, err := filepath.Rel(f.path, path)
		if err != nil {
			return err
		}

		f.cache.Add(entry)
		return nil
	}); err != nil {
		return nil, err
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
