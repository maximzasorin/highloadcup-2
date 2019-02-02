package main

import (
	"sync"
)

type IndexInterest struct {
	rwLock    sync.RWMutex
	interests map[Interest]*IndexID
}

func NewIndexInterest() *IndexInterest {
	return &IndexInterest{
		interests: make(map[Interest]*IndexID),
	}
}

func (index *IndexInterest) Add(interest Interest, id ID) {
	index.rwLock.RLock()
	_, ok := index.interests[interest]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.interests[interest] = NewIndexID(64)
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.interests[interest].Add(id)
		index.rwLock.RUnlock()
		return
	}
	index.interests[interest].Add(id)
	index.rwLock.RUnlock()
}

func (index *IndexInterest) Append(interest Interest, id ID) {
	index.rwLock.RLock()
	_, ok := index.interests[interest]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.interests[interest] = NewIndexID(64)
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.interests[interest].Append(id)
		index.rwLock.RUnlock()
		return
	}
	index.interests[interest].Append(id)
	index.rwLock.RUnlock()
}

func (index *IndexInterest) Update(interest Interest) {
	index.rwLock.Lock()
	_, ok := index.interests[interest]
	if !ok {
		index.rwLock.Unlock()
		return
	}
	index.interests[interest].Update()
	index.rwLock.Unlock()
}

func (index *IndexInterest) UpdateAll() {
	index.rwLock.Lock()
	for interest := range index.interests {
		index.interests[interest].Update()
	}
	index.rwLock.Unlock()
}

func (index *IndexInterest) Remove(interest Interest, id ID) {
	index.rwLock.RLock()
	_, ok := index.interests[interest]
	if !ok {
		index.rwLock.RUnlock()
		return
	}
	index.interests[interest].Remove(id)
	index.rwLock.RUnlock()
}

func (index *IndexInterest) Find(interest Interest) IDS {
	index.rwLock.RLock()
	if _, ok := index.interests[interest]; ok {
		ids := index.interests[interest].FindAll()
		index.rwLock.RUnlock()
		return ids
	}
	index.rwLock.RUnlock()
	return make(IDS, 0)
}

func (index *IndexInterest) Iter(interest Interest) IndexIterator {
	index.rwLock.RLock()
	if _, ok := index.interests[interest]; ok {
		iter := index.interests[interest].Iter()
		index.rwLock.RUnlock()
		return iter
	}
	index.rwLock.RUnlock()
	return EmptyIndexIterator
}

func (index *IndexInterest) Len() int {
	index.rwLock.RLock()
	interestsLen := len(index.interests)
	index.rwLock.RUnlock()
	return interestsLen
}
