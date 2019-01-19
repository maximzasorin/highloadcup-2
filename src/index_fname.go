package main

import (
	"sync"
)

type IndexFname struct {
	rwLock sync.RWMutex
	fnames map[Fname]*IndexID
}

func NewIndexFname() *IndexFname {
	return &IndexFname{
		fnames: make(map[Fname]*IndexID),
	}
}

func (index *IndexFname) Add(fname Fname, id ID) {
	index.rwLock.RLock()
	_, ok := index.fnames[fname]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.fnames[fname] = NewIndexID(64)
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.fnames[fname].Add(id)
		index.rwLock.RUnlock()
		return
	}
	index.fnames[fname].Add(id)
	index.rwLock.RUnlock()
}

func (index *IndexFname) Append(fname Fname, id ID) {
	index.rwLock.RLock()
	_, ok := index.fnames[fname]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.fnames[fname] = NewIndexID(64)
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.fnames[fname].Append(id)
		index.rwLock.RUnlock()
		return
	}
	index.fnames[fname].Append(id)
	index.rwLock.RUnlock()
}

func (index *IndexFname) Update(fname Fname) {
	index.rwLock.Lock()
	_, ok := index.fnames[fname]
	if !ok {
		index.rwLock.Unlock()
		return
	}
	index.fnames[fname].Update()
	index.rwLock.Unlock()
}

func (index *IndexFname) UpdateAll() {
	index.rwLock.Lock()
	for fname := range index.fnames {
		index.fnames[fname].Update()
	}
	index.rwLock.Unlock()
}

func (index *IndexFname) Remove(fname Fname, id ID) {
	index.rwLock.RLock()
	_, ok := index.fnames[fname]
	if !ok {
		index.rwLock.RUnlock()
		return
	}
	index.fnames[fname].Remove(id)
	index.rwLock.RUnlock()
}

func (index *IndexFname) Find(fname Fname) IDS {
	index.rwLock.RLock()
	if _, ok := index.fnames[fname]; ok {
		ids := index.fnames[fname].FindAll()
		index.rwLock.RUnlock()
		return ids
	}
	index.rwLock.RUnlock()
	return make(IDS, 0)
}

func (index *IndexFname) Len() int {
	index.rwLock.RLock()
	fnamesLen := len(index.fnames)
	index.rwLock.RUnlock()
	return fnamesLen
}
