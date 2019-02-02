package main

import (
	"sort"
	"sync"
)

type IndexLiker struct {
	rwLock sync.RWMutex
	likers map[ID]*IndexLikes
}

func NewIndexLiker() *IndexLiker {
	return &IndexLiker{
		likers: make(map[ID]*IndexLikes),
	}
}

func (index *IndexLiker) Add(liker ID, likee ID, ts uint32) {
	index.rwLock.RLock()
	_, ok := index.likers[liker]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.likers[liker] = NewIndexLikes(0)
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.likers[liker].Add(likee, ts)
		index.rwLock.RUnlock()
		return
	}
	index.likers[liker].Add(likee, ts)
	index.rwLock.RUnlock()
}

func (index *IndexLiker) Append(liker ID, likee ID, ts uint32) {
	index.rwLock.RLock()
	_, ok := index.likers[liker]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.likers[liker] = NewIndexLikes(0)
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.likers[liker].Append(likee, ts)
		index.rwLock.RUnlock()
		return
	}
	index.likers[liker].Append(likee, ts)
	index.rwLock.RUnlock()
}

func (index *IndexLiker) Find(liker ID) AccountLikes {
	index.rwLock.RLock()
	if _, ok := index.likers[liker]; ok {
		likes := index.likers[liker].FindAll()
		index.rwLock.RUnlock()
		return likes
	}
	index.rwLock.RUnlock()
	return make(AccountLikes, 0)
}

func (index *IndexLiker) UpdateAll() {
	index.rwLock.Lock()
	for liker := range index.likers {
		index.likers[liker].Update()
	}
	index.rwLock.Unlock()
}

// func (index *IndexLiker) Iter(liker ID) IndexIterator {
// 	index.rwLock.RLock()
// 	if _, ok := index.likers[liker]; ok {
// 		iter := index.likers[liker].Iter()
// 		index.rwLock.RUnlock()
// 		return iter
// 	}
// 	index.rwLock.RUnlock()
// 	return EmptyIndexIterator
// }

type IndexLikes struct {
	rwLock sync.RWMutex
	likes  AccountLikes
}

func NewIndexLikes(N int) *IndexLikes {
	return &IndexLikes{
		likes: make(AccountLikes, 0, N),
	}
}

func (index *IndexLikes) FindAll() AccountLikes {
	return index.likes
}

func (index *IndexLikes) Len() int {
	return len(index.likes)
}

func (index *IndexLikes) Add(likee ID, ts uint32) {
	index.rwLock.Lock()
	n := len(index.likes)
	i := sort.Search(n, func(i int) bool {
		return index.likes[i].ID <= likee
	})
	// if i < n && index.likes[i].ID == likee {
	// 	index.rwLock.Unlock()
	// 	return
	// }
	index.likes = append(index.likes, AccountLike{})
	copy(index.likes[i+1:], index.likes[i:])
	index.likes[i] = AccountLike{
		ID: likee,
		Ts: ts,
	}
	index.rwLock.Unlock()
}

func (index *IndexLikes) Append(id ID, ts uint32) {
	index.rwLock.Lock()
	index.likes = append(index.likes, AccountLike{
		ID: id,
		Ts: ts,
	})
	index.rwLock.Unlock()
}

func (index *IndexLikes) Update() {
	index.rwLock.Lock()
	sort.Sort(index.likes)
	index.rwLock.Unlock()
}

func (al AccountLikes) Len() int {
	return len(al)
}

func (al AccountLikes) Swap(i, j int) {
	al[i], al[j] = al[j], al[i]
}

func (al AccountLikes) Less(i, j int) bool {
	return al[i].ID > al[j].ID
}
