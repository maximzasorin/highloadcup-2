package main

import (
	"sort"
)

func (store *Store) Recommend(account *Account, recommend *Recommend, accounts *AccountsBuffer) {
	sex := byte(0)
	if account.Sex == SexFemale {
		sex = SexMale
	} else {
		sex = SexFemale
	}
	filter := &recommend.Filter

	if len(account.Interests) == 0 || recommend.ExpectEmpty() {
		return
	}

	if filter.City != 0 {
		interestIndexes := make([]IDS, len(account.Interests))
		for i, interest := range account.Interests {
			interestIndexes[i] = store.index.InterestPremium.FindByCity(interest, filter.City)
		}

		ids := UnionIndexes(interestIndexes...)

		recommendPairs := NewRecommendPairs(store, account, len(ids))
		for _, id := range ids {
			if account.ID == id {
				continue
			}
			pair := store.get(id)
			if account.Sex == pair.Sex {
				continue
			}
			recommendPairs.AddPair(pair)
		}

		recommendPairs.Sort()

		*accounts = (*accounts)[:len(recommendPairs.Get(recommend.Limit()))]
		copy(*accounts, recommendPairs.Get(recommend.Limit()))
		if len(*accounts) == recommend.Limit() {
			return
		}
	} else if filter.Country != 0 {
		interestIndexes := make([]IDS, len(account.Interests))
		for i, interest := range account.Interests {
			interestIndexes[i] = store.index.InterestPremium.FindByCountry(interest, filter.Country)
		}

		ids := UnionIndexes(interestIndexes...)

		recommendPairs := NewRecommendPairs(store, account, len(ids))
		for _, id := range ids {
			if account.ID == id {
				continue
			}
			pair := store.get(id)
			if account.Sex == pair.Sex {
				continue
			}
			recommendPairs.AddPair(pair)
		}

		recommendPairs.Sort()

		*accounts = (*accounts)[:len(recommendPairs.Get(recommend.Limit()))]
		copy(*accounts, recommendPairs.Get(recommend.Limit()))
		if len(*accounts) == recommend.Limit() {
			return
		}
	} else {
		// Single
		interestIndexes := make([]IDS, len(account.Interests))
		for i, interest := range account.Interests {
			interestIndexes[i] = store.index.InterestPremium.FindByStatusSex(interest, StatusSingle, sex)
		}

		ids := UnionIndexes(interestIndexes...)

		recommendPairs := NewRecommendPairs(store, account, len(ids))
		for _, id := range ids {
			if account.ID == id {
				continue
			}
			// pair := store.accounts[id]
			// if account.Sex == pair.Sex {
			// 	continue
			// }
			recommendPairs.AddPair(store.get(id))
		}

		recommendPairs.Sort()

		*accounts = (*accounts)[:len(recommendPairs.Get(recommend.Limit()))]
		copy(*accounts, recommendPairs.Get(recommend.Limit()))
		if len(*accounts) == recommend.Limit() {
			return
		}

		// Complicated
		interestIndexes = make([]IDS, len(account.Interests))
		for i, interest := range account.Interests {
			interestIndexes[i] = store.index.InterestPremium.FindByStatusSex(interest, StatusComplicated, sex)
		}

		ids = UnionIndexes(interestIndexes...)

		recommendPairs = NewRecommendPairs(store, account, len(ids))
		for _, id := range ids {
			if account.ID == id {
				continue
			}
			// pair := store.accounts[id]
			// if account.Sex == pair.Sex {
			// 	continue
			// }
			recommendPairs.AddPair(store.get(id))
		}

		recommendPairs.Sort()

		pairs := recommendPairs.Get(recommend.Limit() - len(*accounts))
		*accounts = append(*accounts, pairs...)

		if len(*accounts) == recommend.Limit() {
			return
		}

		// Relationship
		interestIndexes = make([]IDS, len(account.Interests))
		for i, interest := range account.Interests {
			interestIndexes[i] = store.index.InterestPremium.FindByStatusSex(interest, StatusRelationship, sex)
		}

		ids = UnionIndexes(interestIndexes...)

		recommendPairs = NewRecommendPairs(store, account, len(ids))
		for _, id := range ids {
			if account.ID == id {
				continue
			}
			// pair := store.accounts[id]
			// if account.Sex == pair.Sex {
			// 	continue
			// }
			recommendPairs.AddPair(store.get(id))
		}

		recommendPairs.Sort()

		pairs = recommendPairs.Get(recommend.Limit() - len(*accounts))
		*accounts = append(*accounts, pairs...)

		if len(*accounts) == recommend.Limit() {
			return
		}
	}

	// single

	interestIndexes := make([]IDS, len(account.Interests))
	for i, interest := range account.Interests {
		if filter.City != 0 {
			interestIndexes[i] = store.index.InterestSingle.FindByCity(interest, filter.City)
		} else if filter.Country != 0 {
			interestIndexes[i] = store.index.InterestSingle.FindByCountry(interest, filter.Country)
		} else {
			interestIndexes[i] = store.index.InterestSingle.Find(interest)
		}
	}

	ids := UnionIndexes(interestIndexes...)

	recommendPairs := NewRecommendPairs(store, account, len(ids))
	for _, id := range ids {
		if account.ID == id {
			continue
		}
		pair := store.get(id)
		if account.Sex == pair.Sex {
			continue
		}
		recommendPairs.AddPair(pair)
	}

	recommendPairs.Sort()

	pairs := recommendPairs.Get(recommend.Limit() - len(*accounts))
	*accounts = append(*accounts, pairs...)

	if len(*accounts) == recommend.Limit() {
		return
	}

	// complicated

	interestIndexes = make([]IDS, len(account.Interests))
	for i, interest := range account.Interests {
		if filter.City != 0 {
			interestIndexes[i] = store.index.InterestComplicated.FindByCity(interest, filter.City)
		} else if filter.Country != 0 {
			interestIndexes[i] = store.index.InterestComplicated.FindByCountry(interest, filter.Country)
		} else {
			interestIndexes[i] = store.index.InterestComplicated.Find(interest)
		}
	}

	ids = UnionIndexes(interestIndexes...)

	recommendPairs = NewRecommendPairs(store, account, len(ids))
	for _, id := range ids {
		if account.ID == id {
			continue
		}
		pair := store.get(id)
		if account.Sex == pair.Sex {
			continue
		}
		recommendPairs.AddPair(pair)
	}

	recommendPairs.Sort()

	pairs = recommendPairs.Get(recommend.Limit() - len(*accounts))
	*accounts = append(*accounts, pairs...)

	if len(*accounts) == recommend.Limit() {
		return
	}

	// relationship

	interestIndexes = make([]IDS, len(account.Interests))
	for i, interest := range account.Interests {
		if filter.City != 0 {
			interestIndexes[i] = store.index.InterestRelationship.FindByCity(interest, filter.City)
		} else if filter.Country != 0 {
			interestIndexes[i] = store.index.InterestRelationship.FindByCountry(interest, filter.Country)
		} else {
			interestIndexes[i] = store.index.InterestRelationship.Find(interest)
		}
	}

	ids = UnionIndexes(interestIndexes...)

	recommendPairs = NewRecommendPairs(store, account, len(ids))
	for _, id := range ids {
		if account.ID == id {
			continue
		}
		pair := store.get(id)
		if account.Sex == pair.Sex {
			continue
		}
		recommendPairs.AddPair(pair)
	}

	recommendPairs.Sort()

	pairs = recommendPairs.Get(recommend.Limit() - len(*accounts))
	*accounts = append(*accounts, pairs...)
}

// type RecommendPair struct {
// 	account     *Account
// 	compability uint64
// }

type RecommendPairs struct {
	store     *Store
	account   *Account
	pairs     []*Account
	pairComps []uint64
}

func NewRecommendPairs(store *Store, account *Account, capacity int) *RecommendPairs {
	return &RecommendPairs{
		store:     store,
		account:   account,
		pairs:     make([]*Account, 0, capacity),
		pairComps: make([]uint64, 0, capacity),
	}
}

func (ra *RecommendPairs) AddPair(pair *Account) {
	ra.pairs = append(ra.pairs, pair)
	ra.pairComps = append(ra.pairComps, Compability(ra.account, pair))
}

func (ra *RecommendPairs) Sort() {
	sort.Sort(ra)
}

func (ra *RecommendPairs) Get(limit int) []*Account {
	if len(ra.pairs) > limit {
		return ra.pairs[:limit]
	}
	return ra.pairs
}

func (ra *RecommendPairs) Len() int {
	return len(ra.pairs)
}

func (ra *RecommendPairs) Swap(i, j int) {
	ra.pairs[i], ra.pairs[j] = ra.pairs[j], ra.pairs[i]
	ra.pairComps[i], ra.pairComps[j] = ra.pairComps[j], ra.pairComps[i]
}

func (ra *RecommendPairs) Less(i, j int) bool {
	return ra.pairComps[i] > ra.pairComps[j]
}
