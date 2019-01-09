package main

import "sort"

type IndexLikee struct {
	likees map[uint32]IDS // list?
}

func NewIndexLikee() *IndexLikee {
	return &IndexLikee{
		likees: make(map[uint32]IDS),
	}
}

func (indexLikee *IndexLikee) Add(likee uint32, liker uint32) {
	_, ok := indexLikee.likees[likee]
	if !ok {
		indexLikee.likees[likee] = make([]uint32, 1)
		indexLikee.likees[likee][0] = liker
		return
	}

	indexLikee.likees[likee] = append(indexLikee.likees[likee], liker)
}

func (indexLikee *IndexLikee) Update(likee uint32) {
	if likee == 0 {
		for likee := range indexLikee.likees {
			sort.Sort(indexLikee.likees[likee])
		}
		return
	}

	if _, ok := indexLikee.likees[likee]; ok {
		sort.Sort(indexLikee.likees[likee])
	}
}

func (indexLikee *IndexLikee) Get(likee uint32) IDS {
	if _, ok := indexLikee.likees[likee]; ok {
		return indexLikee.likees[likee]
	}
	return make(IDS, 0)
}
