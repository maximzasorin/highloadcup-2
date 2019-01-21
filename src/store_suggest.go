package main

import (
	"sort"
)

func (store *Store) SuggestAll(account *Account, suggest *Suggest) []*Account {
	filter := suggest.Filter
	suggestAccounts := make([]*Account, 0)

	if suggest.ExpectEmpty {
		return suggestAccounts
	}

	myLikees := make(IDS, 0)
	for _, like := range account.Likes {
		myLikees = append(myLikees, like.ID)
	}

	similarLikerIndexes := make([]IDS, 0)
	for _, myLikee := range myLikees {
		similarLikerIndexes = append(similarLikerIndexes, store.indexLikee.Find(myLikee))
	}

	ids := UnionIndexes(similarLikerIndexes...)

	similarLikers := NewSimilarLikers(store, account, len(ids))
	for _, id := range ids {
		liker := store.get(id)
		if account.Sex != liker.Sex {
			continue
		}
		if filter.City != nil {
			if *filter.City != liker.City {
				continue
			}
		}
		if filter.Country != nil {
			if *filter.Country != liker.Country {
				continue
			}
		}
		similarLikers.Add(liker)
	}

	// fmt.Println("len similar likers =", similarLikers.Len())

	sort.Sort(similarLikers)

	suggestIDs := make(IDS, 0)
	taked := make(map[ID]bool)
	prevSug := 0
	for _, similarLiker := range similarLikers.Get() {
		for _, like := range similarLiker.Likes {
			existsMyLike := false
			for _, myLike := range account.Likes {
				if myLike.ID == like.ID {
					existsMyLike = true
					break
				}
			}
			if !existsMyLike {
				if _, ok := taked[like.ID]; !ok {
					suggestIDs = append(suggestIDs, like.ID)
					taked[like.ID] = true
				}
			}
		}
		sort.Sort(suggestIDs[prevSug:])
		prevSug = len(suggestIDs)
		if len(suggestIDs) >= suggest.Limit {
			break
		}
	}

	// sort.Sort(suggestIDs)

	for _, suggestID := range suggestIDs {
		suggestAccounts = append(suggestAccounts, store.get(suggestID))
		if len(suggestAccounts) >= suggest.Limit {
			break
		}
	}
	// fmt.Println("similar likers =", len(ids))
	// fmt.Println("filtered similar likers =", similarLikers.Len())

	return suggestAccounts
}

type SimilarLikers struct {
	store     *Store
	account   *Account
	likers    []*Account
	likerSims []float64
}

func NewSimilarLikers(store *Store, account *Account, capacity int) *SimilarLikers {
	return &SimilarLikers{
		store:     store,
		account:   account,
		likers:    make([]*Account, 0, capacity),
		likerSims: make([]float64, 0, capacity),
	}
}

func (sm *SimilarLikers) Add(liker *Account) {
	sm.likers = append(sm.likers, liker)
	sm.likerSims = append(sm.likerSims, Similarity(sm.account, liker))
	// fmt.Printf("id = %d, sim = %f\n", liker.ID, Similarity(sm.account, liker))
}

func (sm *SimilarLikers) Get() []*Account {
	return sm.likers
}

func (sm *SimilarLikers) Len() int {
	return len(sm.likers)
}

func (sm *SimilarLikers) Swap(i, j int) {
	sm.likers[i], sm.likers[j] = sm.likers[j], sm.likers[i]
	sm.likerSims[i], sm.likerSims[j] = sm.likerSims[j], sm.likerSims[i]
}

func (sm *SimilarLikers) Less(i, j int) bool {
	return sm.likerSims[i] > sm.likerSims[j]
}
