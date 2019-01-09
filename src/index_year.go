package main

import "sort"

type IndexYear struct {
	years map[Year]IDS
}

func NewIndexYear() *IndexYear {
	return &IndexYear{
		years: make(map[Year]IDS),
	}
}

func (indexYear *IndexYear) Add(year Year, ID uint32) {
	_, ok := indexYear.years[year]
	if !ok {
		indexYear.years[year] = make([]uint32, 1)
		indexYear.years[year][0] = ID
		return
	}

	indexYear.years[year] = append(indexYear.years[year], ID)
}

func (indexYear *IndexYear) Remove(year Year, ID uint32) {
	_, ok := indexYear.years[year]
	if !ok {
		return
	}
	for i, accountID := range indexYear.years[year] {
		if accountID == ID {
			indexYear.years[year] = append(indexYear.years[year][:i], indexYear.years[year][i+1:]...)
			return
		}
	}
}

func (indexYear *IndexYear) Update(year Year) {
	if year == 0 {
		for year := range indexYear.years {
			sort.Sort(indexYear.years[year])
		}
		return
	}

	if _, ok := indexYear.years[year]; ok {
		sort.Sort(indexYear.years[year])
	}
}

func (indexYear *IndexYear) Get(year Year) IDS {
	if _, ok := indexYear.years[year]; ok {
		return indexYear.years[year]
	}
	return make(IDS, 0)
}
