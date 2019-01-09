package main

import (
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

type GroupKey byte

const (
	GroupBySex GroupKey = iota + 1
	GroupByStatus
	GroupByInterests
	GroupByCountry
	GroupByCity
)

type Group struct {
	parser *Parser
	dicts  *Dicts

	QueryID  *string
	Limit    *uint8
	OrderAsc *bool

	// Filter
	// https://github.com/MailRuChamps/hlcupdocs/issues/119#issuecomment-450162555
	Filter struct {
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

	// Group
	Keys []GroupKey

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
	for _, k := range group.Keys {
		if key == k {
			return true
		}
	}
	return false
}

func (group *Group) ParseParam(param string, value string) error {
	filter := &group.Filter

	switch param {
	case "keys":
		for _, key := range strings.Split(value, ",") {
			switch key {
			case "sex":
				group.Keys = append(group.Keys, GroupBySex)
			case "status":
				group.Keys = append(group.Keys, GroupByStatus)
			case "interests":
				group.Keys = append(group.Keys, GroupByInterests)
			case "country":
				group.Keys = append(group.Keys, GroupByCountry)
			case "city":
				group.Keys = append(group.Keys, GroupByCity)
			default:
				return errors.New("Unknown group key " + key)
			}
		}
	case "sex":
		sex, err := group.parser.ParseSex(value)
		if err != nil {
			return err
		}
		filter.Sex = &sex
	case "status":
		status, err := group.parser.ParseStatus(value)
		if err != nil {
			return err
		}
		filter.Status = &status
	case "country":
		country, err := group.dicts.GetCountry(value)
		if err != nil {
			filter.ExpectEmpty = true
			return nil
		}
		filter.Country = &country
	case "city":
		city, err := group.dicts.GetCity(value)
		if err != nil {
			filter.ExpectEmpty = true
			return nil
		}
		filter.City = &city
	case "birth":
		ui64, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return err
		}
		birthYear := Year(ui64)
		filter.BirthYear = &birthYear

		birthYearGte, birthYearLte := YearToTimestamp(birthYear)
		filter.BirthYearGte = &birthYearGte
		filter.BirthYearLte = &birthYearLte
	case "interests":
		interest, err := group.dicts.GetInterest(value)
		if err != nil {
			filter.ExpectEmpty = true
			return nil
		}
		filter.Interests = &interest
	case "likes":
		ui64, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return err
		}
		likeID := uint32(ui64)
		filter.Likes = &likeID
	case "joined":
		ui64, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return err
		}
		joinedYear := Year(ui64)
		filter.JoinedYear = &joinedYear

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

type AggregationGroup struct {
	Sex      *byte
	Status   *byte
	Country  *Country
	City     *City
	Interest *Interest
	Count    uint32
}

func (a *AggregationGroup) equal(b *AggregationGroup) bool {
	return ((a.Sex == nil && b.Sex == nil) || (a.Sex != nil && b.Sex != nil && *a.Sex == *b.Sex)) &&
		((a.Status == nil && b.Status == nil) || (a.Status != nil && b.Status != nil && *a.Status == *b.Status)) &&
		((a.Country == nil && b.Country == nil) || (a.Country != nil && b.Country != nil && *a.Country == *b.Country)) &&
		((a.City == nil && b.City == nil) || (a.City != nil && b.City != nil && *a.City == *b.City)) &&
		((a.Interest == nil && b.Interest == nil) || (a.Interest != nil && b.Interest != nil && *a.Interest == *b.Interest))
}

type AggregationGroups []*AggregationGroup

type Aggregation struct {
	group  *Group
	Groups AggregationGroups
}

func (aggregation *Aggregation) Add(group AggregationGroup) {
	for _, existsGroup := range aggregation.Groups {
		if existsGroup.equal(&group) {
			existsGroup.Count++
			return
		}
	}
	group.Count = 1
	aggregation.Groups = append(aggregation.Groups, &group)
}

func (aggregation *Aggregation) Sort(asc bool) {
	if asc {
		sort.Sort(aggregation)
	} else {
		sort.Sort(sort.Reverse(aggregation))
	}
}

func (aggregation *Aggregation) Limit(limit uint8) {
	if len(aggregation.Groups) > int(limit) {
		aggregation.Groups = aggregation.Groups[:limit]
	}
}

func (a *Aggregation) Len() int { return len(a.Groups) }
func (a *Aggregation) Swap(i, j int) {
	a.Groups[i], a.Groups[j] = a.Groups[j], a.Groups[i]
}
func (a *Aggregation) Less(i, j int) bool {
	if a.Groups[i].Count < a.Groups[j].Count {
		return true
	}
	if a.Groups[i].Count > a.Groups[j].Count {
		return false
	}
	for _, key := range a.group.Keys {
		switch key {
		case GroupBySex:
			if *a.Groups[i].Sex < *a.Groups[j].Sex {
				return true
			} else if *a.Groups[i].Sex > *a.Groups[j].Sex {
				return false
			}
		case GroupByStatus:
			if *a.Groups[i].Status < *a.Groups[j].Status {
				return true
			} else if *a.Groups[i].Status > *a.Groups[j].Status {
				return false
			}
		case GroupByInterests:
			if a.Groups[i].Interest != nil {
				// if a.Groups[j].Interest == nil {
				// 	return true
				// }
				interestI, err := a.group.dicts.GetInterestString(*a.Groups[i].Interest)
				if err != nil {
					continue
				}
				interestJ, err := a.group.dicts.GetInterestString(*a.Groups[j].Interest)
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
		case GroupByCountry:
			if a.Groups[i].Country != nil && a.Groups[j].Country == nil {
				return false
			}
			if a.Groups[i].Country == nil && a.Groups[j].Country != nil {
				return true
			}
			if a.Groups[i].Country != nil {
				// if a.Groups[j].Country == nil {
				// 	return true
				// }
				countryI, err := a.group.dicts.GetCountryString(*a.Groups[i].Country)
				if err != nil {
					continue
				}
				countryJ, err := a.group.dicts.GetCountryString(*a.Groups[j].Country)
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
		case GroupByCity:
			if a.Groups[i].City != nil && a.Groups[j].City == nil {
				return false
			}
			if a.Groups[i].City == nil && a.Groups[j].City != nil {
				return true
			}
			if a.Groups[i].City != nil {
				cityI, err := a.group.dicts.GetCityString(*a.Groups[i].City)
				if err != nil {
					continue
				}
				cityJ, err := a.group.dicts.GetCityString(*a.Groups[j].City)
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
