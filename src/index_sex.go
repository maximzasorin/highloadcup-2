package main

import (
	"sync"
)

type IndexSex struct {
	rwLock sync.RWMutex
	sex    map[byte]*IndexID
}

func NewIndexSex() *IndexSex {
	return &IndexSex{
		sex: make(map[byte]*IndexID),
	}
}

func (index *IndexSex) Add(s byte, id ID) {
	index.rwLock.RLock()
	_, ok := index.sex[s]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.sex[s] = NewIndexID(64)
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.sex[s].Add(id)
		index.rwLock.RUnlock()
		return
	}
	index.sex[s].Add(id)
	index.rwLock.RUnlock()
}

func (index *IndexSex) Append(s byte, id ID) {
	index.rwLock.RLock()
	_, ok := index.sex[s]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.sex[s] = NewIndexID(64)
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.sex[s].Append(id)
		index.rwLock.RUnlock()
		return
	}
	index.sex[s].Append(id)
	index.rwLock.RUnlock()
}

func (index *IndexSex) Update(s byte) {
	index.rwLock.Lock()
	_, ok := index.sex[s]
	if !ok {
		index.rwLock.Unlock()
		return
	}
	index.sex[s].Update()
	index.rwLock.Unlock()
}

func (index *IndexSex) UpdateAll() {
	index.rwLock.Lock()
	for s := range index.sex {
		index.sex[s].Update()
	}
	index.rwLock.Unlock()
}

func (index *IndexSex) Remove(s byte, id ID) {
	index.rwLock.RLock()
	_, ok := index.sex[s]
	if !ok {
		index.rwLock.RUnlock()
		return
	}
	index.sex[s].Remove(id)
	index.rwLock.RUnlock()
}

func (index *IndexSex) Find(s byte) IDS {
	index.rwLock.RLock()
	if _, ok := index.sex[s]; ok {
		ids := index.sex[s].FindAll()
		index.rwLock.RUnlock()
		return ids
	}
	index.rwLock.RUnlock()
	return make(IDS, 0)
}

func (index *IndexSex) Len() int {
	index.rwLock.RLock()
	sexLen := len(index.sex)
	index.rwLock.RUnlock()
	return sexLen
}
