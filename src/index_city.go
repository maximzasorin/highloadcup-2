package main

import "sort"

type IndexCity struct {
	cities map[City]IDS
}

func NewIndexCity() *IndexCity {
	return &IndexCity{
		cities: make(map[City]IDS),
	}
}

func (indexCity *IndexCity) Add(city City, ID uint32) {
	_, ok := indexCity.cities[city]
	if !ok {
		indexCity.cities[city] = make([]uint32, 1)
		indexCity.cities[city][0] = ID
		return
	}

	indexCity.cities[city] = append(indexCity.cities[city], ID)
}

func (indexCity *IndexCity) Remove(city City, ID uint32) {
	_, ok := indexCity.cities[city]
	if !ok {
		return
	}
	for i, accountID := range indexCity.cities[city] {
		if accountID == ID {
			indexCity.cities[city] = append(indexCity.cities[city][:i], indexCity.cities[city][i+1:]...)
			return
		}
	}
}

func (indexCity *IndexCity) Update(city City) {
	if city == 0 {
		for city := range indexCity.cities {
			sort.Sort(indexCity.cities[city])
		}
		return
	}

	if _, ok := indexCity.cities[city]; ok {
		sort.Sort(indexCity.cities[city])
	}
}

func (indexCity *IndexCity) Get(city City) IDS {
	if _, ok := indexCity.cities[city]; ok {
		return indexCity.cities[city]
	}
	return make(IDS, 0)
}
