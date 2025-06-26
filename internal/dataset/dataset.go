package dataset

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/jameycribbs/hare"
	"github.com/jameycribbs/hare/datastores/disk"
	"github.com/jameycribbs/hare/datastores/ram"
)

var (
	ErrEmpty                 = errors.New("dataset: empty")
	ErrInvalidIndex          = errors.New("dataset: invalid index")
	ErrInvalidCallback       = errors.New("dataset: invalid callback")
	ErrLookAheadNotSupported = errors.New("dataset: lookahead not supported")
)

type entry struct {
	ID    int    `json:"id"`
	Entry string `json:"entry"`
}

func (e *entry) SetID(id int) {
	e.ID = id
}

func (e *entry) GetID() int {
	return e.ID
}

func (e *entry) AfterFind(*hare.Database) error {
	*e = entry(*e)
	return nil
}

type metadata struct {
	Source    string   `json:"source"`
	Items     []string `json:"items"`
	Randomize bool     `json:"randomize"`
}

func ListTables(tableDir string) []string {
	l, err := os.ReadDir(filepath.Join(tableDir, "meta"))
	if err != nil {
		return nil
	}

	rv := []string{}
	for _, e := range l {
		if strings.HasSuffix(e.Name(), ".json") {
			rv = append(rv, strings.TrimSuffix(e.Name(), ".json"))
		}
	}
	return rv
}

func TableExists(tableDir string, table string) bool {
	return slices.Contains(ListTables(tableDir), table)
}

func TableSource(tableDir string, table string) (string, error) {
	fp, err := os.Open(filepath.Join(tableDir, "meta", table+".json"))
	if err != nil {
		return "", err
	}
	defer fp.Close()

	meta := &metadata{}
	if err := json.NewDecoder(fp).Decode(meta); err != nil {
		return "", err
	}
	return meta.Source, nil
}

type DataSet struct {
	mtx           sync.RWMutex
	db            *hare.Database
	source        string
	table         string
	next          string
	items         []string
	randomize     bool
	withLookahead bool
}

func New(tableDir string, tableName string, tableCreate bool, source string, items []string, randomize bool) (*DataSet, error) {
	rv := &DataSet{
		source:    source,
		table:     tableName,
		randomize: randomize,
	}
	for _, item := range items {
		if !slices.Contains(rv.items, item) {
			rv.items = append(rv.items, item)
		}
	}

	if tableName == "" {
		ds, err := ram.New(nil)
		if err != nil {
			return nil, err
		}

		rv.db, err = hare.New(ds)
		if err != nil {
			return nil, err
		}

		rv.withLookahead = true

		if err := rv.refill(); err != nil {
			return nil, err
		}
		return rv, nil
	}

	dir := filepath.Join(tableDir, "data")
	if err := os.MkdirAll(dir, 0777); err != nil {
		return nil, err
	}

	ds, err := disk.New(dir, ".json")
	if err != nil {
		return nil, err
	}

	rv.db, err = hare.New(ds)
	if err != nil {
		return nil, err
	}

	dir = filepath.Join(tableDir, "meta")
	if err := os.MkdirAll(dir, 0777); err != nil {
		rv.db.Close()
		return nil, err
	}

	if tableCreate {
		fp, err := os.Create(filepath.Join(dir, tableName+".json"))
		if err != nil {
			rv.db.Close()
			return nil, err
		}
		defer fp.Close()

		meta := &metadata{
			Source:    rv.source,
			Items:     rv.items,
			Randomize: rv.randomize,
		}
		if err := json.NewEncoder(fp).Encode(meta); err != nil {
			rv.db.Close()
			return nil, err
		}

		if err := rv.refill(); err != nil {
			return nil, err
		}
		return rv, nil
	}

	fp, err := os.Open(filepath.Join(dir, tableName+".json"))
	if err != nil {
		rv.db.Close()
		return nil, err
	}
	defer fp.Close()

	meta := &metadata{}
	if err := json.NewDecoder(fp).Decode(meta); err != nil {
		rv.db.Close()
		return nil, err
	}

	rv.source = meta.Source
	rv.items = meta.Items
	rv.randomize = meta.Randomize
	return rv, nil
}

func (d *DataSet) Close() error {
	return d.db.Close()
}

func (d *DataSet) len() int {
	return len(d.items)
}

func (d *DataSet) Len() int {
	d.mtx.RLock()
	defer d.mtx.RUnlock()
	return d.len()
}

func (d *DataSet) clen() int {
	tmp, err := d.db.IDs(d.table)
	if err != nil {
		return 0
	}
	return len(tmp)
}

func (d *DataSet) CLen() int {
	d.mtx.RLock()
	defer d.mtx.RUnlock()
	return d.clen()
}

func (d *DataSet) refill() error {
	if d.db.TableExists(d.table) {
		if err := d.db.DropTable(d.table); err != nil {
			return err
		}
	}
	if err := d.db.CreateTable(d.table); err != nil {
		return err
	}
	for _, v := range d.items {
		if _, err := d.db.Insert(d.table, &entry{Entry: v}); err != nil {
			return err
		}
	}
	return nil
}

func (d *DataSet) getIDs() ([]int, error) {
	if d.len() == 0 {
		return nil, ErrEmpty
	}
	if d.clen() == 0 {
		if err := d.refill(); err != nil {
			return nil, err
		}
	}

	tmp, err := d.db.IDs(d.table)
	if err != nil {
		return nil, err
	}
	slices.Sort(tmp)
	return tmp, nil
}

func (d *DataSet) pick() (string, error) {
	tmp, err := d.getIDs()
	if err != nil {
		return "", err
	}

	idx := 0
	if d.randomize {
		bidx, err := rand.Int(rand.Reader, big.NewInt(int64(len(tmp))))
		if err != nil {
			return "", err
		}
		idx = int(bidx.Int64())
	}

	v := entry{}
	if err := d.db.Find(d.table, tmp[idx], &v); err != nil {
		return "", err
	}

	if err := d.db.Delete(d.table, v.ID); err != nil {
		return "", err
	}

	return v.Entry, nil
}

func (d *DataSet) Next() (string, error) {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	if len(d.items) == 0 {
		return "", ErrEmpty
	}

	if d.next != "" {
		rv := d.next
		d.next = ""
		return rv, nil
	}

	return d.pick()
}

func (d *DataSet) LookAhead() (string, error) {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	if len(d.items) == 0 {
		return "", ErrEmpty
	}

	if !d.withLookahead {
		return "", ErrLookAheadNotSupported
	}

	if d.next != "" {
		return d.next, nil
	}

	rv, err := d.pick()
	if err != nil {
		return "", nil
	}
	d.next = rv
	return rv, nil
}

func (d *DataSet) GetItems() []string {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	return slices.Clone(d.items)
}

func (d *DataSet) ForEach(f func(e string)) error {
	if f == nil {
		return ErrInvalidCallback
	}

	d.mtx.Lock()
	defer d.mtx.Unlock()

	tmp, err := d.getIDs()
	if err != nil {
		return err
	}

	if len(tmp) == 0 {
		return nil
	}

	for i := range len(tmp) {
		id := 0
		if d.randomize {
			bidx, err := rand.Int(rand.Reader, big.NewInt(int64(len(tmp))))
			if err != nil {
				return err
			}

			idx := int(bidx.Int64())
			id = tmp[idx]
			tmp = append(tmp[:idx], tmp[idx+1:]...)
		} else {
			id = tmp[i]
		}

		v := entry{}
		if err := d.db.Find(d.table, id, &v); err != nil {
			return err
		}

		f(v.Entry)
	}
	return nil
}
