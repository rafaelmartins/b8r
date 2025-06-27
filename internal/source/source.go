package source

import (
	"errors"
	"regexp"
	"sort"
	"strings"

	"github.com/rafaelmartins/b8r/internal/dataset"
	"github.com/rafaelmartins/b8r/internal/source/fp"
	"github.com/rafaelmartins/b8r/internal/source/local"
)

type SourceBackend interface {
	Name() string
	Remote() bool
	List(entries []string, recursive bool) ([]string, bool, error)
	GetFile(key string) (string, error)
	GetMimeType(key string) (string, error)
	CompletionHandler(prev string, cur string) []string
	FormatItem(key string) (string, error)
	SetItems(items []string) error
}

var registry = []SourceBackend{
	&local.LocalSource{},
	&fp.FpSource{},
}

type Source struct {
	backend SourceBackend
	items   *dataset.DataSet
}

func List() []string {
	rv := []string{}
	for _, b := range registry {
		rv = append(rv, b.Name())
	}
	return rv
}

func CompletionHandler(prev string, cur string) []string {
	for _, b := range registry {
		if b.Name() == prev {
			return b.CompletionHandler(prev, cur)
		}
	}
	return nil
}

func New(name string) (*Source, error) {
	var backend SourceBackend
	for _, r := range registry {
		if r.Name() == name {
			backend = r
			break
		}
	}

	if backend == nil {
		return nil, errors.New("source: not found")
	}

	return &Source{
		backend: backend,
	}, nil
}

func toInclude(r *regexp.Regexp, e string) bool {
	if r == nil {
		return true
	}
	return r.MatchString(e)
}

func toExclude(r *regexp.Regexp, e string) bool {
	if r == nil {
		return false
	}
	return r.MatchString(e)
}

func (s *Source) GetBackendName() string {
	return s.backend.Name()
}

func (s *Source) GetItemsCount() int {
	return s.items.Len()
}

func (s *Source) GetCurrentItemsCount() int {
	return s.items.CLen()
}

func (s *Source) isMimeTypeSupported(key string) bool {
	mt, err := s.backend.GetMimeType(key)
	if err != nil {
		return false
	}

	return strings.HasPrefix(mt, "image/") || strings.HasPrefix(mt, "video/")
}

func (s *Source) SetEntries(tableDir string, tableName string, tableCreate bool, entries []string, recursive bool, randomize bool, include string, exclude string) (bool, error) {
	if s.items != nil {
		return false, errors.New("source: entries already set")
	}

	l := []string{}
	single := false
	loaded := false
	if tableName == "" || tableCreate {
		inc, err := regexp.Compile(include)
		if err != nil {
			return false, err
		}

		exc, err := regexp.Compile(exclude)
		if err != nil {
			return false, err
		}

		lr, si, err := s.backend.List(entries, recursive)
		if err != nil {
			return false, err
		}
		single = si

		sort.Strings(lr)

		for _, v := range lr {
			if toInclude(inc, v) && !toExclude(exc, v) && (s.backend.Remote() || s.isMimeTypeSupported(v)) {
				l = append(l, v)
			}
		}
		loaded = true
	}

	var err error
	s.items, err = dataset.New(tableDir, tableName, tableCreate, s.backend.Name(), l, randomize)
	if err != nil {
		return false, err
	}

	if s.items.Len() == 0 {
		return false, errors.New("source: failed to retrieve items")
	}

	if !loaded {
		if err := s.backend.SetItems(s.items.GetItems()); err != nil {
			return false, err
		}
	}

	return single, nil
}

func (s *Source) NextItem() (string, error) {
	if s.items == nil {
		return "", errors.New("source: items not set")
	}
	return s.items.Next()
}

func (s *Source) LookAheadItem() (string, bool, error) {
	if s.items == nil {
		return "", false, errors.New("source: items not set")
	}

	rv, err := s.items.LookAhead()
	if err != nil {
		if errors.Is(err, dataset.ErrLookAheadNotSupported) {
			return "", false, nil
		}
		return "", false, err
	}
	return rv, true, nil
}

func (s *Source) ForEachItem(f func(e string)) error {
	if s.items == nil {
		return errors.New("source: items not set")
	}
	return s.items.ForEach(f)
}

func (s *Source) GetFile(key string) (string, error) {
	return s.backend.GetFile(key)
}

func (s *Source) FormatItem(key string) (string, error) {
	return s.backend.FormatItem(key)
}
