package main

import (
	"sync"
)

type IndexStatusSex struct {
	rwLock    sync.RWMutex
	statusSex map[byte]map[byte]*IndexID
}

func NewIndexStatusSex() *IndexStatusSex {
	return &IndexStatusSex{
		statusSex: make(map[byte]map[byte]*IndexID),
	}
}

func (index *IndexStatusSex) Add(status byte, sex byte, id ID) {
	index.rwLock.RLock()
	_, ok := index.statusSex[status]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.statusSex[status] = make(map[byte]*IndexID)
		index.rwLock.Unlock()
		index.rwLock.RLock()
	}
	_, ok = index.statusSex[status][sex]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.statusSex[status][sex] = NewIndexID(64)
		index.rwLock.Unlock()
		index.rwLock.RLock()
	}

	index.statusSex[status][sex].Add(id)
	index.rwLock.RUnlock()
}

func (index *IndexStatusSex) Append(status byte, sex byte, id ID) {
	index.rwLock.RLock()
	_, ok := index.statusSex[status]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.statusSex[status] = make(map[byte]*IndexID)
		index.rwLock.Unlock()
		index.rwLock.RLock()
	}
	_, ok = index.statusSex[status][sex]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.statusSex[status][sex] = NewIndexID(64)
		index.rwLock.Unlock()
		index.rwLock.RLock()
	}

	index.statusSex[status][sex].Append(id)
	index.rwLock.RUnlock()
}

func (index *IndexStatusSex) Update(s byte) {
	// index.rwLock.Lock()
	// _, ok := index.statusSex[status]
	// if !ok {
	// 	index.rwLock.Unlock()
	// 	return
	// }
	// index.sex[s].Update()
	// index.rwLock.Unlock()
}

func (index *IndexStatusSex) UpdateAll() {
	index.rwLock.Lock()
	for status := range index.statusSex {
		for sex := range index.statusSex[status] {
			index.statusSex[status][sex].Update()
		}
	}
	index.rwLock.Unlock()
}

func (index *IndexStatusSex) Remove(status byte, sex byte, id ID) {
	index.rwLock.RLock()
	_, ok := index.statusSex[status]
	if !ok {
		index.rwLock.RUnlock()
		return
	}
	_, ok = index.statusSex[status][sex]
	if !ok {
		index.rwLock.RUnlock()
		return
	}
	index.statusSex[status][sex].Remove(id)
}

func (index *IndexStatusSex) Find(status byte, sex byte) IDS {
	index.rwLock.RLock()
	if _, ok := index.statusSex[status]; ok {
		if _, ok := index.statusSex[status][sex]; ok {
			ids := index.statusSex[status][sex].FindAll()
			index.rwLock.RUnlock()
			return ids
		}
	}
	index.rwLock.RUnlock()
	return make(IDS, 0)
}

func (index *IndexStatusSex) Len() int {
	return -1
	// index.rwLock.RLock()
	// sexLen := len(index.sex)
	// index.rwLock.RUnlock()
	// return sexLen
}
