package main

import (
	"sync"
)

type IndexStatus struct {
	rwLock   sync.RWMutex
	statuses map[byte]*IndexID
}

func NewIndexStatus() *IndexStatus {
	return &IndexStatus{
		statuses: make(map[byte]*IndexID),
	}
}

func (index *IndexStatus) Add(s byte, id ID) {
	index.rwLock.RLock()
	_, ok := index.statuses[s]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.statuses[s] = NewIndexID(64)
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.statuses[s].Add(id)
		index.rwLock.RUnlock()
		return
	}
	index.statuses[s].Add(id)
	index.rwLock.RUnlock()
}

func (index *IndexStatus) Append(s byte, id ID) {
	index.rwLock.RLock()
	_, ok := index.statuses[s]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.statuses[s] = NewIndexID(64)
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.statuses[s].Append(id)
		index.rwLock.RUnlock()
		return
	}
	index.statuses[s].Append(id)
	index.rwLock.RUnlock()
}

func (index *IndexStatus) Update(s byte) {
	index.rwLock.Lock()
	_, ok := index.statuses[s]
	if !ok {
		index.rwLock.Unlock()
		return
	}
	index.statuses[s].Update()
	index.rwLock.Unlock()
}

func (index *IndexStatus) UpdateAll() {
	index.rwLock.Lock()
	for s := range index.statuses {
		index.statuses[s].Update()
	}
	index.rwLock.Unlock()
}

func (index *IndexStatus) Remove(s byte, id ID) {
	index.rwLock.RLock()
	_, ok := index.statuses[s]
	if !ok {
		index.rwLock.RUnlock()
		return
	}
	index.statuses[s].Remove(id)
	index.rwLock.RUnlock()
}

func (index *IndexStatus) Find(s byte) IDS {
	index.rwLock.RLock()
	if _, ok := index.statuses[s]; ok {
		ids := index.statuses[s].FindAll()
		index.rwLock.RUnlock()
		return ids
	}
	index.rwLock.RUnlock()
	return make(IDS, 0)
}

func (index *IndexStatus) Len() int {
	index.rwLock.RLock()
	statusesLen := len(index.statuses)
	index.rwLock.RUnlock()
	return statusesLen
}
