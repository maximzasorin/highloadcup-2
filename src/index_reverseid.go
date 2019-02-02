package main

import (
	"sort"
	"sync"
)

type IndexReverseID struct {
	ids    IDS
	rwLock sync.RWMutex
}

func NewIndexReverseID(N int) *IndexReverseID {
	return &IndexReverseID{
		ids: make(IDS, 0, N),
	}
}

func (index *IndexReverseID) FindAll() IDS {
	return index.ids
}

func (index *IndexReverseID) Iter() IndexIterator {
	return NewIndexReverseIDIterator(index.ids)
}

func (index *IndexReverseID) Len() int {
	return len(index.ids)
}

func (index *IndexReverseID) Add(id ID) {
	index.rwLock.Lock()
	n := len(index.ids)
	i := sort.Search(n, func(i int) bool {
		return index.ids[i] >= id
	})
	if i < n && index.ids[i] == id {
		index.rwLock.Unlock()
		return
	}
	index.ids = append(index.ids, 0)
	copy(index.ids[i+1:], index.ids[i:])
	index.ids[i] = id
	index.rwLock.Unlock()
}

func (index *IndexReverseID) Append(id ID) {
	index.rwLock.Lock()
	index.ids = append(index.ids, id)
	index.rwLock.Unlock()
}

func (index *IndexReverseID) Update() {
	index.rwLock.Lock()
	sort.Sort(sort.Reverse(index.ids))
	index.rwLock.Unlock()
}

func (index *IndexReverseID) Remove(id ID) {
	index.rwLock.Lock()
	n := len(index.ids)
	i := sort.Search(n, func(i int) bool {
		return index.ids[i] >= id
	})
	if i < n && index.ids[i] == id {
		index.ids = append(index.ids[:i], index.ids[i+1:]...)
	}
	index.rwLock.Unlock()
}

type IndexReverseIDIterator struct {
	ids   IDS
	index int
}

func NewIndexReverseIDIterator(ids IDS) *IndexReverseIDIterator {
	return &IndexReverseIDIterator{
		ids:   ids,
		index: len(ids) - 1,
	}
}

func (it *IndexReverseIDIterator) Cur() ID {
	if it.index >= 0 {
		return it.ids[it.index]
	}
	return 0
}

func (it *IndexReverseIDIterator) Next() ID {
	it.index--
	if it.index >= 0 {
		return it.ids[it.index]
	}
	return 0
}
