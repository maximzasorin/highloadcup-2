package main

import (
	"sort"
	"sync"
)

type ID uint32

type IDS []ID

type IndexID struct {
	ids    IDS
	rwLock sync.RWMutex
}

func NewIndexID(N int) *IndexID {
	return &IndexID{
		ids: make(IDS, 0, N),
	}
}

func (index *IndexID) FindAll() IDS {
	return index.ids
}

func (index *IndexID) Iter() IndexIterator {
	return NewIndexIDIterator(index.ids)
}

func (index *IndexID) Len() int {
	return len(index.ids)
}

func (index *IndexID) Add(id ID) {
	index.rwLock.Lock()
	n := len(index.ids)
	i := sort.Search(n, func(i int) bool {
		return index.ids[i] <= id
	})
	if i < n && index.ids[i] == id {
		index.rwLock.Unlock()
		return
	}
	index.ids = append(index.ids, 0)
	copy(index.ids[i+1:], index.ids[i:])
	index.ids[i] = id
	index.rwLock.Unlock()
}

func (index *IndexID) Append(id ID) {
	index.rwLock.Lock()
	index.ids = append(index.ids, id)
	index.rwLock.Unlock()
}

func (index *IndexID) Update() {
	index.rwLock.Lock()
	sort.Sort(index.ids)
	index.rwLock.Unlock()
}

func (index *IndexID) Remove(id ID) {
	index.rwLock.Lock()
	n := len(index.ids)
	i := sort.Search(n, func(i int) bool {
		return index.ids[i] <= id
	})
	if i < n && index.ids[i] == id {
		index.ids = append(index.ids[:i], index.ids[i+1:]...)
	}
	index.rwLock.Unlock()
}

func (ids IDS) Len() int           { return len(ids) }
func (ids IDS) Swap(i, j int)      { ids[i], ids[j] = ids[j], ids[i] }
func (ids IDS) Less(i, j int) bool { return ids[i] > ids[j] }

func IntersectIndexes(indexes ...IDS) IDS {
	minIndex := -1
	for i, index := range indexes {
		if minIndex == -1 || len(index) < minIndex {
			minIndex = i
		}
	}

	ids := make(IDS, 0)

	for _, ID := range indexes[minIndex] {
		exists := true
		for i := 0; i < len(indexes); i++ {
			if i == minIndex {
				continue
			}
			curIndex := sort.Search(len(indexes[i]), func(j int) bool {
				return indexes[i][j] <= ID
			})
			if curIndex == len(indexes[i]) || indexes[i][curIndex] != ID {
				exists = false
				break
			}
		}
		if exists {
			ids = append(ids, ID)
		}
	}

	return ids
}

func UnionIndexes(indexes ...IDS) IDS {
	cur := make([]uint32, len(indexes))
	resLen := 0
	for _, index := range indexes {
		resLen += len(index)
	}

	ids := make(IDS, 0, resLen)
	for {
		maxID := ID(0)
		// maxIndex := -1
		for i, curIndex := range cur {
			if curIndex < uint32(len(indexes[i])) && indexes[i][curIndex] > maxID {
				maxID = indexes[i][curIndex]
				// maxIndex = i
			}
		}

		if maxID > 0 {
			ids = append(ids, maxID)
			for i := range cur {
				if cur[i] < uint32(len(indexes[i])) && indexes[i][cur[i]] == maxID {
					cur[i]++
				}
			}
		} else {
			break
		}
	}

	return ids
}

type IndexIterator interface {
	Cur() ID
	Next() ID
}

var EmptyIndexIterator = &IndexIDIterator{}

// ----------------------------------------------------------------------------

type IndexIDIterator struct {
	ids   IDS
	index int
	value ID
}

func NewIndexIDIterator(ids IDS) *IndexIDIterator {
	value := ID(0)
	if len(ids) > 0 {
		value = ids[0]
	}
	return &IndexIDIterator{
		ids:   ids,
		index: 0,
		value: value,
	}
}

func (it *IndexIDIterator) Cur() ID {
	return it.value
}

func (it *IndexIDIterator) Next() ID {
	it.index++
	if it.index < len(it.ids) {
		it.value = it.ids[it.index]
		return it.value
	}
	it.value = 0
	return it.value
}

// ----------------------------------------------------------------------------

type UnionIndexIterator struct {
	iters []IndexIterator
	value ID
}

func NewUnionIndexIterator(iters ...IndexIterator) IndexIterator {
	iter := &UnionIndexIterator{
		iters: iters,
	}
	iter.Next()
	return iter
}

func (it *UnionIndexIterator) Next() ID {
	maxID := ID(0)
	for _, iter := range it.iters {
		if iter.Cur() != 0 && iter.Cur() > maxID {
			maxID = iter.Cur()
		}
	}
	if maxID > 0 {
		for _, iter := range it.iters {
			if iter.Cur() != 0 && iter.Cur() == maxID {
				iter.Next()
			}
		}

	}
	it.value = maxID
	return it.value
}

func (it *UnionIndexIterator) Cur() ID {
	return it.value
}

// ----------------------------------------------------------------------------

type IntersectIndexIterator struct {
	iters  []IndexIterator
	cur    IndexIterator
	value  ID
	finish bool
}

func NewIntersectIndexIterator(iters ...IndexIterator) IndexIterator {
	minIt := -1
	for i, iter := range iters {
		if minIt == -1 || iter.Cur() < iters[minIt].Cur() {
			minIt = i
		}
	}
	iter := &IntersectIndexIterator{
		iters: append(iters[:minIt], iters[minIt:]...),
		cur:   iters[minIt],
	}
	iter.Next()
	return iter
}

func (it *IntersectIndexIterator) Next() ID {
	if it.finish {
		return 0
	}
	for {
		founded := true
		for _, iter := range it.iters {
			for iter.Cur() != 0 && iter.Cur() > it.cur.Cur() {
				iter.Next()
			}
			if iter.Cur() == 0 {
				it.finish = true
				it.value = 0
				return it.value
			}
			if iter.Cur() != it.cur.Cur() {
				founded = false
			}
		}
		if founded {
			it.value = it.cur.Cur()
			it.cur.Next()
			return it.value
		}
		it.cur.Next()
		if it.cur.Cur() == 0 {
			it.finish = true
			it.value = 0
			return it.value
		}
	}
}

func (it *IntersectIndexIterator) Cur() ID {
	return it.value
}
