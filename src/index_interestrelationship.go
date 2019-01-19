package main

import (
	"sync"
)

type IndexInterestRelationship struct {
	rwLock           sync.RWMutex
	interests        map[Interest]*IndexID
	interestsCity    map[Interest]*IndexCity
	interestsCountry map[Interest]*IndexCountry
}

func NewIndexInterestRelationship() *IndexInterestRelationship {
	return &IndexInterestRelationship{
		interests:        make(map[Interest]*IndexID),
		interestsCity:    make(map[Interest]*IndexCity),
		interestsCountry: make(map[Interest]*IndexCountry),
	}
}

func (index *IndexInterestRelationship) Add(interest Interest, city City, country Country, id ID) {
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

func (index *IndexInterestRelationship) Append(interest Interest, city City, country Country, id ID) {
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

// func (index *IndexInterestRelationship) Update(interest Interest) {
// 	index.rwLock.Lock()
// 	_, ok := index.interests[interest]
// 	if !ok {
// 		index.rwLock.Unlock()
// 		return
// 	}
// 	index.interests[interest].Update()
// 	index.interestsCity[interest].UpdateAll()
// 	index.interestsCountry[interest].UpdateAll()
// 	index.rwLock.Unlock()
// }

func (index *IndexInterestRelationship) UpdateAll() {
	index.rwLock.Lock()
	for interest := range index.interests {
		index.interests[interest].Update()
		index.interestsCity[interest].UpdateAll()
		index.interestsCountry[interest].UpdateAll()
	}
	index.rwLock.Unlock()
}

func (index *IndexInterestRelationship) AddCountry(interest Interest, country Country, id ID) {
	index.rwLock.RLock()
	index.interestsCountry[interest].Add(country, id)
	index.rwLock.RUnlock()
}

func (index *IndexInterestRelationship) RemoveCountry(interest Interest, country Country, id ID) {
	index.rwLock.RLock()
	index.interestsCountry[interest].Remove(country, id)
	index.rwLock.RUnlock()
}

func (index *IndexInterestRelationship) AddCity(interest Interest, city City, id ID) {
	index.rwLock.RLock()
	index.interestsCity[interest].Add(city, id)
	index.rwLock.RUnlock()
}

func (index *IndexInterestRelationship) RemoveCity(interest Interest, city City, id ID) {
	index.rwLock.RLock()
	index.interestsCity[interest].Remove(city, id)
	index.rwLock.RUnlock()
}

func (index *IndexInterestRelationship) Remove(interest Interest, city City, country Country, id ID) {
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

func (index *IndexInterestRelationship) Find(interest Interest) IDS {
	index.rwLock.RLock()
	if _, ok := index.interests[interest]; ok {
		ids := index.interests[interest].FindAll()
		index.rwLock.RUnlock()
		return ids
	}
	index.rwLock.RUnlock()
	return make(IDS, 0)
}

func (index *IndexInterestRelationship) FindByCity(interest Interest, city City) IDS {
	index.rwLock.RLock()
	if _, ok := index.interestsCity[interest]; ok {
		ids := index.interestsCity[interest].Find(city)
		index.rwLock.RUnlock()
		return ids
	}
	index.rwLock.RUnlock()
	return make(IDS, 0)
}

func (index *IndexInterestRelationship) FindByCountry(interest Interest, country Country) IDS {
	index.rwLock.RLock()
	if _, ok := index.interestsCountry[interest]; ok {
		ids := index.interestsCountry[interest].Find(country)
		index.rwLock.RUnlock()
		return ids
	}
	index.rwLock.RUnlock()
	return make(IDS, 0)
}

func (index *IndexInterestRelationship) Len() int {
	index.rwLock.RLock()
	interestsLen := len(index.interests)
	index.rwLock.RUnlock()
	return interestsLen
}
