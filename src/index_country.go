package main

import (
	"sync"
)

type IndexCountry struct {
	rwLock    sync.RWMutex
	countries map[Country]*IndexID
}

func NewIndexCountry() *IndexCountry {
	return &IndexCountry{
		countries: make(map[Country]*IndexID),
	}
}

func (index *IndexCountry) Add(country Country, id ID) {
	index.rwLock.RLock()
	_, ok := index.countries[country]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.countries[country] = NewIndexID(64)
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.countries[country].Add(id)
		index.rwLock.RUnlock()
		return
	}
	index.countries[country].Add(id)
	index.rwLock.RUnlock()
}

func (index *IndexCountry) Append(country Country, id ID) {
	index.rwLock.RLock()
	_, ok := index.countries[country]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.countries[country] = NewIndexID(64)
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.countries[country].Append(id)
		index.rwLock.RUnlock()
		return
	}
	index.countries[country].Append(id)
	index.rwLock.RUnlock()
}

func (index *IndexCountry) Update(country Country) {
	index.rwLock.Lock()
	_, ok := index.countries[country]
	if !ok {
		index.rwLock.Unlock()
		return
	}
	index.countries[country].Update()
	index.rwLock.Unlock()
}

func (index *IndexCountry) UpdateAll() {
	index.rwLock.Lock()
	for country := range index.countries {
		index.countries[country].Update()
	}
	index.rwLock.Unlock()
}

func (index *IndexCountry) Remove(country Country, id ID) {
	index.rwLock.RLock()
	_, ok := index.countries[country]
	if !ok {
		index.rwLock.RUnlock()
		return
	}
	index.countries[country].Remove(id)
	index.rwLock.RUnlock()
}

func (index *IndexCountry) Find(country Country) IDS {
	index.rwLock.RLock()
	if _, ok := index.countries[country]; ok {
		ids := index.countries[country].FindAll()
		index.rwLock.RUnlock()
		return ids
	}
	index.rwLock.RUnlock()
	return make(IDS, 0)
}

func (index *IndexCountry) Count(country Country) (count int) {
	index.rwLock.RLock()
	if _, ok := index.countries[country]; ok {
		count = len(index.countries[country].FindAll())
	}
	index.rwLock.RUnlock()
	return count
}

func (index *IndexCountry) Len() int {
	index.rwLock.RLock()
	countriesLen := len(index.countries)
	index.rwLock.RUnlock()
	return countriesLen
}
