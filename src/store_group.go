package main

import (
	"sort"
)

func (store *Store) Group(group *Group, buffer *GroupsBuffer) {
	if group.ExpectEmpty() {
		return
	}

	filter := &group.Filter

	if filter.Likes == 0 {
		aggregation := store.index.Group.Get(
			group.FilterMask,
			group.KeysMask,
			group.GroupHash,
		)
		// Copy???
		if aggregation != nil {
			if group.OrderAsc() {
				if len(aggregation.Get()) > group.Limit() {
					buffer.groups = buffer.groups[:group.Limit()]
					entries := aggregation.Get()
					for i := range entries {
						if i >= group.Limit() {
							break
						}
						buffer.groups[i] = &entries[i]
					}
					// return entries[:group.Limit()], true
				} else {
					entries := aggregation.Get()
					buffer.groups = buffer.groups[:len(entries)]
					// copy(buffer.groups, ag)
					for i := range entries {
						buffer.groups[i] = &entries[i]
					}
				}
				buffer.orderAsc = true
				return
				// return entries, true
			}

			// OrderDesc
			if len(aggregation.Get()) > group.Limit() {
				buffer.groups = buffer.groups[:group.Limit()]
				// copy(buffer.groups, aggregation.Get()[len(aggregation.Get())-group.Limit():])
				entries := aggregation.Get()[len(aggregation.Get())-group.Limit():]
				for i := range entries {
					if i >= group.Limit() {
						break
					}
					buffer.groups[i] = &entries[i]
				}
				// return entries[len(entries)-group.Limit():], false
			} else {
				buffer.groups = buffer.groups[:len(aggregation.Get())]
				// copy(buffer.groups, aggregation.Get())
				entries := aggregation.Get()
				for i := range entries {
					buffer.groups[i] = &entries[i]
				}
			}
			buffer.orderAsc = false
			return

			// return entries, false
			// reverse := make([]*GroupEntry, len(entries))
			// for i, j := len(entries)-1, 0; i >= 0; i, j = i-1, j+1 {
			// 	reverse[j] = entries[i]
			// }
			// return reverse
		}

		// return make([]*GroupEntry, 0), true
		return
	}

	// scan by indexes
	// fmt.Println("from scan")

	aggregation := BorrowAggregation(store.dicts, group.KeysMask)
	defer aggregation.Release()

	iter := store.findGroupIds(filter)

	for iter.Cur() != 0 {
		account := store.get(iter.Cur())

		if !group.NoFilter() {
			if !store.groupFilterAccount(account, filter) {
				iter.Next()
				continue
			}
		}

		// group
		var groupHash GroupHash
		for _, key := range group.Keys {
			switch key {
			case GroupSex:
				groupHash.SetSex(account.Sex)
			case GroupStatus:
				groupHash.SetStatus(account.Status)
			case GroupCountry:
				if account.Country != 0 {
					groupHash.SetCountry(account.Country)
				}
			case GroupCity:
				if account.City != 0 {
					groupHash.SetCity(account.City)
				}
			}
		}

		if group.HasKey(GroupInterestsMask) {
			for _, interest := range account.Interests {
				groupHash.SetInterest(interest)
				aggregation.Append(groupHash)
			}
		} else {
			aggregation.Append(groupHash)
		}
		iter.Next()
	}

	if group.OrderAsc() {
		sort.Sort(aggregation)
	} else {
		sort.Sort(sort.Reverse(aggregation))
	}

	buffer.orderAsc = true
	entries := aggregation.Get()
	if len(entries) > group.Limit() {
		buffer.groups = buffer.groups[:group.Limit()]
		// copy(buffer.groups, entries)
		for i := range entries {
			if i >= group.Limit() {
				break
			}
			buffer.groups[i] = &entries[i]
		}
	} else {
		buffer.groups = buffer.groups[:len(entries)]
		// copy(buffer.groups, entries)
		for i := range entries {
			buffer.groups[i] = &entries[i]
		}
	}
}

func (store *Store) findGroupIds(filter *GroupFilter) IndexIterator {
	if filter.Likes != 0 {
		likee := filter.Likes
		filter.Likes = 0
		return store.index.Likee.Iter(likee)
	}
	if filter.City != 0 {
		city := filter.City
		filter.City = 0
		return store.index.City.Iter(city)
	}
	iters := make([]IndexIterator, 0)
	if filter.Interests != 0 {
		iters = append(iters, store.index.Interest.Iter(filter.Interests))
		filter.Interests = 0
	}
	if filter.Country != 0 {
		iters = append(iters, store.index.Country.Iter(filter.Country))
		filter.Country = 0
	}
	if filter.BirthYear != 0 {
		iters = append(iters, store.index.BirthYear.Iter(filter.BirthYear))
		filter.BirthYear = 0
	}
	if len(iters) == 1 {
		return iters[0]
	} else if len(iters) > 1 {
		return NewIntersectIndexIterator(iters...)
	}
	return store.index.ID.Iter()
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
		likes := store.index.Liker.Find(account.ID)
		if len(likes) == 0 {
			return false
		}
		exists := false
		for _, like := range likes {
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
