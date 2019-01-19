package main

import (
	"sync"
)

type IndexInterestPremium struct {
	rwLock             sync.RWMutex
	interests          map[Interest]*IndexID
	interestsStatusSex map[Interest]*IndexStatusSex
	interestsCity      map[Interest]*IndexCity
	interestsCountry   map[Interest]*IndexCountry
}

func NewIndexInterestPremium() *IndexInterestPremium {
	return &IndexInterestPremium{
		interests:          make(map[Interest]*IndexID),
		interestsStatusSex: make(map[Interest]*IndexStatusSex),
		interestsCity:      make(map[Interest]*IndexCity),
		interestsCountry:   make(map[Interest]*IndexCountry),
	}
}

func (index *IndexInterestPremium) Add(interest Interest, status byte, sex byte, city City, country Country, id ID) {
	index.rwLock.RLock()
	_, ok := index.interests[interest]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.interests[interest] = NewIndexID(64)
		index.interestsCity[interest] = NewIndexCity()
		index.interestsCountry[interest] = NewIndexCountry()
		index.interestsStatusSex[interest] = NewIndexStatusSex()
		index.interestsCountry[interest] = NewIndexCountry()
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.interests[interest].Add(id)
		index.interestsStatusSex[interest].Add(status, sex, id)
		if city != 0 {
			index.interestsCity[interest].Add(city, id)
		}
		if country != 0 {
			index.interestsCountry[interest].Add(country, id)
		}
		index.rwLock.RUnlock()
		return
	}
	index.interests[interest].Add(id)
	index.interestsStatusSex[interest].Add(status, sex, id)
	if city != 0 {
		index.interestsCity[interest].Add(city, id)
	}
	if country != 0 {
		index.interestsCountry[interest].Add(country, id)
	}
	index.rwLock.RUnlock()
}

func (index *IndexInterestPremium) Append(interest Interest, status byte, sex byte, city City, country Country, id ID) {
	index.rwLock.RLock()
	_, ok := index.interests[interest]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.interests[interest] = NewIndexID(64)
		index.interestsCity[interest] = NewIndexCity()
		index.interestsCountry[interest] = NewIndexCountry()
		index.interestsStatusSex[interest] = NewIndexStatusSex()
		index.interestsCountry[interest] = NewIndexCountry()
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.interests[interest].Append(id)
		index.interestsStatusSex[interest].Append(status, sex, id)
		if city != 0 {
			index.interestsCity[interest].Append(city, id)
		}
		if country != 0 {
			index.interestsCountry[interest].Append(country, id)
		}
		index.rwLock.RUnlock()
		return
	}
	index.interests[interest].Append(id)
	index.interestsStatusSex[interest].Append(status, sex, id)
	if city != 0 {
		index.interestsCity[interest].Append(city, id)
	}
	if country != 0 {
		index.interestsCountry[interest].Append(country, id)
	}
	index.rwLock.RUnlock()
}

func (index *IndexInterestPremium) Update(interest Interest) {
	index.rwLock.Lock()
	_, ok := index.interests[interest]
	if !ok {
		index.rwLock.Unlock()
		return
	}
	index.interests[interest].Update()
	index.rwLock.Unlock()
}

func (index *IndexInterestPremium) UpdateAll() {
	index.rwLock.Lock()
	for interest := range index.interests {
		index.interests[interest].Update()
		index.interestsStatusSex[interest].UpdateAll()
		index.interestsCity[interest].UpdateAll()
		index.interestsCountry[interest].UpdateAll()
	}
	index.rwLock.Unlock()
}

func (index *IndexInterestPremium) AddCountry(interest Interest, country Country, id ID) {
	index.rwLock.RLock()
	index.interestsCountry[interest].Add(country, id)
	index.rwLock.RUnlock()
}

func (index *IndexInterestPremium) RemoveCountry(interest Interest, country Country, id ID) {
	index.rwLock.RLock()
	index.interestsCountry[interest].Remove(country, id)
	index.rwLock.RUnlock()
}

func (index *IndexInterestPremium) AddCity(interest Interest, city City, id ID) {
	index.rwLock.RLock()
	index.interestsCity[interest].Add(city, id)
	index.rwLock.RUnlock()
}

func (index *IndexInterestPremium) RemoveCity(interest Interest, city City, id ID) {
	index.rwLock.RLock()
	index.interestsCity[interest].Remove(city, id)
	index.rwLock.RUnlock()
}

func (index *IndexInterestPremium) Remove(interest Interest, status byte, sex byte, city City, country Country, id ID) {
	index.rwLock.RLock()
	_, ok := index.interests[interest]
	if !ok {
		index.rwLock.RUnlock()
		return
	}
	index.interests[interest].Remove(id)
	index.interestsStatusSex[interest].Remove(status, sex, id)
	index.interestsCity[interest].Remove(city, id)
	index.interestsCountry[interest].Remove(country, id)
	index.rwLock.RUnlock()
}

func (index *IndexInterestPremium) Find(interest Interest) IDS {
	index.rwLock.RLock()
	if _, ok := index.interests[interest]; ok {
		ids := index.interests[interest].FindAll()
		index.rwLock.RUnlock()
		return ids
	}
	index.rwLock.RUnlock()
	return make(IDS, 0)
}

func (index *IndexInterestPremium) FindByCity(interest Interest, city City) IDS {
	index.rwLock.RLock()
	if _, ok := index.interestsCity[interest]; ok {
		ids := index.interestsCity[interest].Find(city)
		index.rwLock.RUnlock()
		return ids
	}
	index.rwLock.RUnlock()
	return make(IDS, 0)
}

func (index *IndexInterestPremium) FindByCountry(interest Interest, country Country) IDS {
	index.rwLock.RLock()
	if _, ok := index.interestsCountry[interest]; ok {
		ids := index.interestsCountry[interest].Find(country)
		index.rwLock.RUnlock()
		return ids
	}
	index.rwLock.RUnlock()
	return make(IDS, 0)
}

func (index *IndexInterestPremium) FindByStatusSex(interest Interest, status byte, sex byte) IDS {
	index.rwLock.RLock()
	if _, ok := index.interestsStatusSex[interest]; ok {
		ids := index.interestsStatusSex[interest].Find(status, sex)
		index.rwLock.RUnlock()
		return ids
	}
	index.rwLock.RUnlock()
	return make(IDS, 0)
}

func (index *IndexInterestPremium) Len() int {
	index.rwLock.RLock()
	interestsLen := len(index.interests)
	index.rwLock.RUnlock()
	return interestsLen
}
