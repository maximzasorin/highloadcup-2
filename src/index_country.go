package main

import "sort"

type IndexCountry struct {
	countries map[Country]IDS
}

func NewIndexCountry() *IndexCountry {
	return &IndexCountry{
		countries: make(map[Country]IDS),
	}
}

func (indexCountry *IndexCountry) Add(country Country, ID uint32) {
	_, ok := indexCountry.countries[country]
	if !ok {
		indexCountry.countries[country] = make(IDS, 1)
		indexCountry.countries[country][0] = ID
		return
	}

	indexCountry.countries[country] = append(indexCountry.countries[country], ID)
}

func (indexCountry *IndexCountry) Remove(country Country, ID uint32) {
	_, ok := indexCountry.countries[country]
	if !ok {
		return
	}
	for i, accountID := range indexCountry.countries[country] {
		if accountID == ID {
			indexCountry.countries[country] = append(indexCountry.countries[country][:i], indexCountry.countries[country][i+1:]...)
			return
		}
	}
}

func (indexCountry *IndexCountry) Update(country Country) {
	if country == 0 {
		for country := range indexCountry.countries {
			sort.Sort(indexCountry.countries[country])
		}
		return
	}

	if _, ok := indexCountry.countries[country]; ok {
		sort.Sort(indexCountry.countries[country])
	}
}

func (indexCountry *IndexCountry) Get(country Country) IDS {
	if _, ok := indexCountry.countries[country]; ok {
		return indexCountry.countries[country]
	}
	return make(IDS, 0)
}
