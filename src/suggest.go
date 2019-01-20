package main

import (
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

type SuggestFilter struct {
	Country *Country
	City    *City
}

type Suggest struct {
	store *Store
	dicts *Dicts
	// Account     *Account
	QueryID     string
	ExpectEmpty bool
	Filter      SuggestFilter
	Limit       int
}

func NewSuggest(store *Store, dicts *Dicts) *Suggest {
	return &Suggest{
		store: store,
		dicts: dicts,
	}
}

func (suggest *Suggest) Parse(query string) error {
	values, err := url.ParseQuery(query)
	if err != nil {
		return err
	}

	for param, paramValues := range values {
		if len(paramValues) != 1 || paramValues[0] == "" {
			return errors.New("Invalid suggest param value")
		}

		err := suggest.ParseParam(param, paramValues[0])
		if err != nil {
			return err
		}
	}
	if suggest.Limit == 0 {
		return errors.New("Limit should be specified")
	}
	if suggest.Limit > 20 {
		return errors.New("Limit should be less or equal 20")
	}
	return nil
}

func (suggest *Suggest) ParseParam(param string, value string) error {
	switch param {
	case "country":
		country, err := suggest.dicts.GetCountry(value)
		if err != nil {
			suggest.ExpectEmpty = true
			return nil
		}
		suggest.Filter.Country = &country
	case "city":
		city, err := suggest.dicts.GetCity(value)
		if err != nil {
			suggest.ExpectEmpty = true
			return nil
		}
		suggest.Filter.City = &city
	case "limit":
		ui64, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return errors.New("Invalid limit value")
		}
		suggest.Limit = int(ui64)
	case "query_id":
		suggest.QueryID = value
	default:
		return errors.New("Unknown suggest param")
	}

	return nil
}

func Similarity(me *Account, account *Account) float64 {
	similarity := float64(0)

	if len(me.Likes) == 0 {
		return similarity
	}

	i := 0
	j := 0

	for i < len(me.Likes) && j < len(account.Likes) {
		myTsTotal := uint64(me.Likes[i].Ts)
		myTsCount := uint64(1)
		i++

		for i < len(me.Likes) && me.Likes[i-1].ID == me.Likes[i].ID {
			myTsTotal += uint64(me.Likes[i].Ts)
			myTsCount++
			i++
		}

		anotherTsTotal := uint64(0)
		anotherTsCount := uint64(0)
		for j < len(account.Likes) && account.Likes[j].ID >= me.Likes[i-1].ID {
			if account.Likes[j].ID == me.Likes[i-1].ID {
				anotherTsTotal += uint64(account.Likes[j].Ts)
				anotherTsCount++
			}
			j++
		}

		if myTsCount > 0 && anotherTsCount > 0 {
			myTs := float64(myTsTotal / myTsCount)
			anotherTs := float64(anotherTsTotal / anotherTsCount)

			dTs := myTs - anotherTs
			if dTs < 0 {
				dTs = -dTs
			}
			if dTs == 0 {
				similarity++
			} else {
				similarity += 1 / dTs
			}
		}
	}

	// taked := make(map[ID]bool)
	// for _, like := range me.Likes {
	// 	if _, ok := taked[like.ID]; ok {
	// 		continue
	// 	}
	// 	taked[like.ID] = true

	// 	myTsTotal := uint64(like.Ts)
	// 	myTsCount := uint64(1)
	// 	for _, myLike := range me.Likes {
	// 		if myLike.ID == like.ID {
	// 			myTsTotal += uint64(myLike.Ts)
	// 			myTsCount++
	// 		}
	// 	}
	// 	anotherTsTotal := uint64(0)
	// 	anotherTsCount := uint64(0)
	// 	for _, anotherLike := range account.Likes {
	// 		if anotherLike.ID == like.ID {
	// 			anotherTsTotal += uint64(anotherLike.Ts)
	// 			anotherTsCount++
	// 		}
	// 	}
	// 	if myTsCount > 0 && anotherTsCount > 0 {
	// 		myTs := float64(myTsTotal / myTsCount)
	// 		anotherTs := float64(anotherTsTotal / anotherTsCount)

	// 		dTs := myTs - anotherTs
	// 		if dTs < 0 {
	// 			dTs = -dTs
	// 		}
	// 		if dTs == 0 {
	// 			similarity++
	// 		} else {
	// 			similarity += 1 / dTs
	// 		}
	// 	}
	// }
	return similarity
}

func (suggest *Suggest) Sex() bool       { return false }
func (suggest *Suggest) Status() bool    { return true }
func (suggest *Suggest) Fname() bool     { return true }
func (suggest *Suggest) Sname() bool     { return true }
func (suggest *Suggest) Phone() bool     { return false }
func (suggest *Suggest) Country() bool   { return false }
func (suggest *Suggest) City() bool      { return false }
func (suggest *Suggest) Birth() bool     { return false }
func (suggest *Suggest) Premium() bool   { return false }
func (suggest *Suggest) Interests() bool { return false }
