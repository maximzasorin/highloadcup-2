package main

import (
	"sync"
)

type IndexGroup struct {
	dicts        *Dicts
	filterGroups map[GroupHash][]GroupHash
	entries      map[GroupHash]map[GroupHash]map[GroupHash]*Aggregation
	rwLock       sync.RWMutex
}

func NewIndexGroup(dicts *Dicts) *IndexGroup {
	filterGroups := map[GroupHash][]GroupHash{
		0: []GroupHash{
			GroupSexMask,
			GroupStatusMask,
			GroupCityMask,
			GroupCountryMask,
			GroupInterestsMask,
			GroupCityMask | GroupSexMask,
			GroupCityMask | GroupStatusMask,
			GroupCountryMask | GroupSexMask,
			GroupCountryMask | GroupStatusMask,
		},
		GroupSexMask: []GroupHash{
			GroupCityMask,
			GroupCountryMask,
			GroupCityMask | GroupSexMask,
			GroupCityMask | GroupStatusMask,
			GroupCountryMask | GroupSexMask,
			GroupCountryMask | GroupStatusMask,
		},
		GroupStatusMask: []GroupHash{
			GroupCityMask,
			GroupCountryMask,
			GroupCityMask | GroupSexMask,
			GroupCityMask | GroupStatusMask,
			GroupCountryMask | GroupSexMask,
			GroupCountryMask | GroupStatusMask,
		},
		GroupCityMask: []GroupHash{
			GroupSexMask,
			GroupStatusMask,
			GroupInterestsMask,
		},
		GroupCountryMask: []GroupHash{
			GroupSexMask,
			GroupStatusMask,
			GroupInterestsMask,
		},
		GroupJoinedMask: []GroupHash{
			GroupSexMask,
			GroupStatusMask,
			GroupCityMask,
			GroupCountryMask,
			GroupInterestsMask,
			GroupCityMask | GroupSexMask,
			GroupCityMask | GroupStatusMask,
			GroupCountryMask | GroupSexMask,
			GroupCountryMask | GroupStatusMask,
		},
		GroupBirthMask: []GroupHash{
			GroupSexMask,
			GroupStatusMask,
			GroupCityMask,
			GroupCountryMask,
			GroupInterestsMask,
			GroupCountryMask | GroupSexMask,
			GroupCountryMask | GroupStatusMask,
			GroupCityMask | GroupSexMask,
			GroupCityMask | GroupStatusMask,
		},
		GroupInterestsMask: []GroupHash{
			GroupCityMask,
			GroupCountryMask,
			GroupCityMask | GroupSexMask,
			GroupCityMask | GroupStatusMask,
			GroupCountryMask | GroupSexMask,
			GroupCountryMask | GroupStatusMask,
		},
		GroupCountryMask | GroupJoinedMask: []GroupHash{
			GroupSexMask,
			GroupStatusMask,
			GroupInterestsMask,
		},
		GroupCountryMask | GroupBirthMask: []GroupHash{
			GroupSexMask,
			GroupStatusMask,
			GroupInterestsMask,
		},
		GroupCityMask | GroupJoinedMask: []GroupHash{
			GroupSexMask,
			GroupStatusMask,
			GroupInterestsMask,
		},
		GroupCityMask | GroupBirthMask: []GroupHash{
			GroupSexMask,
			GroupStatusMask,
			GroupInterestsMask,
		},
		GroupJoinedMask | GroupStatusMask: []GroupHash{
			GroupCityMask,
			GroupCountryMask,
			GroupCityMask | GroupSexMask,
			GroupCityMask | GroupStatusMask,
			GroupCountryMask | GroupSexMask,
			GroupCountryMask | GroupStatusMask,
		},
		GroupJoinedMask | GroupSexMask: []GroupHash{
			GroupCityMask,
			GroupCountryMask,
			GroupCityMask | GroupSexMask,
			GroupCityMask | GroupStatusMask,
			GroupCountryMask | GroupSexMask,
			GroupCountryMask | GroupStatusMask,
		},
		GroupBirthMask | GroupStatusMask: []GroupHash{
			GroupCityMask,
			GroupCountryMask,
			GroupCityMask | GroupSexMask,
			GroupCityMask | GroupStatusMask,
			GroupCountryMask | GroupSexMask,
			GroupCountryMask | GroupStatusMask,
		},
		GroupBirthMask | GroupSexMask: []GroupHash{
			GroupCityMask,
			GroupCountryMask,
			GroupCityMask | GroupSexMask,
			GroupCityMask | GroupStatusMask,
			GroupCountryMask | GroupSexMask,
			GroupCountryMask | GroupStatusMask,
		},
		GroupInterestsMask | GroupJoinedMask: []GroupHash{
			GroupCityMask,
			GroupCountryMask,
			GroupCityMask | GroupSexMask,
			GroupCityMask | GroupStatusMask,
			GroupCountryMask | GroupSexMask,
			GroupCountryMask | GroupStatusMask,
		},
		GroupInterestsMask | GroupBirthMask: []GroupHash{
			GroupCityMask,
			GroupCountryMask,
			GroupCityMask | GroupSexMask,
			GroupCityMask | GroupStatusMask,
			GroupCountryMask | GroupSexMask,
			GroupCountryMask | GroupStatusMask,
		},
	}

	entries := make(map[GroupHash]map[GroupHash]map[GroupHash]*Aggregation)

	for filter := range filterGroups {
		entries[filter] = make(map[GroupHash]map[GroupHash]*Aggregation)
		for _, group := range filterGroups[filter] {
			entries[filter][group] = make(map[GroupHash]*Aggregation)
		}
	}

	return &IndexGroup{
		dicts:        dicts,
		filterGroups: filterGroups,
		entries:      entries,
	}
}

func (index *IndexGroup) Append(account *Account) {
	accEntry := NewGroupEntry(0)
	accEntry.SetSex(account.Sex)
	accEntry.SetStatus(account.Status)
	accEntry.SetCity(account.City)
	accEntry.SetCountry(account.Country)
	accEntry.SetBirth(timestampToYear(int64(account.Birth)))
	accEntry.SetJoined(timestampToYear(int64(account.Joined)))

	for filter := range index.filterGroups {
		for _, group := range index.filterGroups[filter] {
			if filter&GroupInterestsMask > 0 || group&GroupInterestsMask > 0 {
				for _, interest := range account.Interests {
					accEntry.SetInterest(interest)
					index.appendGroup(
						filter,
						group,
						accEntry.GetHash(),
					)
				}
			} else {
				index.appendGroup(
					filter,
					group,
					accEntry.GetHash(),
				)
			}
		}
	}
}

func (index *IndexGroup) Add(account *Account) {
	accEntry := NewGroupEntry(0)
	accEntry.SetSex(account.Sex)
	accEntry.SetStatus(account.Status)
	accEntry.SetCity(account.City)
	accEntry.SetCountry(account.Country)
	accEntry.SetBirth(timestampToYear(int64(account.Birth)))
	accEntry.SetJoined(timestampToYear(int64(account.Joined)))

	for filter := range index.filterGroups {
		for _, group := range index.filterGroups[filter] {
			if filter&GroupInterestsMask > 0 || group&GroupInterestsMask > 0 {
				for _, interest := range account.Interests {
					accEntry.SetInterest(interest)
					index.addGroup(
						filter,
						group,
						accEntry.GetHash(),
					)
				}
			} else {
				index.addGroup(
					filter,
					group,
					accEntry.GetHash(),
				)
			}
		}
	}
}

func (index *IndexGroup) Sub(account *Account) {
	accEntry := NewGroupEntry(0)
	accEntry.SetSex(account.Sex)
	accEntry.SetStatus(account.Status)
	accEntry.SetCity(account.City)
	accEntry.SetCountry(account.Country)
	accEntry.SetBirth(timestampToYear(int64(account.Birth)))
	accEntry.SetJoined(timestampToYear(int64(account.Joined)))

	for filter := range index.filterGroups {
		for _, group := range index.filterGroups[filter] {
			if filter&GroupInterestsMask > 0 || group&GroupInterestsMask > 0 {
				for _, interest := range account.Interests {
					accEntry.SetInterest(interest)
					index.subGroup(
						filter,
						group,
						accEntry.GetHash(),
					)
				}
			} else {
				index.subGroup(
					filter,
					group,
					accEntry.GetHash(),
				)
			}
		}
	}
}

func (index *IndexGroup) AddHash(hash GroupHash, interests ...Interest) {
	for filter := range index.filterGroups {
		for _, group := range index.filterGroups[filter] {
			if filter&GroupInterestsMask > 0 || group&GroupInterestsMask > 0 {
				for _, interest := range interests {
					hash.SetInterest(interest)
					index.addGroup(
						filter,
						group,
						hash,
					)
				}
			} else {
				hash.SetInterest(0)
				index.addGroup(
					filter,
					group,
					hash,
				)
			}
		}
	}
}

func (index *IndexGroup) SubHash(hash GroupHash, interests ...Interest) {
	for filter := range index.filterGroups {
		for _, group := range index.filterGroups[filter] {
			if filter&GroupInterestsMask > 0 || group&GroupInterestsMask > 0 {
				for _, interest := range interests {
					hash.SetInterest(interest)
					index.subGroup(
						filter,
						group,
						hash,
					)
				}
			} else {
				hash.SetInterest(0)
				index.subGroup(
					filter,
					group,
					hash,
				)
			}
		}
	}
}

func (index *IndexGroup) Get(filter, group, filterVal GroupHash) *Aggregation {
	index.rwLock.RLock()
	if _, ok := index.entries[filter]; !ok {
		index.rwLock.RUnlock()
		return nil
	}
	if _, ok := index.entries[filter][group]; !ok {
		index.rwLock.RUnlock()
		return nil
	}
	if _, ok := index.entries[filter][group][filterVal]; !ok {
		index.rwLock.RUnlock()
		return nil
	}
	entries := index.entries[filter][group][filterVal]
	index.rwLock.RUnlock()
	return entries
}

func (index *IndexGroup) UpdateAll() {
	index.rwLock.RLock()
	for filter := range index.entries {
		for group := range index.entries[filter] {
			for filterVal := range index.entries[filter][group] {
				index.entries[filter][group][filterVal].Update()
			}
		}
	}
	index.rwLock.RUnlock()
}

func (index *IndexGroup) appendGroup(filter, group, accHash GroupHash) {
	index.rwLock.RLock()
	filterHash := accHash & filter
	if _, ok := index.entries[filter][group][filterHash]; !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.entries[filter][group][filterHash] = NewAggregation(index.dicts, group)
		index.rwLock.Unlock()
		index.rwLock.RLock()
	}
	index.entries[filter][group][filterHash].Append(accHash)
	index.rwLock.RUnlock()
}

func (index *IndexGroup) addGroup(filter, group, accHash GroupHash) {
	index.rwLock.RLock()
	filterHash := accHash & filter
	if _, ok := index.entries[filter][group][filterHash]; !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.entries[filter][group][filterHash] = NewAggregation(index.dicts, group)
		index.rwLock.Unlock()
		index.rwLock.RLock()
	}
	index.entries[filter][group][filterHash].Add(accHash)
	index.rwLock.RUnlock()
}

func (index *IndexGroup) subGroup(filter, group, accHash GroupHash) {
	index.rwLock.RLock()
	filterHash := accHash & filter
	if _, ok := index.entries[filter][group][filterHash]; !ok {
		index.rwLock.RUnlock()
		return
	}
	index.entries[filter][group][filterHash].Sub(accHash)
	index.rwLock.RUnlock()
}
