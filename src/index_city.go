package main

import (
	"sync"
)

type IndexCity struct {
	rwLock sync.RWMutex
	cities map[City]*IndexID
}

func NewIndexCity() *IndexCity {
	return &IndexCity{
		cities: make(map[City]*IndexID),
	}
}

func (index *IndexCity) Add(city City, id ID) {
	index.rwLock.RLock()
	_, ok := index.cities[city]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.cities[city] = NewIndexID(50)
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.cities[city].Add(id)
		index.rwLock.RUnlock()
		return
	}
	index.cities[city].Add(id)
	index.rwLock.RUnlock()
}

func (index *IndexCity) Append(city City, id ID) {
	index.rwLock.RLock()
	_, ok := index.cities[city]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.cities[city] = NewIndexID(50)
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.cities[city].Append(id)
		index.rwLock.RUnlock()
		return
	}
	index.cities[city].Append(id)
	index.rwLock.RUnlock()
}

func (index *IndexCity) Update(city City) {
	index.rwLock.Lock()
	_, ok := index.cities[city]
	if !ok {
		index.rwLock.Unlock()
		return
	}
	index.cities[city].Update()
	index.rwLock.Unlock()
}

func (index *IndexCity) UpdateAll() {
	index.rwLock.Lock()
	for city := range index.cities {
		index.cities[city].Update()
	}
	index.rwLock.Unlock()
}

func (index *IndexCity) Remove(city City, id ID) {
	index.rwLock.RLock()
	_, ok := index.cities[city]
	if !ok {
		index.rwLock.RUnlock()
		return
	}
	index.cities[city].Remove(id)
	index.rwLock.RUnlock()
}

func (index *IndexCity) Find(city City) IDS {
	index.rwLock.RLock()
	if _, ok := index.cities[city]; ok {
		ids := index.cities[city].FindAll()
		index.rwLock.RUnlock()
		return ids
	}
	index.rwLock.RUnlock()
	return make(IDS, 0)
}

func (index *IndexCity) Len() int {
	index.rwLock.RLock()
	citiesLen := len(index.cities)
	index.rwLock.RUnlock()
	return citiesLen
}
