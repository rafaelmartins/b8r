package local

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rafaelmartins/b8r/internal/mime"
)

type LocalSource struct {
	root string
}

func (f *LocalSource) Name() string {
	return "local"
}

func (f *LocalSource) Remote() bool {
	return false
}

func commonRoot(dirs []string) string {
	common := ""
	for i, d := range dirs {
		dir := filepath.Clean(d)
		if i == 0 {
			common = dir
			continue
		}

		for j, r := range dir {
			found := false
			for k, v := range common {
				if j == k && r == v {
					found = true
					break
				}
			}
			if found {
				continue
			}
			common = common[:j]
			break
		}
	}
	if common == "" {
		return ""
	}
	if info, err := os.Stat(common); err == nil && info.IsDir() {
		return filepath.Clean(common)
	}
	if f := filepath.Dir(common); f != "" && f != "." {
		return f
	}
	return ""
}

func rel(base string, target string) (string, error) {
	if base == "" {
		return target, nil
	}

	return filepath.Rel(base, target)
}

func (f *LocalSource) List(entries []string, recursive bool) ([]string, bool, error) {
	ent := []string{}
	for _, e := range entries {
		p, err := filepath.Abs(e)
		if err != nil {
			return nil, false, err
		}
		ent = append(ent, p)
	}
	if l := len(ent); l == 0 {
		d, err := os.Getwd()
		if err != nil {
			return nil, false, err
		}
		ent = append(ent, d)
	}

	f.root = commonRoot(ent)

	rv := []string{}
	for _, entry := range ent {
		info, err := os.Stat(entry)
		if err != nil {
			return nil, false, err
		}
		if info.IsDir() {
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
	}

	return rv, len(entries) == 1 && len(rv) == 1 && entries[0] == rv[0], nil
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

func (f *LocalSource) FormatEntry(key string) (string, error) {
	return rel(f.root, key)
}

func (f *LocalSource) SetItems(items []string) error {
	f.root = commonRoot(items)
	return nil
}
