package main

import (
	"sync"
)

type IndexLikee struct {
	rwLock sync.RWMutex
	likees map[ID]*IndexID
}

func NewIndexLikee() *IndexLikee {
	return &IndexLikee{
		likees: make(map[ID]*IndexID),
	}
}

func (index *IndexLikee) Add(likee ID, liker ID) {
	index.rwLock.RLock()
	_, ok := index.likees[likee]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.likees[likee] = NewIndexID(64)
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.likees[likee].Add(liker)
		index.rwLock.RUnlock()
		return
	}
	index.likees[likee].Add(liker)
	index.rwLock.RUnlock()
}

func (index *IndexLikee) Remove(likee ID, liker ID) {
	index.rwLock.RLock()
	_, ok := index.likees[likee]
	if !ok {
		index.rwLock.RUnlock()
		return
	}
	index.likees[likee].Remove(liker)
	index.rwLock.RUnlock()
}

func (index *IndexLikee) Find(likee ID) IDS {
	index.rwLock.RLock()
	if _, ok := index.likees[likee]; ok {
		ids := index.likees[likee].FindAll()
		index.rwLock.RUnlock()
		return ids
	}
	index.rwLock.RUnlock()
	return make(IDS, 0)
}
