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
	SetParameter(key string, value interface{}) error
	List() ([]string, error)
	GetFile(key string) (string, error)
	GetMimeType(key string) (string, error)
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
	found     bool
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
	if s.items != nil && len(s.items) == 0 && !s.found {
		return "", errors.New("source: no media found")
	}

	if s.items == nil || len(s.items) == 0 {
		s.found = false
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

	mt, err := s.backend.GetMimeType(pop)
	if err != nil {
		return "", errSkip
	}
	if !strings.HasPrefix(mt, "image/") && !strings.HasPrefix(mt, "video/") {
		return "", errSkip
	}

	s.found = true
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
