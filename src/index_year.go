package main

import (
	"sync"
)

type IndexYear struct {
	rwLock sync.RWMutex
	years  map[Year]*IndexID
}

func NewIndexYear() *IndexYear {
	return &IndexYear{
		years: make(map[Year]*IndexID),
	}
}

func (index *IndexYear) Add(year Year, id ID) {
	index.rwLock.RLock()
	_, ok := index.years[year]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.years[year] = NewIndexID(64)
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.years[year].Add(id)
		index.rwLock.RUnlock()
		return
	}
	index.years[year].Add(id)
	index.rwLock.RUnlock()
}

func (index *IndexYear) Append(year Year, id ID) {
	index.rwLock.RLock()
	_, ok := index.years[year]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.years[year] = NewIndexID(64)
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.years[year].Append(id)
		index.rwLock.RUnlock()
		return
	}
	index.years[year].Append(id)
	index.rwLock.RUnlock()
}

func (index *IndexYear) Update(year Year) {
	index.rwLock.Lock()
	_, ok := index.years[year]
	if !ok {
		index.rwLock.Unlock()
		return
	}
	index.years[year].Update()
	index.rwLock.Unlock()
}

func (index *IndexYear) UpdateAll() {
	index.rwLock.Lock()
	for year := range index.years {
		index.years[year].Update()
	}
	index.rwLock.Unlock()
}

func (index *IndexYear) Remove(year Year, id ID) {
	index.rwLock.RLock()
	_, ok := index.years[year]
	if !ok {
		index.rwLock.RUnlock()
		return
	}
	index.years[year].Remove(id)
	index.rwLock.RUnlock()
}

func (index *IndexYear) Find(year Year) IDS {
	index.rwLock.RLock()
	if _, ok := index.years[year]; ok {
		ids := index.years[year].FindAll()
		index.rwLock.RUnlock()
		return ids
	}
	index.rwLock.RUnlock()
	return make(IDS, 0)
}
