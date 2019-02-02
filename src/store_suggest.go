package main

import (
	"sort"
	"sync"
)

func (store *Store) Suggest(account *Account, suggest *Suggest, accounts *AccountsBuffer) {
	filter := suggest.Filter

	if suggest.ExpectEmpty() {
		return
	}

	similarLikerIndexes := make([]IDS, 0)
	for _, like := range store.index.Liker.Find(account.ID) {
		similarLikerIndexes = append(similarLikerIndexes, store.index.Likee.Find(like.ID))
	}

	ids := UnionIndexes(similarLikerIndexes...)

	similarLikers := BorrowSimilarLikers(store, account)
	defer similarLikers.Release()

	// similarLikers := NewSimilarLikers(store, account, len(ids))
	for _, id := range ids {
		liker := store.get(id)
		if account.Sex != liker.Sex {
			continue
		}
		if filter.City != 0 {
			if filter.City != liker.City {
				continue
			}
		}
		if filter.Country != 0 {
			if filter.Country != liker.Country {
				continue
			}
		}
		similarLikers.Add(liker)
	}

	sort.Sort(similarLikers)

	suggestIDs := BorrowIDS()
	defer suggestIDs.Release()

	taked := make(map[ID]bool)
	prevSug := 0
	for _, similarLiker := range similarLikers.Get() {
		for _, like := range store.index.Liker.Find(similarLiker.ID) {
			existsMyLike := false
			for _, myLike := range store.index.Liker.Find(account.ID) {
				if myLike.ID == like.ID {
					existsMyLike = true
					break
				}
			}
			if !existsMyLike {
				if _, ok := taked[like.ID]; !ok {
					*suggestIDs = append(*suggestIDs, like.ID)
					taked[like.ID] = true
				}
			}
		}
		sort.Sort((*suggestIDs)[prevSug:])
		prevSug = len(*suggestIDs)
		if len(*suggestIDs) >= suggest.Limit() {
			break
		}
	}

	for _, suggestID := range *suggestIDs {
		*accounts = append(*accounts, store.get(suggestID))
		if len(*accounts) >= suggest.Limit() {
			break
		}
	}
}

type SimilarLikers struct {
	store     *Store
	account   *Account
	likers    []*Account
	likerSims []float64
}

var similarLikersPool = sync.Pool{
	New: func() interface{} {
		return &SimilarLikers{
			likers:    make([]*Account, 0, 4*1024),
			likerSims: make([]float64, 0, 4*1024),
		}
	},
}

func BorrowSimilarLikers(store *Store, account *Account) *SimilarLikers {
	sl := similarLikersPool.Get().(*SimilarLikers)
	sl.Reset()
	sl.store = store
	sl.account = account
	return sl
}

func NewSimilarLikers(store *Store, account *Account, capacity int) *SimilarLikers {
	return &SimilarLikers{
		store:     store,
		account:   account,
		likers:    make([]*Account, 0, capacity),
		likerSims: make([]float64, 0, capacity),
	}
}

func (sm *SimilarLikers) Reset() {
	sm.likers = sm.likers[:0]
	sm.likerSims = sm.likerSims[:0]
}

func (sm *SimilarLikers) Release() {
	similarLikersPool.Put(sm)
}

func (sm *SimilarLikers) Add(liker *Account) {
	sm.likers = append(sm.likers, liker)
	sm.likerSims = append(sm.likerSims, sm.store.Similarity(sm.account, liker))
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
