package source

import (
	"errors"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/rafaelmartins/b8r/internal/source/fp"
	"github.com/rafaelmartins/b8r/internal/source/local"
)

var errSkip = errors.New("source: skip")

type SourceBackend interface {
	Name() string
	PreFilterMimeType() bool
	SetParameter(key string, value interface{}) error
	List() ([]string, error)
	GetFile(key string) (string, error)
	GetMimeType(key string) (string, error)
	CompletionHandler(prev string, cur string) []string
}

var registry = []SourceBackend{
	&local.LocalSource{},
	&fp.FpSource{},
}

type Source struct {
	backend   SourceBackend
	items     []string
	randomize bool
	filter    *regexp2.Regexp
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

func New(name string, randomize bool, filter string) (*Source, error) {
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

	f, err := regexp2.Compile(filter, 0)
	if err != nil {
		return nil, err
	}

	return &Source{
		backend:   backend,
		randomize: randomize,
		filter:    f,
	}, nil
}

func (s *Source) SetParameter(key string, value interface{}) error {
	return s.backend.SetParameter(key, value)
}

func (s *Source) isMimeTypeSupported(key string) bool {
	mt, err := s.backend.GetMimeType(key)
	if err != nil {
		return false
	}

	return strings.HasPrefix(mt, "image/") || strings.HasPrefix(mt, "video/")
}

func (s *Source) List() ([]string, error) {
	lr, err := s.backend.List()
	if err != nil {
		return nil, err
	}

	sort.Strings(lr)

	l := []string{}
	if s.filter != nil {
		for _, v := range lr {
			if ok, err := s.filter.MatchString(v); err == nil && ok {
				if s.backend.PreFilterMimeType() && !s.isMimeTypeSupported(v) {
					continue
				}
				l = append(l, v)
			}
		}
	}

	if s.randomize {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(l), func(i int, j int) {
			l[i], l[j] = l[j], l[i]
		})
	}

	return l, nil
}

func (s *Source) pop() (string, error) {
	if s.items == nil || len(s.items) == 0 {
		items, err := s.List()
		if err != nil {
			return "", err
		}

		if len(items) == 0 {
			return "", errors.New("source: failed to retrieve items")
		}

		s.items = items
	}

	var pop string
	pop, s.items = s.items[0], s.items[1:]

	if !s.backend.PreFilterMimeType() && !s.isMimeTypeSupported(pop) {
		return "", errSkip
	}

	return pop, nil
}

func (s *Source) Pop() (string, error) {
	for {
		p, err := s.pop()
		if err == nil {
			return p, nil
		}

		if err != errSkip {
			return "", err
		}
	}
}

func (s *Source) GetFile(key string) (string, error) {
	return s.backend.GetFile(key)
}
