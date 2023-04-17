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
	ent := append([]string{}, entries...)
	if l := len(ent); l == 0 {
		ent = append(ent, ".")
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
					e, err := rel(f.root, path)
					if err != nil {
						return err
					}
					rv = append(rv, e)
				}

				return nil
			}); err != nil {
				return nil, false, err
			}
			continue
		}

		if info.Mode().IsRegular() {
			e, err := rel(f.root, entry)
			if err != nil {
				return nil, false, err
			}
			rv = append(rv, e)
			continue
		}
	}

	return rv, len(entries) == 1 && len(rv) == 1 && filepath.Clean(entries[0]) == filepath.Join(f.root, rv[0]), nil
}

func (f *LocalSource) GetFile(key string) (string, error) {
	return filepath.Join(f.root, key), nil
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
