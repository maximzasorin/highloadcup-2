package main

import (
	"sort"
	"sync"
)

type ID uint32

type IDS []ID

type IndexID struct {
	ids    IDS
	rwLock sync.RWMutex
}

func NewIndexID(N int) *IndexID {
	return &IndexID{
		ids: make(IDS, 0, N),
	}
}

func (index *IndexID) FindAll() IDS {
	return index.ids
}

func (index *IndexID) Add(id ID) {
	index.rwLock.Lock()
	n := len(index.ids)
	i := sort.Search(n, func(i int) bool {
		return index.ids[i] <= id
	})
	if i < n && index.ids[i] == id {
		index.rwLock.Unlock()
		return
	}
	// index.ids = append(index.ids[:i], append(IDS{id}, index.ids[i:]...)...)
	index.ids = append(index.ids, 0 /* use the zero value of the element type */)
	copy(index.ids[i+1:], index.ids[i:])
	index.ids[i] = id
	index.rwLock.Unlock()
}

func (index *IndexID) Append(id ID) {
	index.rwLock.Lock()
	index.ids = append(index.ids, id)
	index.rwLock.Unlock()
}

func (index *IndexID) Update() {
	index.rwLock.Lock()
	sort.Sort(index.ids)
	index.rwLock.Unlock()
}

func (index *IndexID) Remove(id ID) {
	index.rwLock.Lock()
	n := len(index.ids)
	i := sort.Search(n, func(i int) bool {
		return index.ids[i] <= id
	})
	if i < n && index.ids[i] == id {
		index.ids = append(index.ids[:i], index.ids[i+1:]...)
	}
	index.rwLock.Unlock()
}

func (ids IDS) Len() int           { return len(ids) }
func (ids IDS) Swap(i, j int)      { ids[i], ids[j] = ids[j], ids[i] }
func (ids IDS) Less(i, j int) bool { return ids[i] > ids[j] }
