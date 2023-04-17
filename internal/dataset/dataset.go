package dataset

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

var (
	ErrEmpty           = errors.New("dataset: empty")
	ErrInvalidIndex    = errors.New("dataset: invalid index")
	ErrInvalidCallback = errors.New("dataset: invalid callback")
)

type DataSet struct {
	mtx       sync.RWMutex
	items     []string
	randomize bool
	idx       int
}

func (d *DataSet) contains(v string) bool {
	for _, d := range d.items {
		if d == v {
			return true
		}
	}
	return false
}

func (d *DataSet) rand() {
	if d.randomize && len(d.items) > 0 {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(d.items), func(i int, j int) {
			d.items[i], d.items[j] = d.items[j], d.items[i]
		})
	}
}

func New(items []string, randomize bool) *DataSet {
	rv := &DataSet{
		randomize: randomize,
	}
	for _, item := range items {
		if !rv.contains(item) {
			rv.items = append(rv.items, item)
		}
	}
	rv.rand()
	return rv
}

func (d *DataSet) Len() int {
	d.mtx.RLock()
	defer d.mtx.RUnlock()
	return len(d.items)
}

func (d *DataSet) Next() (string, error) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	if len(d.items) == 0 {
		return "", ErrEmpty
	}
	v := d.items[d.idx]
	d.idx++
	if d.idx >= len(d.items) {
		d.idx = 0
		d.rand()
	}
	return v, nil
}

func (d *DataSet) ForEach(f func(e string)) error {
	if f == nil {
		return ErrInvalidCallback
	}

	d.mtx.RLock()
	defer d.mtx.RUnlock()

	for _, item := range d.items {
		f(item)
	}
	return nil
}
