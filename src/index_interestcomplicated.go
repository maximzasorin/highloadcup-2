package main

import (
	"sync"
)

type IndexInterestComplicated struct {
	rwLock           sync.RWMutex
	interests        map[Interest]*IndexID
	interestsCity    map[Interest]*IndexCity
	interestsCountry map[Interest]*IndexCountry
}

func NewIndexInterestComplicated() *IndexInterestComplicated {
	return &IndexInterestComplicated{
		interests:        make(map[Interest]*IndexID),
		interestsCity:    make(map[Interest]*IndexCity),
		interestsCountry: make(map[Interest]*IndexCountry),
	}
}

func (index *IndexInterestComplicated) Add(interest Interest, city City, country Country, id ID) {
	index.rwLock.RLock()
	_, ok := index.interests[interest]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.interests[interest] = NewIndexID(64)
		index.interestsCity[interest] = NewIndexCity()
		index.interestsCountry[interest] = NewIndexCountry()
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.interests[interest].Add(id)
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
	if city != 0 {
		index.interestsCity[interest].Add(city, id)
	}
	if country != 0 {
		index.interestsCountry[interest].Add(country, id)
	}
	index.rwLock.RUnlock()
}

func (index *IndexInterestComplicated) Append(interest Interest, city City, country Country, id ID) {
	index.rwLock.RLock()
	_, ok := index.interests[interest]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.interests[interest] = NewIndexID(64)
		index.interestsCity[interest] = NewIndexCity()
		index.interestsCountry[interest] = NewIndexCountry()
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.interests[interest].Append(id)
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
	if city != 0 {
		index.interestsCity[interest].Append(city, id)
	}
	if country != 0 {
		index.interestsCountry[interest].Append(country, id)
	}
	index.rwLock.RUnlock()
}

// func (index *IndexInterestComplicated) Update(interest Interest) {
// 	index.rwLock.Lock()
// 	_, ok := index.interests[interest]
// 	if !ok {
// 		index.rwLock.Unlock()
// 		return
// 	}
// 	index.interests[interest].Update()
// 	index.rwLock.Unlock()
// }

func (index *IndexInterestComplicated) UpdateAll() {
	index.rwLock.Lock()
	for interest := range index.interests {
		index.interests[interest].Update()
		index.interestsCity[interest].UpdateAll()
		index.interestsCountry[interest].UpdateAll()
	}
	index.rwLock.Unlock()
}

func (index *IndexInterestComplicated) AddCountry(interest Interest, country Country, id ID) {
	index.rwLock.RLock()
	index.interestsCountry[interest].Add(country, id)
	index.rwLock.RUnlock()
}

func (index *IndexInterestComplicated) RemoveCountry(interest Interest, country Country, id ID) {
	index.rwLock.RLock()
	index.interestsCountry[interest].Remove(country, id)
	index.rwLock.RUnlock()
}

func (index *IndexInterestComplicated) AddCity(interest Interest, city City, id ID) {
	index.rwLock.RLock()
	index.interestsCity[interest].Add(city, id)
	index.rwLock.RUnlock()
}

func (index *IndexInterestComplicated) RemoveCity(interest Interest, city City, id ID) {
	index.rwLock.RLock()
	index.interestsCity[interest].Remove(city, id)
	index.rwLock.RUnlock()
}

func (index *IndexInterestComplicated) Remove(interest Interest, city City, country Country, id ID) {
	index.rwLock.RLock()
	_, ok := index.interests[interest]
	if !ok {
		index.rwLock.RUnlock()
		return
	}
	index.interests[interest].Remove(id)
	index.interestsCity[interest].Remove(city, id)
	index.interestsCountry[interest].Remove(country, id)
	index.rwLock.RUnlock()
}

func (index *IndexInterestComplicated) Find(interest Interest) IDS {
	index.rwLock.RLock()
	if _, ok := index.interests[interest]; ok {
		ids := index.interests[interest].FindAll()
		index.rwLock.RUnlock()
		return ids
	}
	index.rwLock.RUnlock()
	return make(IDS, 0)
}

func (index *IndexInterestComplicated) FindByCity(interest Interest, city City) IDS {
	index.rwLock.RLock()
	if _, ok := index.interestsCity[interest]; ok {
		ids := index.interestsCity[interest].Find(city)
		index.rwLock.RUnlock()
		return ids
	}
	index.rwLock.RUnlock()
	return make(IDS, 0)
}

func (index *IndexInterestComplicated) FindByCountry(interest Interest, country Country) IDS {
	index.rwLock.RLock()
	if _, ok := index.interestsCountry[interest]; ok {
		ids := index.interestsCountry[interest].Find(country)
		index.rwLock.RUnlock()
		return ids
	}
	index.rwLock.RUnlock()
	return make(IDS, 0)
}

func (index *IndexInterestComplicated) Len() int {
	index.rwLock.RLock()
	interestsLen := len(index.interests)
	index.rwLock.RUnlock()
	return interestsLen
}
