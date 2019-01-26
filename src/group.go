package main

import (
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type (
	GroupKey  byte
	GroupHash uint64
)

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

const (
	GroupSexMask       GroupHash = 255
	GroupStatusMask    GroupHash = 255 << 8
	GroupCityMask      GroupHash = (256*256 - 1) << 16
	GroupCountryMask   GroupHash = 255 << 32
	GroupInterestsMask GroupHash = 255 << 40
	GroupJoinedMask    GroupHash = 255 << 48
	GroupBirthMask     GroupHash = 255 << 56
)

// https://github.com/MailRuChamps/hlcupdocs/issues/119#issuecomment-450162555
type GroupFilter struct {
	ExpectEmpty   bool
	NoFilter      bool
	Sex           byte
	Status        byte
	Country       Country
	City          City
	BirthYear     Year
	BirthYearGte  int64
	BirthYearLte  int64
	Interests     Interest // one interest
	Likes         ID       // account id
	JoinedYear    Year
	JoinedYearGte uint32
	JoinedYearLte uint32
}

type Group struct {
	parser *Parser
	dicts  *Dicts

	QueryID   string
	Limit     uint8
	OrderAsc  bool
	OrderDesc bool

	// Filter
	Filter     GroupFilter
	FilterHash GroupHash

	// Group
	Keys     []GroupKey
	KeysHash GroupHash

	Entry *GroupEntry
}

func NewGroup(parser *Parser, dicts *Dicts) *Group {
	return &Group{
		parser: parser,
		dicts:  dicts,
		Entry:  NewGroupEntry(0),
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

	if group.KeysHash == 0 {
		return errors.New("Keys field should be specified")
	}

	if group.Limit == 0 {
		return errors.New("Limit should be specified")
	}

	if !group.OrderAsc && !group.OrderDesc {
		return errors.New("Order should be specified")
	}

	group.Filter.NoFilter = group.Filter.Sex == 0 &&
		group.Filter.Status == 0 &&
		group.Filter.Country == 0 &&
		group.Filter.City == 0 &&
		group.Filter.BirthYear == 0 &&
		group.Filter.Interests == 0 &&
		group.Filter.Likes == 0 &&
		group.Filter.JoinedYear == 0

	return nil
}

func (group *Group) HasKey(hash GroupHash) bool {
	return (group.KeysHash & hash) > 0
}

func (group *Group) ParseParam(param string, value string) error {
	filter := &group.Filter

	switch param {
	case "keys":
		for _, key := range strings.Split(value, ",") {
			switch key {
			case "sex":
				group.Keys = append(group.Keys, GroupSex)
				group.KeysHash |= GroupSexMask
			case "status":
				group.Keys = append(group.Keys, GroupStatus)
				group.KeysHash |= GroupStatusMask
			case "interests":
				group.Keys = append(group.Keys, GroupInterests)
				group.KeysHash |= GroupInterestsMask
			case "country":
				group.Keys = append(group.Keys, GroupCountry)
				group.KeysHash |= GroupCountryMask
			case "city":
				group.Keys = append(group.Keys, GroupCity)
				group.KeysHash |= GroupCityMask
			default:
				return errors.New("Unknown group key " + key)
			}
		}
	case "sex":
		sex, err := group.parser.ParseSex(value)
		if err != nil {
			return err
		}
		group.FilterHash |= GroupSexMask
		group.Entry.SetSex(sex)
		filter.Sex = sex
	case "status":
		status, err := group.parser.ParseStatus(value)
		if err != nil {
			return err
		}
		group.FilterHash |= GroupStatusMask
		filter.Status = status
		group.Entry.SetStatus(status)
	case "country":
		country, err := group.dicts.GetCountry(value)
		if err != nil {
			filter.ExpectEmpty = true
			return nil
		}
		group.FilterHash |= GroupCountryMask
		group.Entry.SetCountry(country)
		filter.Country = country
	case "city":
		city, err := group.dicts.GetCity(value)
		if err != nil {
			filter.ExpectEmpty = true
			return nil
		}
		group.FilterHash |= GroupCityMask
		group.Entry.SetCity(city)
		filter.City = city
	case "birth":
		ui64, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return err
		}
		filter.BirthYear = Year(ui64)
		group.FilterHash |= GroupBirthMask
		group.Entry.SetBirth(filter.BirthYear)

		birthYearGte, birthYearLte := YearToTimestamp(filter.BirthYear)
		filter.BirthYearGte = birthYearGte
		filter.BirthYearLte = birthYearLte
	case "interests":
		interest, err := group.dicts.GetInterest(value)
		if err != nil {
			filter.ExpectEmpty = true
			return nil
		}
		filter.Interests = interest
		group.FilterHash |= GroupInterestsMask
		group.Entry.SetInterest(filter.Interests)
	case "likes":
		ui64, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return err
		}
		filter.Likes = ID(ui64)
	case "joined":
		ui64, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return err
		}
		filter.JoinedYear = Year(ui64)
		group.FilterHash |= GroupJoinedMask
		group.Entry.SetJoined(filter.JoinedYear)

		gte64, lte64 := YearToTimestamp(filter.JoinedYear)
		filter.JoinedYearGte = uint32(gte64)
		filter.JoinedYearLte = uint32(lte64)
	case "order":
		i8, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return err
		}
		if i8 == -1 {
			group.OrderDesc = true
		} else if i8 == 1 {
			group.OrderAsc = true
		} else {
			return errors.New("Invalid order value")
		}
	case "limit":
		ui64, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return errors.New("Invalid limit value")
		}
		group.Limit = uint8(ui64)
	case "query_id":
		group.QueryID = value
	default:
		return errors.New("Unknown group param")
	}

	return nil
}

type GroupEntry struct {
	// Sex      byte     // 2
	// Status   byte     // 3
	// City     City     // 610
	// Country  Country  // 71
	// Interest Interest // 90
	// Birth    Year     // 27
	// Joined   Year     // 5
	Hash  GroupHash
	Count uint32
}

func NewGroupEntry(hash GroupHash) *GroupEntry {
	return &GroupEntry{Hash: hash, Count: 1}
}

func (entry *GroupEntry) Reset() {
	entry.Hash = 0
	entry.Count = 1
}

func (entry *GroupEntry) GetSex() byte {
	return byte(entry.Hash)
}

func (entry *GroupEntry) SetSex(sex byte) {
	entry.Hash = entry.Hash&(^GroupSexMask) | GroupHash(sex)
}

func (entry *GroupEntry) GetStatus() byte {
	return byte(entry.Hash >> 8)
}

func (entry *GroupEntry) SetStatus(status byte) {
	entry.Hash = entry.Hash&(^GroupStatusMask) | GroupHash(status)<<8
}

func (entry *GroupEntry) GetCity() City {
	return City(entry.Hash >> 16)
}

func (entry *GroupEntry) SetCity(city City) {
	entry.Hash = entry.Hash&(^GroupCityMask) | GroupHash(city)<<16
}

func (entry *GroupEntry) GetCountry() Country {
	return Country(entry.Hash >> 32)
}

func (entry *GroupEntry) SetCountry(country Country) {
	entry.Hash = entry.Hash&(^GroupCountryMask) | GroupHash(country)<<32
}

func (entry *GroupEntry) GetInterest() Interest {
	return Interest(entry.Hash >> 40)
}

func (entry *GroupEntry) SetInterest(interest Interest) {
	entry.Hash = entry.Hash&(^GroupInterestsMask) | GroupHash(interest)<<40
}

func (entry *GroupEntry) GetJoined() Year {
	return Year(byte(entry.Hash>>48)) + 2010
}

func (entry *GroupEntry) SetJoined(joined Year) {
	entry.Hash = entry.Hash&(^GroupJoinedMask) | GroupHash(byte(joined-2010))<<48
}

func (entry *GroupEntry) GetBirth() Year {
	return Year(byte(entry.Hash>>56)) + 1949
}

func (entry *GroupEntry) SetBirth(birth Year) {
	entry.Hash = entry.Hash&(^GroupBirthMask) | GroupHash(byte(birth-1949))<<56
}

func (entry *GroupEntry) GetHash() GroupHash {
	return entry.Hash
}

type Aggregation struct {
	dicts     *Dicts
	groupMask GroupHash
	entries   []*GroupEntry
	rwLock    sync.RWMutex
}

func NewAggregation(dicts *Dicts, groupMask GroupHash) *Aggregation {
	return &Aggregation{
		dicts:     dicts,
		groupMask: groupMask,
		entries:   make([]*GroupEntry, 0),
	}
}

func (ag *Aggregation) Append(hash GroupHash) {
	ag.rwLock.Lock()
	// ag.rwLock.RLock()
	for _, entry := range ag.entries {
		if entry.Hash == hash&ag.groupMask {
			// ag.rwLock.RUnlock()
			entry.Count++
			ag.rwLock.Unlock()
			return
		}
	}
	// ag.rwLock.RUnlock()
	// ag.rwLock.Lock()
	ag.entries = append(ag.entries, NewGroupEntry(hash&ag.groupMask))
	ag.rwLock.Unlock()
}

func (ag *Aggregation) Add(hash GroupHash) {
	// ag.rwLock.RLock()
	ag.rwLock.Lock()
	index := 0
	founded := false
	for i, entry := range ag.entries {
		if entry.Hash == hash&ag.groupMask {
			// ag.rwLock.RUnlock()
			entry.Count++
			founded = true
			index = i
			break
		}
	}
	if !founded {
		// ag.rwLock.RUnlock()
		// ag.rwLock.Lock()
		ag.entries = append([]*GroupEntry{NewGroupEntry(hash & ag.groupMask)}, ag.entries...)
		// ag.rwLock.Unlock()
	}
	// ag.rwLock.RLock()
	if len(ag.entries) <= 1 {
		// ag.rwLock.RUnlock()
		ag.rwLock.Unlock()
		return
	}
	// ag.rwLock.RUnlock()
	// ag.rwLock.Lock()
	for i := index; i < len(ag.entries)-1 && !ag.gLess(ag.entries[i], ag.entries[i+1]); i++ {
		ag.entries[i], ag.entries[i+1] = ag.entries[i+1], ag.entries[i]
	}
	ag.rwLock.Unlock()
}

func (ag *Aggregation) Sub(hash GroupHash) {
	// ag.rwLock.RLock()
	ag.rwLock.Lock()
	index := 0
	for i, entry := range ag.entries {
		if entry.Hash == hash&ag.groupMask {
			// ag.rwLock.RUnlock()
			entry.Count--
			if entry.Count == 0 {
				// ag.rwLock.Lock()
				ag.entries = append(ag.entries[:i], ag.entries[i+1:]...)
				ag.rwLock.Unlock()
				return
			}
			// ag.rwLock.RLock()
			index = i
			break
		}
	}
	if len(ag.entries) <= 1 {
		// ag.rwLock.RUnlock()
		ag.rwLock.Unlock()
		return
	}
	// ag.rwLock.RUnlock()
	// ag.rwLock.Lock()
	for i := index; i > 0 && ag.gLess(ag.entries[i], ag.entries[i-1]); i-- {
		ag.entries[i], ag.entries[i-1] = ag.entries[i-1], ag.entries[i]
	}
	ag.rwLock.Unlock()
}

func (ag *Aggregation) Update() {
	ag.rwLock.Lock()
	sort.Sort(ag)
	ag.rwLock.Unlock()
}

func (ag *Aggregation) Get() []*GroupEntry {
	return ag.entries
}

func (ag *Aggregation) Len() int { return len(ag.entries) }
func (ag *Aggregation) Swap(i, j int) {
	ag.entries[i], ag.entries[j] = ag.entries[j], ag.entries[i]
}
func (ag *Aggregation) Less(i, j int) bool {
	return ag.gLess(ag.entries[i], ag.entries[j])
}

func (ag *Aggregation) gLess(a, b *GroupEntry) bool {
	if a.Count < b.Count {
		return true
	}
	if a.Count > b.Count {
		return false
	}
	if ag.groupMask&GroupCityMask > 0 {
		if a.GetCity() != 0 && b.GetCity() == 0 {
			return false
		}
		if a.GetCity() == 0 && b.GetCity() != 0 {
			return true
		}
		if a.GetCity() != 0 {
			cityI, _ := ag.dicts.GetCityString(a.GetCity())
			cityJ, _ := ag.dicts.GetCityString(b.GetCity())
			comp := strings.Compare(cityI, cityJ)
			if comp == -1 {
				return true
			} else if comp == 1 {
				return false
			}
		}
	}
	if ag.groupMask&GroupCountryMask > 0 {
		if a.GetCountry() != 0 && b.GetCountry() == 0 {
			return false
		}
		if a.GetCountry() == 0 && b.GetCountry() != 0 {
			return true
		}
		if a.GetCountry() != 0 {
			countryI, _ := ag.dicts.GetCountryString(a.GetCountry())
			countryJ, _ := ag.dicts.GetCountryString(b.GetCountry())
			comp := strings.Compare(countryI, countryJ)
			if comp == -1 {
				return true
			} else if comp == 1 {
				return false
			}
		}
	}
	if ag.groupMask&GroupInterestsMask > 0 {
		if a.GetInterest() != 0 {
			interestI, _ := ag.dicts.GetInterestString(a.GetInterest())
			interestJ, _ := ag.dicts.GetInterestString(b.GetInterest())
			comp := strings.Compare(interestI, interestJ)
			if comp == -1 {
				return true
			} else if comp == 1 {
				return false
			}
		}
	}
	if ag.groupMask&GroupStatusMask > 0 {
		if a.GetStatus() < b.GetStatus() {
			return true
		} else if a.GetStatus() > b.GetStatus() {
			return false
		}
	}
	if ag.groupMask&GroupSexMask > 0 {
		if a.GetSex() < b.GetSex() {
			return true
		} else if a.GetSex() > b.GetSex() {
			return false
		}
	}
	return false
}

func CreateHashFromAccount(account *Account) GroupHash {
	var hash GroupHash
	hash.SetSex(account.Sex)
	hash.SetStatus(account.Status)
	hash.SetCity(account.City)
	hash.SetCountry(account.Country)
	// hash.SetInterest(account.Interest)
	hash.SetJoined(timestampToYear(int64(account.Joined)))
	hash.SetBirth(timestampToYear(int64(account.Birth)))
	return hash
}

func (hash *GroupHash) GetSex() byte {
	return byte(*hash)
}

func (hash *GroupHash) SetSex(sex byte) {
	*hash = *hash&(^GroupSexMask) | GroupHash(sex)
}

func (hash *GroupHash) GetStatus() byte {
	return byte(*hash >> 8)
}

func (hash *GroupHash) SetStatus(status byte) {
	*hash = *hash&(^GroupStatusMask) | GroupHash(status)<<8
}

func (hash *GroupHash) GetCity() City {
	return City(*hash >> 16)
}

func (hash *GroupHash) SetCity(city City) {
	*hash = *hash&(^GroupCityMask) | GroupHash(city)<<16
}

func (hash *GroupHash) GetCountry() Country {
	return Country(*hash >> 32)
}

func (hash *GroupHash) SetCountry(country Country) {
	*hash = *hash&(^GroupCountryMask) | GroupHash(country)<<32
}

func (hash *GroupHash) GetInterest() Interest {
	return Interest(*hash >> 40)
}

func (hash *GroupHash) SetInterest(interest Interest) {
	*hash = *hash&(^GroupInterestsMask) | GroupHash(interest)<<40
}

func (hash *GroupHash) GetJoined() Year {
	return Year(byte(*hash>>48)) + 2010
}

func (hash *GroupHash) SetJoined(joined Year) {
	*hash = *hash&(^GroupJoinedMask) | GroupHash(byte(joined-2010))<<48
}

func (hash *GroupHash) GetBirth() Year {
	return Year(byte(*hash>>56)) + 1949
}

func (hash *GroupHash) SetBirth(birth Year) {
	*hash = *hash&(^GroupBirthMask) | GroupHash(byte(birth-1949))<<56
}
