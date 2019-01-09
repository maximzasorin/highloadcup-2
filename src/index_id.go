package main

import "sort"

type IDS []uint32

type IndexID struct {
	ids IDS
}

func NewIndexID() *IndexID {
	return &IndexID{
		ids: make([]uint32, 0, 10000),
	}
}

func (indexID *IndexID) FindAll() IDS {
	return indexID.ids
}

func (indexID *IndexID) Update() {
	// if !sort.IsSorted(indexID.ids) {
	sort.Sort(indexID.ids)
	// }
}

func (indexID *IndexID) Add(ID uint32) {
	indexID.ids = append(indexID.ids, ID)
}

func (indexID *IndexID) Remove(ID uint32) {
	n := len(indexID.ids)
	i := sort.Search(n, func(i int) bool {
		return indexID.ids[i] == ID
	})

	if i != n {
		indexID.ids = append(indexID.ids[:i], indexID.ids[i+1:]...)
	}
}

func (ids IDS) Len() int           { return len(ids) }
func (ids IDS) Swap(i, j int)      { ids[i], ids[j] = ids[j], ids[i] }
func (ids IDS) Less(i, j int) bool { return ids[i] > ids[j] }
