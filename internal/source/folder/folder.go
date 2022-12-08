package folder

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/rafaelmartins/b8r/internal/dataset"
	"github.com/rafaelmartins/b8r/internal/mime"
)

type FolderSource struct {
	path  string
	cache *dataset.DataSet
}

func (f *FolderSource) Name() string {
	return "folder"
}

func (f *FolderSource) SetParameter(key string, value string) error {
	if key == "path" {
		f.path = value
		return nil
	}
	return errors.New("folder: invalid parameter")
}

func (f *FolderSource) List() ([]string, error) {
	if f.path == "" {
		return nil, errors.New("folder: missing folder")
	}

	entries, err := os.ReadDir(f.path)
	if err != nil {
		return nil, err
	}

	f.cache = dataset.New()
	for _, entry := range entries {
		f.cache.Add(entry.Name())
	}
	return f.cache.Slice(), nil
}

func (f *FolderSource) GetFile(key string) (string, error) {
	if f.path == "" {
		return "", errors.New("folder: missing folder")
	}

	if !f.cache.Contains(key) {
		return "", errors.New("folder: invalid key")
	}

	return filepath.Join(f.path, key), nil
}

func (f *FolderSource) GetMimeType(key string) (string, error) {
	filename, err := f.GetFile(key)
	if err != nil {
		return "", err
	}
	return mime.Detect(filename)
}
