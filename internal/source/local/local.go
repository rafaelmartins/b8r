package local

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rafaelmartins/b8r/internal/mime"
)

type LocalSource struct{}

func (f *LocalSource) Name() string {
	return "local"
}

func (f *LocalSource) Remote() bool {
	return false
}

func (f *LocalSource) List(entries []string, recursive bool) ([]string, bool, error) {
	ent := append([]string{}, entries...)
	single := false
	if l := len(ent); l == 0 {
		ent = append(ent, ".")
	} else if l == 1 {
		single = true
	}

	rv := []string{}
	for _, entry := range ent {
		info, err := os.Stat(entry)
		if err != nil {
			return nil, false, err
		}
		if info.IsDir() {
			single = false
			root := true

			if err := filepath.WalkDir(entry, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return nil
				}

				if d.IsDir() {
					if root {
						root = false
						return nil
					}
					if !recursive {
						return fs.SkipDir
					}
					return nil
				}

				info, err := d.Info()
				if err != nil {
					return err
				}

				if info.Mode().IsRegular() {
					rv = append(rv, path)
				}

				return nil
			}); err != nil {
				return nil, false, err
			}
			continue
		}

		if info.Mode().IsRegular() {
			rv = append(rv, entry)
			continue
		}

		single = false
	}

	return rv, single, nil
}

func (f *LocalSource) GetFile(key string) (string, error) {
	return key, nil
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
