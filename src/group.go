package main

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

type GroupKey byte
type GroupHash uint64

const (
	GroupSex GroupKey = 1 << iota
	GroupStatus
	GroupInterests
	GroupCountry
	GroupCity
	GroupBirth
	GroupJoined
	GroupLikes
)

// https://github.com/MailRuChamps/hlcupdocs/issues/119#issuecomment-450162555
type GroupFilter struct {
	ExpectEmpty   bool
	NoFilter      bool
	Sex           *byte
	Status        *byte
	Country       *Country
	City          *City
	BirthYear     *Year
	BirthYearGte  *int64
	BirthYearLte  *int64
	Interests     *Interest // one interest
	Likes         *uint32   // account id
	JoinedYear    *Year
	JoinedYearGte *uint32
	JoinedYearLte *uint32
}

type Group struct {
	parser *Parser
	dicts  *Dicts

	QueryID  *string
	Limit    *uint8
	OrderAsc *bool

	// Filter
	Filter     GroupFilter
	FilterBits GroupKey

	// Group
	Keys     []GroupKey
	KeysBits GroupKey

	AllBits GroupKey

	// // Group
	// Keys struct {
	// 	Sex       bool
	// 	Status    bool
	// 	Interests bool
	// 	Country   bool
	// 	City      bool
	// }
}

func NewGroup(parser *Parser, dicts *Dicts) *Group {
	return &Group{
		parser: parser,
		dicts:  dicts,
	}
}

func (group *Group) Parse(query string) error {
	values, err := url.ParseQuery(query)
	if err != nil {
		return err
	}

	for param, paramValues := range values {
		if len(paramValues) != 1 || paramValues[0] == "" {
			return errors.New("Invalid group param value")
		}

		err := group.ParseParam(param, paramValues[0])
		if err != nil {
			return err
		}
	}

	if group.Limit == nil {
		return errors.New("Limit should be specified")
	}

	if group.OrderAsc == nil {
		return errors.New("Order should be specified")
	}

	group.AllBits = group.KeysBits | group.FilterBits

	filter := &group.Filter
	filter.NoFilter = filter.Sex == nil &&
		filter.Status == nil &&
		filter.Country == nil &&
		filter.City == nil &&
		filter.BirthYear == nil &&
		filter.Interests == nil &&
		filter.Likes == nil &&
		filter.JoinedYear == nil

	return nil
}

func (group *Group) HasKey(key GroupKey) bool {
	return (group.KeysBits & key) > 0
}

func (group *Group) ParseParam(param string, value string) error {
	filter := &group.Filter

	switch param {
	case "keys":
		for _, key := range strings.Split(value, ",") {
			switch key {
			case "sex":
				group.Keys = append(group.Keys, GroupSex)
				group.KeysBits |= GroupSex
			case "status":
				group.Keys = append(group.Keys, GroupStatus)
				group.KeysBits |= GroupStatus
			case "interests":
				group.Keys = append(group.Keys, GroupInterests)
				group.KeysBits |= GroupInterests
			case "country":
				group.Keys = append(group.Keys, GroupCountry)
				group.KeysBits |= GroupCountry
			case "city":
				group.Keys = append(group.Keys, GroupCity)
				group.KeysBits |= GroupCity
			default:
				return errors.New("Unknown group key " + key)
			}
		}
	case "sex":
		sex, err := group.parser.ParseSex(value)
		if err != nil {
			return err
		}
		group.FilterBits |= GroupSex
		filter.Sex = &sex
	case "status":
		status, err := group.parser.ParseStatus(value)
		if err != nil {
			return err
		}
		group.FilterBits |= GroupStatus
		filter.Status = &status
	case "country":
		country, err := group.dicts.GetCountry(value)
		if err != nil {
			filter.ExpectEmpty = true
			return nil
		}
		group.FilterBits |= GroupCountry
		filter.Country = &country
	case "city":
		city, err := group.dicts.GetCity(value)
		if err != nil {
			filter.ExpectEmpty = true
			return nil
		}
		group.FilterBits |= GroupCity
		filter.City = &city
	case "birth":
		ui64, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return err
		}
		birthYear := Year(ui64)
		filter.BirthYear = &birthYear
		group.FilterBits |= GroupBirth

		birthYearGte, birthYearLte := YearToTimestamp(birthYear)
		filter.BirthYearGte = &birthYearGte
		filter.BirthYearLte = &birthYearLte
	case "interests":
		interest, err := group.dicts.GetInterest(value)
		if err != nil {
			filter.ExpectEmpty = true
			return nil
		}
		group.FilterBits |= GroupInterests
		filter.Interests = &interest
	case "likes":
		ui64, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return err
		}
		likeID := uint32(ui64)
		filter.Likes = &likeID
		group.FilterBits |= GroupLikes
	case "joined":
		ui64, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return err
		}
		joinedYear := Year(ui64)
		filter.JoinedYear = &joinedYear
		group.FilterBits |= GroupJoined

		gte64, lte64 := YearToTimestamp(joinedYear)
		joinedYearGte := uint32(gte64)
		joinedYearLte := uint32(lte64)
		filter.JoinedYearGte = &joinedYearGte
		filter.JoinedYearLte = &joinedYearLte
	case "order":
		i8, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return err
		}
		if i8 == -1 {
			orderAsc := false
			group.OrderAsc = &orderAsc
		} else if i8 == 1 {
			orderAsc := true
			group.OrderAsc = &orderAsc
		} else {
			return errors.New("Invalid order value")
		}
	case "limit":
		ui64, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return errors.New("Invalid limit value")
		}
		limit := uint8(ui64)
		group.Limit = &limit
	case "query_id":
		group.QueryID = &value
	default:
		return errors.New("Unknown group param")
	}

	return nil
}

type GroupEntry struct {
	Sex        byte     // 2
	Status     byte     // 3
	City       City     // 610
	Country    Country  // 71
	Interest   Interest // 90
	BirthYear  Year     // 27
	JoinedYear Year     // 5
	Hash       GroupHash
	Count      uint32
}

type GroupEntries struct {
	dicts   *Dicts
	keys    []GroupKey
	entries []*GroupEntry
}

func (groupEntries *GroupEntries) Get() []*GroupEntry {
	return groupEntries.entries
}

// func NewGroupEntries(dicts *Dicts, keys []GroupKey, entries []*GroupEntry) *GroupEntries {
// 	return &GroupEntries{dicts, keys, entries}
// }

func CreateGroupHash(groupEntry *GroupEntry) GroupHash {
	// var arr [8]byte
	// arr[0] = groupEntry.Sex
	// arr[1] = groupEntry.Status
	// binary.LittleEndian.PutUint16(arr[2:4], uint16(groupEntry.City))
	// arr[4] = byte(groupEntry.Country)
	// arr[5] = byte(groupEntry.Interest)
	// arr[6] = byte(groupEntry.JoinedYear - 1949)
	// arr[7] = byte(groupEntry.BirthYear - 2010)

	// return GroupHash(binary.LittleEndian.Uint64(arr[:]))

	return GroupHash(uint64(groupEntry.Sex) | uint64(groupEntry.Status)<<8 | uint64(groupEntry.City)<<16 |
		uint64(groupEntry.Country)<<32 | uint64(groupEntry.Interest)<<40 | uint64(byte(groupEntry.JoinedYear-2010))<<48 |
		uint64(byte(groupEntry.BirthYear-1949))<<56)
}

type Aggregation struct {
	group        *Group
	groupEntries map[GroupHash]*GroupEntry
}

func NewAggregation(group *Group) *Aggregation {
	return &Aggregation{
		group:        group,
		groupEntries: make(map[GroupHash]*GroupEntry),
	}
}

func (ag *Aggregation) Add(groupEntry *GroupEntry) {
	if en, ok := ag.groupEntries[groupEntry.Hash]; ok {
		en.Count += groupEntry.Count
		return
	}
	ag.groupEntries[groupEntry.Hash] = groupEntry
}

func (ag *Aggregation) Get(asc bool, limit int) []*GroupEntry {
	if asc {
		marked := make(map[GroupHash]bool)
		entries := make([]*GroupEntry, 0)
		for len(marked) < len(ag.groupEntries) && len(marked) < limit {
			var minEntry *GroupEntry
			for _, groupEntry := range ag.groupEntries {
				if _, ok := marked[groupEntry.Hash]; ok {
					continue
				}
				if minEntry == nil || ag.groupEntryLess(groupEntry, minEntry) {
					minEntry = groupEntry
				}
			}
			if minEntry == nil {
				break
			}
			entries = append(entries, minEntry)
			marked[minEntry.Hash] = true
		}
		return entries
	}

	marked := make(map[GroupHash]bool)
	entries := make([]*GroupEntry, 0)
	for len(marked) < len(ag.groupEntries) && len(marked) < limit {
		var maxEntry *GroupEntry
		for _, groupEntry := range ag.groupEntries {
			if _, ok := marked[groupEntry.Hash]; ok {
				continue
			}
			if maxEntry == nil || !ag.groupEntryLess(groupEntry, maxEntry) {
				maxEntry = groupEntry
			}
		}
		if maxEntry == nil {
			break
		}
		entries = append(entries, maxEntry)
		marked[maxEntry.Hash] = true
	}
	return entries
}

// entries := make([]*GroupEntry, 0)
// for _, groupEntry := range ag.groupEntries {
// 	entries = append(entries, groupEntry)
// }

// groupEntries := &GroupEntries{
// 	dicts:   ag.group.dicts,
// 	keys:    ag.group.Keys,
// 	entries: entries,
// }
// if asc {
// 	sort.Sort(groupEntries)
// } else {
// 	sort.Sort(sort.Reverse(groupEntries))
// }
// entries = groupEntries.Get()
// if len(entries) > limit {
// 	entries = entries[:limit]
// }

// return entries
// }

func (groupEntries *GroupEntries) Len() int { return len(groupEntries.entries) }
func (groupEntries *GroupEntries) Swap(i, j int) {
	groupEntries.entries[i], groupEntries.entries[j] = groupEntries.entries[j], groupEntries.entries[i]
}
func (groupEntries *GroupEntries) Less(i, j int) bool {
	if groupEntries.entries[i].Count < groupEntries.entries[j].Count {
		return true
	}
	if groupEntries.entries[i].Count > groupEntries.entries[j].Count {
		return false
	}
	for _, key := range groupEntries.keys {
		switch key {
		case GroupSex:
			if groupEntries.entries[i].Sex < groupEntries.entries[j].Sex {
				return true
			} else if groupEntries.entries[i].Sex > groupEntries.entries[j].Sex {
				return false
			}
		case GroupStatus:
			if groupEntries.entries[i].Status < groupEntries.entries[j].Status {
				return true
			} else if groupEntries.entries[i].Status > groupEntries.entries[j].Status {
				return false
			}
		case GroupInterests:
			if groupEntries.entries[i].Interest != 0 {
				// if groupEntries.entries[j].Interest == nil {
				// 	return true
				// }
				interestI, err := groupEntries.dicts.GetInterestString(groupEntries.entries[i].Interest)
				if err != nil {
					continue
				}
				interestJ, err := groupEntries.dicts.GetInterestString(groupEntries.entries[j].Interest)
				if err != nil {
					continue
				}
				comp := strings.Compare(interestI, interestJ)
				if comp == -1 {
					return true
				} else if comp == 1 {
					return false
				}
			}
		case GroupCountry:
			if groupEntries.entries[i].Country != 0 && groupEntries.entries[j].Country == 0 {
				return false
			}
			if groupEntries.entries[i].Country == 0 && groupEntries.entries[j].Country != 0 {
				return true
			}
			if groupEntries.entries[i].Country != 0 {
				// if groupEntries.entries[j].Country == nil {
				// 	return true
				// }
				countryI, err := groupEntries.dicts.GetCountryString(groupEntries.entries[i].Country)
				if err != nil {
					continue
				}
				countryJ, err := groupEntries.dicts.GetCountryString(groupEntries.entries[j].Country)
				if err != nil {
					continue
				}
				comp := strings.Compare(countryI, countryJ)
				if comp == -1 {
					return true
				} else if comp == 1 {
					return false
				}
			}
		case GroupCity:
			if groupEntries.entries[i].City != 0 && groupEntries.entries[j].City == 0 {
				return false
			}
			if groupEntries.entries[i].City == 0 && groupEntries.entries[j].City != 0 {
				return true
			}
			if groupEntries.entries[i].City != 0 {
				cityI, err := groupEntries.dicts.GetCityString(groupEntries.entries[i].City)
				if err != nil {
					continue
				}
				cityJ, err := groupEntries.dicts.GetCityString(groupEntries.entries[j].City)
				if err != nil {
					continue
				}
				comp := strings.Compare(cityI, cityJ)
				if comp == -1 {
					return true
				} else if comp == 1 {
					return false
				}
			}
		}
	}
	return false
}

func (ag *Aggregation) groupEntryLess(a *GroupEntry, b *GroupEntry) bool {
	if a.Count < b.Count {
		return true
	}
	if a.Count > b.Count {
		return false
	}
	for _, key := range ag.group.Keys {
		switch key {
		case GroupSex:
			if a.Sex < b.Sex {
				return true
			} else if a.Sex > b.Sex {
				return false
			}
		case GroupStatus:
			if a.Status < b.Status {
				return true
			} else if a.Status > b.Status {
				return false
			}
		case GroupInterests:
			if a.Interest != 0 {
				// if b.Interest == nil {
				// 	return true
				// }
				interestI, err := ag.group.dicts.GetInterestString(a.Interest)
				if err != nil {
					continue
				}
				interestJ, err := ag.group.dicts.GetInterestString(b.Interest)
				if err != nil {
					continue
				}
				comp := strings.Compare(interestI, interestJ)
				if comp == -1 {
					return true
				} else if comp == 1 {
					return false
				}
			}
		case GroupCountry:
			if a.Country != 0 && b.Country == 0 {
				return false
			}
			if a.Country == 0 && b.Country != 0 {
				return true
			}
			if a.Country != 0 {
				// if b.Country == nil {
				// 	return true
				// }
				countryI, err := ag.group.dicts.GetCountryString(a.Country)
				if err != nil {
					continue
				}
				countryJ, err := ag.group.dicts.GetCountryString(b.Country)
				if err != nil {
					continue
				}
				comp := strings.Compare(countryI, countryJ)
				if comp == -1 {
					return true
				} else if comp == 1 {
					return false
				}
			}
		case GroupCity:
			if a.City != 0 && b.City == 0 {
				return false
			}
			if a.City == 0 && b.City != 0 {
				return true
			}
			if a.City != 0 {
				cityI, err := ag.group.dicts.GetCityString(a.City)
				if err != nil {
					continue
				}
				cityJ, err := ag.group.dicts.GetCityString(b.City)
				if err != nil {
					continue
				}
				comp := strings.Compare(cityI, cityJ)
				if comp == -1 {
					return true
				} else if comp == 1 {
					return false
				}
			}
		}
	}
	return false
}

// type AggregationGroup struct {
// 	Sex      byte
// 	Status   byte
// 	Country  Country
// 	City     City
// 	Interest Interest
// 	Count    uint32
// }

// func (a *AggregationGroup) equal(b *AggregationGroup) bool {
// 	return a.Sex == b.Sex && a.Status == b.Status && a.Country == b.Country && a.City == b.City && a.Interest == b.Interest
// }

// type AggregationGroups []*AggregationGroup

// type Aggregation struct {
// 	group     *Group
// 	Groups    AggregationGroups
// 	groupsMap map[[5]uint16]int
// 	search    [5]uint16
// }

// func (aggregation *Aggregation) Add(group AggregationGroup) {
// 	aggregation.search[0] = uint16(group.Sex)
// 	aggregation.search[1] = uint16(group.Status)
// 	aggregation.search[2] = uint16(group.City)
// 	aggregation.search[3] = uint16(group.Country)
// 	aggregation.search[4] = uint16(group.Interest)

// 	index, ok := aggregation.groupsMap[aggregation.search]
// 	if ok {
// 		aggregation.Groups[index].Count += group.Count
// 		return
// 	}
// 	index = len(aggregation.Groups)
// 	aggregation.Groups = append(aggregation.Groups, &group)
// 	aggregation.groupsMap[aggregation.search] = index
// }

// func (aggregation *Aggregation) Sort(asc bool) {
// 	if asc {
// 		sort.Sort(aggregation)
// 	} else {
// 		sort.Sort(sort.Reverse(aggregation))
// 	}
// }

// func (aggregation *Aggregation) Limit(limit uint8) {
// 	if len(aggregation.Groups) > int(limit) {
// 		aggregation.Groups = aggregation.Groups[:limit]
// 	}
// }

// func (a *Aggregation) Len() int { return len(a.Groups) }
// func (a *Aggregation) Swap(i, j int) {
// 	a.Groups[i], a.Groups[j] = a.Groups[j], a.Groups[i]
// }
// func (a *Aggregation) Less(i, j int) bool {
// 	if a.Groups[i].Count < a.Groups[j].Count {
// 		return true
// 	}
// 	if a.Groups[i].Count > a.Groups[j].Count {
// 		return false
// 	}
// 	for _, key := range a.group.Keys {
// 		switch key {
// 		case GroupSex:
// 			if a.Groups[i].Sex < a.Groups[j].Sex {
// 				return true
// 			} else if a.Groups[i].Sex > a.Groups[j].Sex {
// 				return false
// 			}
// 		case GroupStatus:
// 			if a.Groups[i].Status < a.Groups[j].Status {
// 				return true
// 			} else if a.Groups[i].Status > a.Groups[j].Status {
// 				return false
// 			}
// 		case GroupInterests:
// 			if a.Groups[i].Interest != 0 {
// 				// if a.Groups[j].Interest == nil {
// 				// 	return true
// 				// }
// 				interestI, err := a.group.dicts.GetInterestString(a.Groups[i].Interest)
// 				if err != nil {
// 					continue
// 				}
// 				interestJ, err := a.group.dicts.GetInterestString(a.Groups[j].Interest)
// 				if err != nil {
// 					continue
// 				}
// 				comp := strings.Compare(interestI, interestJ)
// 				if comp == -1 {
// 					return true
// 				} else if comp == 1 {
// 					return false
// 				}
// 			}
// 		case GroupCountry:
// 			if a.Groups[i].Country != 0 && a.Groups[j].Country == 0 {
// 				return false
// 			}
// 			if a.Groups[i].Country == 0 && a.Groups[j].Country != 0 {
// 				return true
// 			}
// 			if a.Groups[i].Country != 0 {
// 				// if a.Groups[j].Country == nil {
// 				// 	return true
// 				// }
// 				countryI, err := a.group.dicts.GetCountryString(a.Groups[i].Country)
// 				if err != nil {
// 					continue
// 				}
// 				countryJ, err := a.group.dicts.GetCountryString(a.Groups[j].Country)
// 				if err != nil {
// 					continue
// 				}
// 				comp := strings.Compare(countryI, countryJ)
// 				if comp == -1 {
// 					return true
// 				} else if comp == 1 {
// 					return false
// 				}
// 			}
// 		case GroupCity:
// 			if a.Groups[i].City != 0 && a.Groups[j].City == 0 {
// 				return false
// 			}
// 			if a.Groups[i].City == 0 && a.Groups[j].City != 0 {
// 				return true
// 			}
// 			if a.Groups[i].City != 0 {
// 				cityI, err := a.group.dicts.GetCityString(a.Groups[i].City)
// 				if err != nil {
// 					continue
// 				}
// 				cityJ, err := a.group.dicts.GetCityString(a.Groups[j].City)
// 				if err != nil {
// 					continue
// 				}
// 				comp := strings.Compare(cityI, cityJ)
// 				if comp == -1 {
// 					return true
// 				} else if comp == 1 {
// 					return false
// 				}
// 			}
// 		}
// 	}
// 	return false
// }
