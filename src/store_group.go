package main

// func (store *Store) GroupAll(group *Group) []*GroupEntry {
// 	filter := &group.Filter
// 	keys := &group.Keys

// 	if filter.ExpectEmpty {
// 		return make([]*GroupEntry, 0)
// 	}

// 	// precalculate
// 	groupEntries, ok := store.indexGroup.Get(group.AllBits)
// 	if ok {
// 		return store.groupFromEntries(group, groupEntries, filter)
// 	}

// 	// // group
// 	// fields := make([]string, 0)
// 	// if group.HasKey(GroupSex) || filter.Sex != nil {
// 	// 	fields = append(fields, "sex")
// 	// }
// 	// if group.HasKey(GroupStatus) || filter.Status != nil {
// 	// 	fields = append(fields, "status")
// 	// }
// 	// if group.HasKey(GroupCity) || filter.City != nil {
// 	// 	fields = append(fields, "city")
// 	// }
// 	// if group.HasKey(GroupCountry) || filter.Country != nil {
// 	// 	fields = append(fields, "country")
// 	// }
// 	// if group.HasKey(GroupInterests) || filter.Interests != nil {
// 	// 	fields = append(fields, "interests")
// 	// }
// 	// if filter.Likes != nil {
// 	// 	fields = append(fields, "likes")
// 	// }
// 	// if filter.BirthYear != nil {
// 	// 	fields = append(fields, "birth")
// 	// }
// 	// if filter.JoinedYear != nil {
// 	// 	fields = append(fields, "joined")
// 	// }

// 	// fieldsKey := strings.Join(fields, ",")
// 	// // fmt.Println("fieldsKey =", fieldsKey)
// 	// entries, ok := store.indexGroup.Get(fieldsKey)
// 	// if ok {
// 	// 	// fmt.Println("from index")
// 	// 	for _, entry := range entries {
// 	// 		if filter.Sex != nil {
// 	// 			if *filter.Sex != entry.Sex {
// 	// 				continue
// 	// 			}
// 	// 		}
// 	// 		if filter.Status != nil {
// 	// 			if *filter.Status != entry.Status {
// 	// 				continue
// 	// 			}
// 	// 		}
// 	// 		if filter.City != nil {
// 	// 			if *filter.City != entry.City {
// 	// 				continue
// 	// 			}
// 	// 		}
// 	// 		if filter.Country != nil {
// 	// 			if *filter.Country != entry.Country {
// 	// 				continue
// 	// 			}
// 	// 		}
// 	// 		if filter.BirthYear != nil {
// 	// 			if *filter.BirthYear != entry.BirthYear {
// 	// 				continue
// 	// 			}
// 	// 		}
// 	// 		if filter.JoinedYear != nil {
// 	// 			if *filter.JoinedYear != entry.JoinedYear {
// 	// 				continue
// 	// 			}
// 	// 		}
// 	// 		if filter.Interests != nil {
// 	// 			if *filter.Interests != entry.Interest {
// 	// 				continue
// 	// 			}
// 	// 		}
// 	// 		ag := AggregationGroup{Count: entry.Count}
// 	// 		if entry.Sex != 0 && group.HasKey(GroupSex) {
// 	// 			ag.Sex = entry.Sex
// 	// 		}
// 	// 		if entry.Status != 0 && group.HasKey(GroupStatus) {
// 	// 			ag.Status = entry.Status
// 	// 		}
// 	// 		if entry.City != 0 && group.HasKey(GroupCity) {
// 	// 			ag.City = entry.City
// 	// 		}
// 	// 		if entry.Country != 0 && group.HasKey(GroupCountry) {
// 	// 			ag.Country = entry.Country
// 	// 		}
// 	// 		if entry.Interest != 0 && group.HasKey(GroupInterests) {
// 	// 			ag.Interest = entry.Interest
// 	// 		}
// 	// 		aggregation.Add(ag)
// 	// 	}
// 	// 	aggregation.Sort(*group.OrderAsc)
// 	// 	aggregation.Limit(*group.Limit)
// 	// 	return &aggregation
// 	// }

// 	// scan by indexes
// 	aggregation := NewAggregation(group)

// 	for _, id := range store.findGroupIds(filter) {
// 		account := store.accounts[id]

// 		if !filter.NoFilter {
// 			if !store.groupFilterAccount(account, filter) {
// 				continue
// 			}
// 		}

// 		// group
// 		groupEntry := GroupEntry{Count: 1}

// 		for _, key := range *keys {
// 			switch key {
// 			case GroupSex:
// 				groupEntry.Sex = account.Sex
// 			case GroupStatus:
// 				groupEntry.Status = account.Status
// 			case GroupCountry:
// 				if account.Country != 0 {
// 					groupEntry.Country = account.Country
// 				}
// 			case GroupCity:
// 				if account.City != 0 {
// 					groupEntry.City = account.City
// 				}
// 			}
// 		}

// 		if group.HasKey(GroupInterests) {
// 			for _, interest := range account.Interests {
// 				ge := &GroupEntry{
// 					Sex:      groupEntry.Sex,
// 					Status:   groupEntry.Status,
// 					Country:  groupEntry.Country,
// 					City:     groupEntry.City,
// 					Interest: interest,
// 					Count:    1,
// 				}
// 				ge.Hash = CreateGroupHash(ge)
// 				aggregation.Add(ge)
// 			}
// 		} else {
// 			groupEntry.Hash = CreateGroupHash(&groupEntry)
// 			aggregation.Add(&groupEntry)
// 		}
// 	}

// 	return aggregation.Get(*group.OrderAsc, int(*group.Limit))
// }

// func (store *Store) groupFromEntries(group *Group, groupEntries map[GroupHash]*GroupEntry, filter *GroupFilter) []*GroupEntry {
// 	aggregation := NewAggregation(group)

// 	// filterEntry := &GroupEntry{}
// 	// if filter.Sex != nil {
// 	// 	filterEntry.Sex = *filter.Sex
// 	// }
// 	// if filter.Status != nil {
// 	// 	filterEntry.Status = *filter.Status
// 	// }
// 	// if filter.City != nil {
// 	// 	filterEntry.City = *filter.City
// 	// }
// 	// if filter.Country != nil {
// 	// 	filterEntry.Country = *filter.Country
// 	// }
// 	// if filter.BirthYear != nil {
// 	// 	filterEntry.BirthYear = *filter.BirthYear
// 	// }
// 	// if filter.JoinedYear != nil {
// 	// 	filterEntry.JoinedYear = *filter.JoinedYear
// 	// }
// 	// if filter.Interests != nil {
// 	// 	filterEntry.Interest = *filter.Interests
// 	// }
// 	// filterHash := CreateGroupHash(filterEntry)
// 	for _, groupEntry := range groupEntries {
// 		// if groupHash&filterHash != filterHash {
// 		// 	continue
// 		// }
// 		if filter.Sex != nil {
// 			if *filter.Sex != groupEntry.Sex {
// 				continue
// 			}
// 		}
// 		if filter.Status != nil {
// 			if *filter.Status != groupEntry.Status {
// 				continue
// 			}
// 		}
// 		if filter.City != nil {
// 			if *filter.City != groupEntry.City {
// 				continue
// 			}
// 		}
// 		if filter.Country != nil {
// 			if *filter.Country != groupEntry.Country {
// 				continue
// 			}
// 		}
// 		if filter.BirthYear != nil {
// 			if *filter.BirthYear != groupEntry.BirthYear {
// 				continue
// 			}
// 		}
// 		if filter.JoinedYear != nil {
// 			if *filter.JoinedYear != groupEntry.JoinedYear {
// 				continue
// 			}
// 		}
// 		if filter.Interests != nil {
// 			if *filter.Interests != groupEntry.Interest {
// 				continue
// 			}
// 		}
// 		ge := &GroupEntry{Count: groupEntry.Count}
// 		if groupEntry.Sex != 0 && group.HasKey(GroupSex) {
// 			ge.Sex = groupEntry.Sex
// 		}
// 		if groupEntry.Status != 0 && group.HasKey(GroupStatus) {
// 			ge.Status = groupEntry.Status
// 		}
// 		if groupEntry.City != 0 && group.HasKey(GroupCity) {
// 			ge.City = groupEntry.City
// 		}
// 		if groupEntry.Country != 0 && group.HasKey(GroupCountry) {
// 			ge.Country = groupEntry.Country
// 		}
// 		if groupEntry.Interest != 0 && group.HasKey(GroupInterests) {
// 			ge.Interest = groupEntry.Interest
// 		}
// 		ge.Hash = CreateGroupHash(ge)
// 		aggregation.Add(ge)
// 	}
// 	return aggregation.Get(*group.OrderAsc, int(*group.Limit))
// }

// func (store *Store) findGroupIds(filter *GroupFilter) IDS {
// 	if filter.Likes != nil {
// 		likee := ID(*filter.Likes)
// 		filter.Likes = nil
// 		return store.indexLikee.Find(likee)
// 	}
// 	if filter.City != nil {
// 		city := *filter.City
// 		filter.City = nil
// 		return store.indexCity.Find(city)
// 	}
// 	indexes := make([]IDS, 0)
// 	if filter.Interests != nil {
// 		indexes = append(indexes, store.indexInterest.Find(*filter.Interests))
// 		filter.Interests = nil
// 	}
// 	if filter.Country != nil {
// 		indexes = append(indexes, store.indexCountry.Find(*filter.Country))
// 		filter.Country = nil
// 	}
// 	if filter.BirthYear != nil {
// 		indexes = append(indexes, store.indexBirthYear.Find(*filter.BirthYear))
// 		filter.BirthYear = nil
// 	}
// 	// if filter.JoinedYear != nil {
// 	// 	indexes = append(indexes, store.indexJoinedYear.Find(*filter.JoinedYear))
// 	// 	useIndexes = true
// 	// 	filter.JoinedYear = nil
// 	// }
// 	if len(indexes) == 1 {
// 		return indexes[0]
// 	} else if len(indexes) > 1 {
// 		return IntersectIndexes(indexes...)
// 	}
// 	return store.indexID.FindAll()
// }

// func (store *Store) groupFilterAccount(account *Account, filter *GroupFilter) bool {
// 	if filter.Sex != nil {
// 		if account.Sex != *filter.Sex {
// 			return false
// 		}
// 	}

// 	if filter.Status != nil {
// 		if account.Status != *filter.Status {
// 			return false
// 		}
// 	}

// 	if filter.Country != nil {
// 		if account.Country == 0 {
// 			return false
// 		}

// 		if account.Country != *filter.Country {
// 			return false
// 		}
// 	}

// 	if filter.City != nil {
// 		if account.City == 0 {
// 			return false
// 		}

// 		if account.City != *filter.City {
// 			return false
// 		}
// 	}

// 	if filter.BirthYear != nil {
// 		if account.Birth < *filter.BirthYearGte || account.Birth > *filter.BirthYearLte {
// 			return false
// 		}
// 	}

// 	if filter.Interests != nil {
// 		if len(account.Interests) == 0 {
// 			return false
// 		}
// 		exists := false
// 		for _, interest := range account.Interests {
// 			if interest == *filter.Interests {
// 				exists = true
// 				break
// 			}
// 		}
// 		if !exists {
// 			return false
// 		}
// 	}

// 	if filter.Likes != nil {
// 		if len(account.Likes) == 0 {
// 			return false
// 		}
// 		exists := false
// 		for _, like := range account.Likes {
// 			if like.ID == ID(*filter.Likes) {
// 				exists = true
// 				break
// 			}
// 		}
// 		if !exists {
// 			return false
// 		}
// 	}

// 	if filter.JoinedYear != nil {
// 		if account.Joined < *filter.JoinedYearGte || account.Joined > *filter.JoinedYearLte {
// 			return false
// 		}
// 	}

// 	return true
// }
