package main

func (store *Store) GroupAll(group *Group) *Aggregation {
	filter := &group.Filter
	keys := &group.Keys

	aggregation := Aggregation{group: group}

	if filter.ExpectEmpty {
		return &aggregation
	}

	// scan all
	for _, id := range store.indexID.FindAll() {
		account := store.accounts[id]

		if !filter.NoFilter {
			if filter.Sex != nil {
				if account.Sex != *filter.Sex {
					continue
				}
			}

			if filter.Status != nil {
				if account.Status != *filter.Status {
					continue
				}
			}

			if filter.Country != nil {
				if account.Country == 0 {
					continue
				}

				if account.Country != *filter.Country {
					continue
				}
			}

			if filter.City != nil {
				if account.City == 0 {
					continue
				}

				if account.City != *filter.City {
					continue
				}
			}

			if filter.BirthYear != nil {
				if account.Birth < *filter.BirthYearGte || account.Birth > *filter.BirthYearLte {
					continue
				}
			}

			if filter.Interests != nil {
				if len(account.Interests) == 0 {
					continue
				}
				exists := false
				for _, interest := range account.Interests {
					if interest == *filter.Interests {
						exists = true
						break
					}
				}
				if !exists {
					continue
				}
			}

			if filter.Likes != nil {
				if len(account.Likes) == 0 {
					continue
				}
				exists := false
				for _, like := range account.Likes {
					if like.ID == ID(*filter.Likes) {
						exists = true
						break
					}
				}
				if !exists {
					continue
				}
			}

			if filter.JoinedYear != nil {
				if account.Joined < *filter.JoinedYearGte || account.Joined > *filter.JoinedYearLte {
					continue
				}
			}
		}

		// group
		ag := AggregationGroup{}

		for _, key := range *keys {
			switch key {
			case GroupBySex:
				ag.Sex = &account.Sex
			case GroupByStatus:
				ag.Status = &account.Status
			case GroupByCountry:
				if account.Country != 0 {
					ag.Country = &account.Country
				}
			case GroupByCity:
				if account.City != 0 {
					ag.City = &account.City
				}
			}
		}

		if group.HasKey(GroupByInterests) {
			for _, i := range account.Interests {
				interest := i
				ag.Interest = &interest
				aggregation.Add(ag)
			}
		} else {
			aggregation.Add(ag)
		}
	}

	aggregation.Sort(*group.OrderAsc)
	aggregation.Limit(*group.Limit)

	return &aggregation
}
