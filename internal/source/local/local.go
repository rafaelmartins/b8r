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
	entry     string
	isDir     bool
	recursive bool
	cache     *dataset.DataSet
}

func (f *LocalSource) Name() string {
	return "local"
}

func (f *LocalSource) PreFilterMimeType() bool {
	return true
}

func (f *LocalSource) IsSingleItem() bool {
	return !f.isDir
}

func (f *LocalSource) SetParameter(key string, value interface{}) error {
	switch key {
	case "entry":
		v, ok := value.(string)
		if !ok {
			return errors.New("local: entry must be a string")
		}

		f.mtx.Lock()
		defer f.mtx.Unlock()

		if v == "" {
			v = "."
		}

		info, err := os.Stat(v)
		if err != nil {
			return err
		}

		f.entry = v
		f.isDir = info.IsDir()
		f.cache = nil

	case "recursive":
		v, ok := value.(bool)
		if !ok {
			return errors.New("local: recursive must be a bool")
		}

		f.mtx.Lock()
		defer f.mtx.Unlock()

		f.recursive = v
	}

	return nil
}

func (f *LocalSource) List() ([]string, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	if f.entry == "" {
		return nil, errors.New("local: missing entry")
	}

	if !f.isDir {
		return []string{path.Base(f.entry)}, nil
	}

	f.cache = dataset.New()
	root := true

	if err := filepath.WalkDir(f.entry, func(path string, d fs.DirEntry, err error) error {
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

		entry, err := filepath.Rel(f.entry, path)
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

	if f.entry == "" {
		return "", errors.New("local: missing entry")
	}

	if f.cache == nil {
		return f.entry, nil
	}

	if !f.cache.Contains(key) {
		return "", errors.New("local: invalid key")
	}

	return filepath.Join(f.entry, key), nil
}

func (f *LocalSource) GetMimeType(key string) (string, error) {
	filename, err := f.GetFile(key)
	if err != nil {
		return "", err
	}
	return mime.Detect(filename)
}

func (f *LocalSource) CompletionHandler(prev string, cur string) []string {
	// empty list means that bash will list files
	return nil
}
