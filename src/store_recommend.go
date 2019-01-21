package main

import (
	"sort"
)

func (store *Store) RecommendAll(account *Account, recommend *Recommend) []*Account {
	sex := byte(0)
	if account.Sex == SexFemale {
		sex = SexMale
	} else {
		sex = SexFemale
	}
	filter := &recommend.Filter
	allPairs := make([]*Account, 0)

	if len(account.Interests) == 0 || recommend.ExpectEmpty {
		return allPairs
	}

	if filter.City != nil {
		interestIndexes := make([]IDS, len(account.Interests))
		for i, interest := range account.Interests {
			interestIndexes[i] = store.indexInterestPremium.FindByCity(interest, *filter.City)
		}

		ids := UnionIndexes(interestIndexes...)

		// fmt.Println("Founded =", len(ids))

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

		allPairs = recommendPairs.Get(recommend.Limit)

		if len(allPairs) == recommend.Limit {
			return allPairs
		}
	} else if filter.Country != nil {
		interestIndexes := make([]IDS, len(account.Interests))
		for i, interest := range account.Interests {
			interestIndexes[i] = store.indexInterestPremium.FindByCountry(interest, *filter.Country)
		}

		ids := UnionIndexes(interestIndexes...)

		// fmt.Println("Founded =", len(ids))

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

		allPairs = recommendPairs.Get(recommend.Limit)

		if len(allPairs) == recommend.Limit {
			return allPairs
		}
	} else {
		// Single
		interestIndexes := make([]IDS, len(account.Interests))
		for i, interest := range account.Interests {
			interestIndexes[i] = store.indexInterestPremium.FindByStatusSex(interest, StatusSingle, sex)
		}

		ids := UnionIndexes(interestIndexes...)

		// fmt.Println("Founded =", len(ids))

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

		allPairs = recommendPairs.Get(recommend.Limit)

		if len(allPairs) == recommend.Limit {
			return allPairs
		}

		// Complicated
		interestIndexes = make([]IDS, len(account.Interests))
		for i, interest := range account.Interests {
			interestIndexes[i] = store.indexInterestPremium.FindByStatusSex(interest, StatusComplicated, sex)
		}

		ids = UnionIndexes(interestIndexes...)

		// fmt.Println("Founded =", len(ids))

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

		pairs := recommendPairs.Get(recommend.Limit - len(allPairs))
		allPairs = append(allPairs, pairs...)

		if len(allPairs) == recommend.Limit {
			return allPairs
		}

		// Relationship
		interestIndexes = make([]IDS, len(account.Interests))
		for i, interest := range account.Interests {
			interestIndexes[i] = store.indexInterestPremium.FindByStatusSex(interest, StatusRelationship, sex)
		}

		ids = UnionIndexes(interestIndexes...)

		// fmt.Println("Founded =", len(ids))

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

		pairs = recommendPairs.Get(recommend.Limit - len(allPairs))
		allPairs = append(allPairs, pairs...)

		if len(allPairs) == recommend.Limit {
			return allPairs
		}
	}

	// single

	interestIndexes := make([]IDS, len(account.Interests))
	for i, interest := range account.Interests {
		if filter.City != nil {
			interestIndexes[i] = store.indexInterestSingle.FindByCity(interest, *filter.City)
		} else if filter.Country != nil {
			interestIndexes[i] = store.indexInterestSingle.FindByCountry(interest, *filter.Country)
		} else {
			interestIndexes[i] = store.indexInterestSingle.Find(interest)
		}
	}

	ids := UnionIndexes(interestIndexes...)

	// fmt.Println("Founded =", len(ids))

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

	pairs := recommendPairs.Get(recommend.Limit - len(allPairs))
	allPairs = append(allPairs, pairs...)

	if len(allPairs) == recommend.Limit {
		return allPairs
	}

	// complicated

	interestIndexes = make([]IDS, len(account.Interests))
	for i, interest := range account.Interests {
		if filter.City != nil {
			interestIndexes[i] = store.indexInterestComplicated.FindByCity(interest, *filter.City)
		} else if filter.Country != nil {
			interestIndexes[i] = store.indexInterestComplicated.FindByCountry(interest, *filter.Country)
		} else {
			interestIndexes[i] = store.indexInterestComplicated.Find(interest)
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

	pairs = recommendPairs.Get(recommend.Limit - len(allPairs))
	allPairs = append(allPairs, pairs...)

	if len(allPairs) == recommend.Limit {
		return allPairs
	}

	// relationship

	interestIndexes = make([]IDS, len(account.Interests))
	for i, interest := range account.Interests {
		if filter.City != nil {
			interestIndexes[i] = store.indexInterestRelationship.FindByCity(interest, *filter.City)
		} else if filter.Country != nil {
			interestIndexes[i] = store.indexInterestRelationship.FindByCountry(interest, *filter.Country)
		} else {
			interestIndexes[i] = store.indexInterestRelationship.Find(interest)
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

	pairs = recommendPairs.Get(recommend.Limit - len(allPairs))
	allPairs = append(allPairs, pairs...)

	return allPairs
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
	// res := make([]*Account, 0)
	// for _, rp := range ra.pairs {
	// 	res = append(res, rp.account)
	// 	if len(res) >= limit {
	// 		break
	// 	}
	// }
	// return res
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
	// a := ra.pairs[i]
	// b := ra.pairs[j]

	// intsA := 0
	// intsB := 0
	// for _, interest := range ra.account.Interests {
	// 	for _, intA := range a.Interests {
	// 		if intA == interest {
	// 			intsA++
	// 		}
	// 	}
	// 	for _, intB := range b.Interests {
	// 		if intB == interest {
	// 			intsB++
	// 		}
	// 	}
	// }

	// if ra.store.PremiumNow(a) && !ra.store.PremiumNow(b) {
	// 	return true
	// }
	// if !ra.store.PremiumNow(a) && ra.store.PremiumNow(b) {
	// 	return false
	// }

	// // compA := Compability(ra.account, a)
	// // compB := Compability(ra.account, b)

	// // if compA > compB {
	// // 	return true
	// // }

	// // if compB < compA {
	// // 	return false
	// // }

	// if a.Status != b.Status {
	// 	if a.Status == StatusSingle {
	// 		return true
	// 	}
	// 	if b.Status == StatusSingle {
	// 		return false
	// 	}
	// 	if a.Status == StatusComplicated {
	// 		return true
	// 	}
	// 	if b.Status == StatusComplicated {
	// 		return false
	// 	}
	// }

	// if intsA > intsB {
	// 	return true
	// }
	// if intsA < intsB {
	// 	return false
	// }

	// diffA := ra.account.Birth - a.Birth
	// if diffA < 0 {
	// 	diffA = -diffA
	// }
	// diffB := ra.account.Birth - b.Birth
	// if diffB < 0 {
	// 	diffB = -diffB
	// }
	// if diffA < diffB {
	// 	return true
	// }
	// if diffA > diffB {
	// 	return false
	// }

	// return a.ID < b.ID
}
