package dataset

import (
	"crypto/rand"
	"errors"
	"math/big"
	"sync"
)

var (
	ErrEmpty           = errors.New("dataset: empty")
	ErrInvalidIndex    = errors.New("dataset: invalid index")
	ErrInvalidCallback = errors.New("dataset: invalid callback")
)

type DataSet struct {
	mtx       sync.RWMutex
	items     []string
	citems    []string
	randomize bool
}

func (d *DataSet) contains(v string) bool {
	for _, d := range d.items {
		if d == v {
			return true
		}
	}
	return false
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
	return rv
}

func (d *DataSet) Len() int {
	d.mtx.RLock()
	defer d.mtx.RUnlock()
	return len(d.items)
}

func (d *DataSet) next() (string, error) {
	if len(d.items) == 0 {
		return "", ErrEmpty
	}

	if len(d.citems) == 0 {
		d.citems = append(d.citems, d.items...)
	}

	idx := 0
	if d.randomize {
		bidx, err := rand.Int(rand.Reader, big.NewInt(int64(len(d.citems))))
		if err != nil {
			return "", err
		}

		idx = int(bidx.Int64())
	}

	v := d.citems[idx]
	d.citems = append(d.citems[:idx], d.citems[idx+1:]...)
	return v, nil
}

func (d *DataSet) Next() (string, error) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	return d.next()
}

func (d *DataSet) ForEach(f func(e string)) error {
	if f == nil {
		return ErrInvalidCallback
	}

	d.mtx.Lock()
	defer d.mtx.Unlock()

	if len(d.items) == 0 {
		return nil
	}

	for i := 0; i < len(d.items); i++ {
		next, err := d.next()
		if err != nil {
			return err
		}

		f(next)
	}

	return nil
}
