package dataset

import "sync"

type DataSet struct {
	mtx   sync.RWMutex
	items []string
}

func New() *DataSet {
	return &DataSet{
		items: []string{},
	}
}

func (d *DataSet) Contains(v string) bool {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	for _, d := range d.items {
		if d == v {
			return true
		}
	}
	return false
}

func (d *DataSet) Add(v string) {
	if d.Contains(v) {
		return
	}

	d.mtx.Lock()
	defer d.mtx.Unlock()
	d.items = append(d.items, v)
}

func (d *DataSet) Slice() []string {
	d.mtx.RLock()
	defer d.mtx.RUnlock()
	return append([]string{}, d.items...)
}
