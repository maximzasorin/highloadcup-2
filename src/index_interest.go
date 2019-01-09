package main

import "sort"

type IndexInterest struct {
	interests map[Interest]IDS
}

func NewIndexInterest() *IndexInterest {
	return &IndexInterest{
		interests: make(map[Interest]IDS),
	}
}

func (indexInterest *IndexInterest) Add(interest Interest, ID uint32) {
	_, ok := indexInterest.interests[interest]
	if !ok {
		indexInterest.interests[interest] = make([]uint32, 1)
		indexInterest.interests[interest][0] = ID
		return
	}

	indexInterest.interests[interest] = append(indexInterest.interests[interest], ID)
}

func (indexInterest *IndexInterest) Remove(interest Interest, ID uint32) {
	_, ok := indexInterest.interests[interest]
	if !ok {
		return
	}
	for i, accountID := range indexInterest.interests[interest] {
		if accountID == ID {
			indexInterest.interests[interest] = append(indexInterest.interests[interest][:i], indexInterest.interests[interest][i+1:]...)
			return
		}
	}
}

func (indexInterest *IndexInterest) Update(interest Interest) {
	if interest == 0 {
		for interest := range indexInterest.interests {
			sort.Sort(indexInterest.interests[interest])
		}
		return
	}

	if _, ok := indexInterest.interests[interest]; ok {
		sort.Sort(indexInterest.interests[interest])
	}
}

func (indexInterest *IndexInterest) Get(interest Interest) IDS {
	if _, ok := indexInterest.interests[interest]; ok {
		return indexInterest.interests[interest]
	}
	return make(IDS, 0)
}
