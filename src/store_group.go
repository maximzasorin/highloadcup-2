package main

import "sort"

func (store *Store) GroupAll(group *Group) ([]*GroupEntry, bool) {
	filter := &group.Filter

	if filter.ExpectEmpty {
		return make([]*GroupEntry, 0), true
	}

	if filter.Likes == 0 {
		aggregation := store.indexGroup.Get(
			group.FilterHash,
			group.KeysHash,
			group.Entry.GetHash(),
		)
		if aggregation != nil {
			// fmt.Println("from index")
			entries := aggregation.Get()

			if group.OrderAsc {
				if len(entries) > int(group.Limit) {
					return entries[:group.Limit], true
				}
				return entries, true
			}
			if len(entries) > int(group.Limit) {
				return entries[len(entries)-int(group.Limit):], false
			}
			return entries, false
			// reverse := make([]*GroupEntry, len(entries))
			// for i, j := len(entries)-1, 0; i >= 0; i, j = i-1, j+1 {
			// 	reverse[j] = entries[i]
			// }
			// return reverse
		}

		return make([]*GroupEntry, 0), true
	}

	// scan by indexes
	// fmt.Println("from scan")

	aggregation := NewAggregation(store.dicts, group.KeysHash)
	groupEntry := NewGroupEntry(0)

	for _, id := range store.findGroupIds(filter) {
		account := store.get(id)

		if !filter.NoFilter {
			if !store.groupFilterAccount(account, filter) {
				continue
			}
		}

		// group
		groupEntry.Reset()
		for _, key := range group.Keys {
			switch key {
			case GroupSex:
				groupEntry.SetSex(account.Sex)
			case GroupStatus:
				groupEntry.SetStatus(account.Status)
			case GroupCountry:
				if account.Country != 0 {
					groupEntry.SetCountry(account.Country)
				}
			case GroupCity:
				if account.City != 0 {
					groupEntry.SetCity(account.City)
				}
			}
		}

		if group.HasKey(GroupInterestsMask) {
			for _, interest := range account.Interests {
				groupEntry.SetInterest(interest)
				aggregation.Append(groupEntry.GetHash())
			}
		} else {
			aggregation.Append(groupEntry.GetHash())
		}
	}

	if group.OrderAsc {
		sort.Sort(aggregation)
	} else {
		sort.Sort(sort.Reverse(aggregation))
	}

	entries := aggregation.Get()
	if len(entries) > int(group.Limit) {
		return entries[:group.Limit], true
	}
	return entries, true
}

func (store *Store) findGroupIds(filter *GroupFilter) IDS {
	if filter.Likes != 0 {
		likee := filter.Likes
		filter.Likes = 0
		return store.indexLikee.Find(likee)
	}
	if filter.City != 0 {
		city := filter.City
		filter.City = 0
		return store.indexCity.Find(city)
	}
	indexes := make([]IDS, 0)
	if filter.Interests != 0 {
		indexes = append(indexes, store.indexInterest.Find(filter.Interests))
		filter.Interests = 0
	}
	if filter.Country != 0 {
		indexes = append(indexes, store.indexCountry.Find(filter.Country))
		filter.Country = 0
	}
	if filter.BirthYear != 0 {
		indexes = append(indexes, store.indexBirthYear.Find(filter.BirthYear))
		filter.BirthYear = 0
	}
	// if filter.JoinedYear != nil {
	// 	indexes = append(indexes, store.indexJoinedYear.Find(filter.JoinedYear))
	// 	useIndexes = true
	// 	filter.JoinedYear = nil
	// }
	if len(indexes) == 1 {
		return indexes[0]
	} else if len(indexes) > 1 {
		return IntersectIndexes(indexes...)
	}
	return store.indexID.FindAll()
}

func (store *Store) groupFilterAccount(account *Account, filter *GroupFilter) bool {
	if filter.Sex != 0 {
		if account.Sex != filter.Sex {
			return false
		}
	}

	if filter.Status != 0 {
		if account.Status != filter.Status {
			return false
		}
	}

	if filter.Country != 0 {
		if account.Country == 0 {
			return false
		}

		if account.Country != filter.Country {
			return false
		}
	}

	if filter.City != 0 {
		if account.City == 0 {
			return false
		}

		if account.City != filter.City {
			return false
		}
	}

	if filter.BirthYear != 0 {
		if account.Birth < filter.BirthYearGte || account.Birth > filter.BirthYearLte {
			return false
		}
	}

	if filter.Interests != 0 {
		if len(account.Interests) == 0 {
			return false
		}
		exists := false
		for _, interest := range account.Interests {
			if interest == filter.Interests {
				exists = true
				break
			}
		}
		if !exists {
			return false
		}
	}

	if filter.Likes != 0 {
		if len(account.Likes) == 0 {
			return false
		}
		exists := false
		for _, like := range account.Likes {
			if like.ID == ID(filter.Likes) {
				exists = true
				break
			}
		}
		if !exists {
			return false
		}
	}

	if filter.JoinedYear != 0 {
		if account.Joined < filter.JoinedYearGte || account.Joined > filter.JoinedYearLte {
			return false
		}
	}

	return true
}
