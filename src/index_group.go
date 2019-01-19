package main

type IndexGroup struct {
	entries map[GroupKey]map[GroupHash]*GroupEntry
}

// type IndexGroupEntry struct {
// 	Sex      byte     // 2
// 	Status   byte     // 3
// 	City     City     // 610
// 	Country  Country  // 71
// 	Interest Interest // 90

// 	// filters
// 	BirthYear  Year // 27
// 	JoinedYear Year // 5
// 	Hash       GroupHash
// 	Count      uint32
// }

func NewIndexGroup() *IndexGroup {
	indexGroup := &IndexGroup{entries: make(map[GroupKey]map[GroupHash]*GroupEntry)}
	indexGroup.Init()
	return indexGroup
}

func (index *IndexGroup) Init() {
	// groups := [][]string{
	// 	// 1
	// 	[]string{"sex"},
	// 	[]string{"status"},
	// 	[]string{"city"},
	// 	[]string{"country"},
	// 	[]string{"interests"},
	// 	// "Birth",
	// 	// "Joined",

	// 	// 2
	// 	[]string{"sex", "status"},
	// 	[]string{"sex", "city"},
	// 	[]string{"sex", "country"},
	// 	[]string{"sex", "interests"},
	// 	[]string{"sex", "birth"},
	// 	[]string{"sex", "joined"},
	// 	[]string{"status", "city"},
	// 	[]string{"status", "country"},
	// 	[]string{"status", "interests"},
	// 	[]string{"status", "birth"},
	// 	[]string{"status", "joined"},
	// 	[]string{"city", "country"},
	// 	// []string{"city", "interests"},
	// 	[]string{"city", "birth"},
	// 	[]string{"city", "joined"},
	// 	[]string{"country", "interests"},
	// 	[]string{"country", "birth"},
	// 	[]string{"country", "joined"},
	// 	[]string{"interests", "birth"},
	// 	[]string{"interests", "joined"},
	// 	[]string{"birth", "joined"},

	// 	// 3
	// 	[]string{"sex", "status", "city"},
	// 	[]string{"sex", "status", "country"},
	// 	[]string{"sex", "status", "interests"},
	// 	[]string{"sex", "status", "birth"},
	// 	[]string{"sex", "status", "joined"},
	// 	[]string{"sex", "city", "country"},
	// 	// []string{"sex", "city", "interests"},
	// 	[]string{"sex", "city", "birth"},
	// 	[]string{"sex", "city", "joined"},
	// 	// []string{"sex", "country", "interests"},
	// 	[]string{"sex", "country", "birth"},
	// 	[]string{"sex", "country", "joined"},
	// 	[]string{"sex", "interests", "birth"},
	// 	[]string{"sex", "interests", "joined"},
	// 	[]string{"sex", "birth", "joined"},
	// 	[]string{"status", "city", "country"},
	// 	// []string{"status", "city", "interests"},
	// 	[]string{"status", "city", "birth"},
	// 	[]string{"status", "city", "joined"},
	// 	// []string{"status", "country", "interests"},
	// 	[]string{"status", "country", "birth"},
	// 	[]string{"status", "country", "joined"},
	// 	[]string{"status", "interests", "birth"},
	// 	[]string{"status", "interests", "joined"},
	// 	[]string{"status", "birth", "joined"},
	// 	[]string{"city", "country", "interest"},
	// 	[]string{"city", "country", "birth"},
	// 	[]string{"city", "country", "joined"},
	// 	// []string{"city", "interests", "birth"},
	// 	// []string{"city", "interests", "joined"},
	// 	[]string{"city", "birth", "joined"},
	// 	// []string{"country", "interests", "birth"},
	// 	// []string{"country", "interests", "joined"},
	// 	[]string{"country", "birth", "joined"},
	// 	[]string{"interests", "birth", "joined"},
	// }

	for _, group := range []GroupKey{
		// 1
		GroupSex,
		GroupStatus,
		GroupCity,
		GroupCountry,
		GroupInterests,

		// 2
		GroupSex | GroupStatus,
		GroupSex | GroupCity,
		GroupSex | GroupCountry,
		GroupSex | GroupInterests,
		GroupSex | GroupBirth,
		GroupSex | GroupJoined,
		GroupStatus | GroupCity,
		GroupStatus | GroupCountry,
		GroupStatus | GroupInterests,
		GroupStatus | GroupBirth,
		GroupStatus | GroupJoined,
		GroupCity | GroupCountry,
		GroupCity | GroupInterests,
		GroupCity | GroupBirth,
		GroupCity | GroupJoined,
		GroupCountry | GroupInterests,
		GroupCountry | GroupBirth,
		GroupCountry | GroupJoined,
		GroupInterests | GroupBirth,
		GroupInterests | GroupJoined,
		GroupBirth | GroupJoined,

		// 3
		GroupSex | GroupStatus | GroupCity,
		GroupSex | GroupStatus | GroupCountry,
		GroupSex | GroupStatus | GroupInterests,
		GroupSex | GroupStatus | GroupBirth,
		GroupSex | GroupStatus | GroupJoined,
		GroupSex | GroupCity | GroupCountry,
		GroupSex | GroupCity | GroupInterests,
		GroupSex | GroupCity | GroupBirth,
		GroupSex | GroupCity | GroupJoined,
		GroupSex | GroupCountry | GroupInterests,
		GroupSex | GroupCountry | GroupBirth,
		GroupSex | GroupCountry | GroupJoined,
		GroupSex | GroupInterests | GroupBirth,
		GroupSex | GroupInterests | GroupJoined,
		GroupSex | GroupBirth | GroupJoined,
		GroupStatus | GroupCity | GroupCountry,
		GroupStatus | GroupCity | GroupInterests,
		GroupStatus | GroupCity | GroupBirth,
		GroupStatus | GroupCity | GroupJoined,
		GroupStatus | GroupCountry | GroupInterests,
		GroupStatus | GroupCountry | GroupBirth,
		GroupStatus | GroupCountry | GroupJoined,
		GroupStatus | GroupInterests | GroupBirth,
		GroupStatus | GroupInterests | GroupJoined,
		GroupStatus | GroupBirth | GroupJoined,
		GroupCity | GroupCountry | GroupInterests,
		GroupCity | GroupCountry | GroupBirth,
		GroupCity | GroupCountry | GroupJoined,
		GroupCity | GroupInterests | GroupBirth,
		GroupCity | GroupInterests | GroupJoined,
		GroupCity | GroupBirth | GroupJoined,
		GroupCountry | GroupInterests | GroupBirth,
		GroupCountry | GroupInterests | GroupJoined,
		GroupCountry | GroupBirth | GroupJoined,
		GroupInterests | GroupBirth | GroupJoined,

		// 4
		GroupSex | GroupStatus | GroupCity | GroupCountry,
		// GroupSex | GroupStatus | GroupCity | GroupInterests,
		GroupSex | GroupStatus | GroupCity | GroupBirth,
		GroupSex | GroupStatus | GroupCity | GroupJoined,
		// GroupSex | GroupStatus | GroupCountry | GroupInterests,
		GroupSex | GroupStatus | GroupCountry | GroupBirth,
		GroupSex | GroupStatus | GroupCountry | GroupJoined,
		GroupSex | GroupStatus | GroupInterests | GroupBirth,
		GroupSex | GroupStatus | GroupInterests | GroupJoined,
		GroupSex | GroupStatus | GroupBirth | GroupJoined,
		// GroupSex | GroupCity | GroupCountry | GroupInterests,
		// GroupSex | GroupCity | GroupCountry | GroupInterests,
		GroupSex | GroupCity | GroupCountry | GroupBirth,
		GroupSex | GroupCity | GroupCountry | GroupJoined,
		// GroupSex | GroupCity | GroupInterests | GroupBirth,
		// GroupSex | GroupCity | GroupInterests | GroupJoined,
		GroupSex | GroupCity | GroupBirth | GroupJoined,
		// GroupSex | GroupCountry | GroupInterests | GroupBirth,
		// GroupSex | GroupCountry | GroupInterests | GroupJoined,
		GroupSex | GroupCountry | GroupBirth | GroupJoined,
		GroupSex | GroupInterests | GroupBirth | GroupJoined,
		// GroupStatus | GroupCity | GroupCountry | GroupInterests,
		GroupStatus | GroupCity | GroupCountry | GroupBirth,
		GroupStatus | GroupCity | GroupCountry | GroupJoined,
		// GroupStatus | GroupCity | GroupInterests | GroupBirth,
		// GroupStatus | GroupCity | GroupInterests | GroupJoined,
		// GroupStatus | GroupCountry | GroupInterests | GroupBirth,
		// GroupStatus | GroupCountry | GroupInterests | GroupJoined,
		GroupStatus | GroupCountry | GroupBirth | GroupJoined,
		GroupStatus | GroupInterests | GroupBirth | GroupJoined,
		// GroupCity | GroupCountry | GroupInterests | GroupBirth,
		// GroupCity | GroupCountry | GroupInterests | GroupJoined,
		GroupCity | GroupCountry | GroupBirth | GroupJoined,
		// GroupCity | GroupInterests | GroupBirth | GroupJoined,
		// GroupCountry | GroupInterests | GroupBirth | GroupJoined,
	} {
		index.entries[group] = make(map[GroupHash]*GroupEntry)
	}

	// for _, group := range groups {

	// 	for _, account := range *store.GetAll() {
	// 		birthYear := timestampToYear(int64(account.Birth))
	// 		joinedYear := timestampToYear(int64(account.Joined))
	// 		entry := &GroupEntry{Count: 1}

	// 		if group&GroupSex > 0 {
	// 			entry.Sex = account.Sex
	// 		}
	// 		if group&GroupStatus > 0 {
	// 			entry.Status = account.Status
	// 		}
	// 		if group&GroupCity > 0 {
	// 			entry.City = account.City
	// 		}
	// 		if group&GroupCountry > 0 {
	// 			entry.Country = account.Country
	// 		}
	// 		if group&GroupBirth > 0 {
	// 			entry.BirthYear = birthYear
	// 		}
	// 		if group&GroupJoined > 0 {
	// 			entry.JoinedYear = joinedYear
	// 		}
	// 		if group&GroupInterests > 0 {
	// 			for _, interest := range account.Interests {
	// 				en := &GroupEntry{
	// 					Sex:        entry.Sex,
	// 					Status:     entry.Status,
	// 					City:       entry.City,
	// 					Country:    entry.Country,
	// 					BirthYear:  entry.BirthYear,
	// 					JoinedYear: entry.JoinedYear,
	// 					Interest:   interest,
	// 					Count:      1,
	// 				}
	// 				index.Add(group, en)
	// 			}
	// 		} else {
	// 			index.Add(group, entry)
	// 		}
	// 	}
	// }
}

func (index *IndexGroup) Add(account *Account) {
	birthYear := timestampToYear(int64(account.Birth))
	joinedYear := timestampToYear(int64(account.Joined))

	for groupKey := range index.entries {
		groupEntry := &GroupEntry{Count: 1}
		if groupKey&GroupSex > 0 {
			groupEntry.Sex = account.Sex
		}
		if groupKey&GroupStatus > 0 {
			groupEntry.Status = account.Status
		}
		if groupKey&GroupCountry > 0 {
			groupEntry.Country = account.Country
		}
		if groupKey&GroupCity > 0 {
			groupEntry.City = account.City
		}
		if groupKey&GroupBirth > 0 {
			groupEntry.BirthYear = birthYear
		}
		if groupKey&GroupJoined > 0 {
			groupEntry.JoinedYear = joinedYear
		}
		if groupKey&GroupInterests > 0 {
			for _, interest := range account.Interests {
				groupEntry.Interest = interest
				groupHash := CreateGroupHash(groupEntry)
				if en, ok := index.entries[groupKey][groupHash]; ok {
					en.Count++
				} else {
					index.entries[groupKey][groupHash] = &GroupEntry{
						Sex:        groupEntry.Sex,
						Status:     groupEntry.Status,
						Country:    groupEntry.Country,
						City:       groupEntry.City,
						BirthYear:  groupEntry.BirthYear,
						JoinedYear: groupEntry.JoinedYear,
						Interest:   groupEntry.Interest,
						Count:      1,
					}
				}
			}
		} else {
			groupHash := CreateGroupHash(groupEntry)
			if en, ok := index.entries[groupKey][groupHash]; ok {
				en.Count++
			} else {
				index.entries[groupKey][groupHash] = groupEntry
			}
		}
	}
}

// func (index *IndexGroup) Add(keys GroupKey, entry *GroupEntry) {
// 	if _, ok := index.entries[keys]; !ok {
// 		index.entries[keys] = make(map[GroupHash]*GroupEntry, 0)
// 	}

// 	founded := false
// 	for _, en := range index.entries[keys] {
// 		if en.Sex == entry.Sex && en.Status == entry.Status && en.City == entry.City && en.Country == entry.Country && en.BirthYear == entry.BirthYear && en.JoinedYear == entry.JoinedYear && en.Interest == entry.Interest {
// 			founded = true
// 			en.Count++
// 			break
// 		}
// 	}

// 	if !founded {
// 		index.entries[keys] = append(index.entries[keys], entry)
// 	}
// }

func (index *IndexGroup) Get(keys GroupKey) (map[GroupHash]*GroupEntry, bool) {
	if entries, ok := index.entries[keys]; ok {
		return entries, true
	}
	return make(map[GroupHash]*GroupEntry), false
}

// founded := false
// birthYear := timestampToYear(int64(account.Birth))
// joinedYear := timestampToYear(int64(account.Joined))
// for _, entry := range index.entries {
// 	if entry.Sex != account.Sex {
// 		continue
// 	}
// 	if entry.Status != account.Status {
// 		continue
// 	}
// 	if entry.BirthYear != birthYear {
// 		continue
// 	}
// 	if entry.JoinedYear != joinedYear {
// 		continue
// 	}
// 	if entry.Country != account.Country {
// 		continue
// 	}
// 	// hasInterest := false
// 	// for _, interest := range account.Interests {
// 	// 	if entry.Interest == interest {
// 	// 		hasInterest = true
// 	// 	}
// 	// }
// 	// if !hasInterest {
// 	// 	continue
// 	// }
// 	founded = true
// 	entry.Count++
// }
// if !founded {
// 	// if len(account.Interests) > 0 {
// 	// 	for _, interest := range account.Interests {
// 	// 		index.entries = append(index.entries, IndexGroupEntry{
// 	// 			Sex:        account.Sex,
// 	// 			Status:     account.Status,
// 	// 			BirthYear:  birthYear,
// 	// 			JoinedYear: joinedYear,
// 	// 			Country:    account.Country,
// 	// 			Interest:   interest,
// 	// 		})
// 	// 	}
// 	// } else {
// 	index.entries = append(index.entries, IndexGroupEntry{
// 		Sex:        account.Sex,
// 		Status:     account.Status,
// 		BirthYear:  birthYear,
// 		JoinedYear: joinedYear,
// 		Country:    account.Country,
// 	})
// 	// }
// }

// func (index *IndexGroup) Entries() []IndexGroupEntry {
// 	return index.entries
// }

func (index *IndexGroup) Len() int {
	// index.rwLock.RLock()
	entriesLen := len(index.entries)
	// index.rwLock.RUnlock()
	return entriesLen
}
