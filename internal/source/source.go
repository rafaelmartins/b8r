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
}

var registry = []SourceBackend{
	&local.LocalSource{},
	&fp.FpSource{},
}

type Source struct {
	backend SourceBackend
	entries *dataset.DataSet
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

func (s *Source) isMimeTypeSupported(key string) bool {
	mt, err := s.backend.GetMimeType(key)
	if err != nil {
		return false
	}

	return strings.HasPrefix(mt, "image/") || strings.HasPrefix(mt, "video/")
}

func (s *Source) SetEntries(entries []string, recursive bool, randomize bool, include string, exclude string) (bool, error) {
	if s.entries != nil {
		return false, errors.New("source: entries already set")
	}

	inc, err := regexp.Compile(include)
	if err != nil {
		return false, err
	}

	exc, err := regexp.Compile(exclude)
	if err != nil {
		return false, err
	}

	lr, single, err := s.backend.List(entries, recursive)
	if err != nil {
		return false, err
	}

	sort.Strings(lr)

	l := []string{}

	for _, v := range lr {
		if toInclude(inc, v) && !toExclude(exc, v) && (s.backend.Remote() || s.isMimeTypeSupported(v)) {
			l = append(l, v)
		}
	}

	s.entries = dataset.New(l, randomize)
	if s.entries.Len() == 0 {
		return false, errors.New("source: failed to retrieve entries")
	}
	return single, nil
}

func (s *Source) NextEntry() (string, error) {
	if s.entries != nil {
		return s.entries.Next()
	}
	return "", errors.New("source: entries not set")
}

func (s *Source) ForEachEntry(f func(e string)) error {
	if s.entries != nil {
		return s.entries.ForEach(f)
	}
	return errors.New("source: entries not set")
}

func (s *Source) GetFile(key string) (string, error) {
	return s.backend.GetFile(key)
}
