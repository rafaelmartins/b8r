package source

import (
	"errors"
	"math/rand"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/rafaelmartins/b8r/internal/source/fp"
	"github.com/rafaelmartins/b8r/internal/source/local"
)

var errSkip = errors.New("source: skip")

type SourceBackend interface {
	Name() string
	PreFilterMimeType() bool
	IsSingleItem() bool
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
	include   *regexp.Regexp
	exclude   *regexp.Regexp
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

func New(name string, randomize bool, include string, exclude string) (*Source, error) {
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

	i, err := regexp.Compile(include)
	if err != nil {
		return nil, err
	}

	e, err := regexp.Compile(exclude)
	if err != nil {
		return nil, err
	}

	return &Source{
		backend:   backend,
		randomize: randomize,
		include:   i,
		exclude:   e,
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

func (s *Source) toInclude(e string) bool {
	if s.include == nil {
		return true
	}
	return s.include.MatchString(e)
}

func (s *Source) toExclude(e string) bool {
	if s.exclude == nil {
		return false
	}
	return s.exclude.MatchString(e)
}

func (s *Source) IsSingleItem() bool {
	return s.backend.IsSingleItem()
}

func (s *Source) List() ([]string, error) {
	lr, err := s.backend.List()
	if err != nil {
		return nil, err
	}

	sort.Strings(lr)

	l := []string{}

	if s.backend.IsSingleItem() {
		l = append(l, lr...)
	} else {
		for _, v := range lr {
			if s.toInclude(v) && !s.toExclude(v) && !(s.backend.PreFilterMimeType() && !s.isMimeTypeSupported(v)) {
				l = append(l, v)
			}
		}

		if s.randomize {
			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(len(l), func(i int, j int) {
				l[i], l[j] = l[j], l[i]
			})
		}
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
