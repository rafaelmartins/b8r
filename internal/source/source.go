package source

import (
	"errors"
	"math/rand"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/rafaelmartins/b8r/internal/source/folder"
	"github.com/rafaelmartins/b8r/internal/source/fp"
)

var errSkip = errors.New("source: skip")

type SourceBackend interface {
	Name() string
	SetParameter(key string, value string) error
	List() ([]string, error)
	GetFile(key string) (string, error)
	GetMimeType(key string) (string, error)
}

var registry = []SourceBackend{
	&folder.FolderSource{},
	&fp.FpSource{},
}

type Source struct {
	backend   SourceBackend
	items     []string
	randomize bool
	filter    *regexp.Regexp
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

	f, err := regexp.Compile(filter)
	if err != nil {
		return nil, err
	}

	return &Source{
		backend:   backend,
		randomize: randomize,
		filter:    f,
	}, nil
}

func (s *Source) SetParameter(key string, value string) error {
	return s.backend.SetParameter(key, value)
}

func (s *Source) list() ([]string, error) {
	lr, err := s.backend.List()
	if err != nil {
		return nil, err
	}

	sort.Strings(lr)

	l := []string{}
	if s.filter != nil {
		for _, v := range lr {
			if s.filter.MatchString(v) {
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
		items, err := s.list()
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
