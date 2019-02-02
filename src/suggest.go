package main

import (
	"net/url"
	"strconv"
	"sync"

	"github.com/pkg/errors"
)

type SuggestFilter struct {
	Country Country
	City    City
}

type Suggest struct {
	store *Store
	dicts *Dicts

	// queryID     string
	expectEmpty bool
	limit       int
	Filter      SuggestFilter
}

var suggestsPool = sync.Pool{
	New: func() interface{} {
		return &Suggest{}
	},
}

func BorrowSuggest(store *Store, dicts *Dicts) *Suggest {
	r := suggestsPool.Get().(*Suggest)
	r.Reset()
	r.store = store
	r.dicts = dicts
	return r
}

func NewSuggests(store *Store, dicts *Dicts) *Suggest {
	return &Suggest{
		store: store,
		dicts: dicts,
	}
}

func (suggest *Suggest) Release() {
	suggestsPool.Put(suggest)
}

func (suggest *Suggest) Reset() {
	// suggest.queryID = ""
	suggest.expectEmpty = false
	suggest.limit = 0
	suggest.Filter.Country = 0
	suggest.Filter.City = 0
}

func NewSuggest(store *Store, dicts *Dicts) *Suggest {
	return &Suggest{
		store: store,
		dicts: dicts,
	}
}

func (suggest *Suggest) ExpectEmpty() bool {
	return suggest.expectEmpty
}

func (suggest *Suggest) Limit() int {
	return suggest.limit
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
	if suggest.limit == 0 {
		return errors.New("Limit should be specified")
	}
	if suggest.limit > 20 {
		return errors.New("Limit should be less or equal 20")
	}
	return nil
}

func (suggest *Suggest) ParseParam(param string, value string) error {
	switch param {
	case "country":
		country, err := suggest.dicts.GetCountry(value)
		if err != nil {
			suggest.expectEmpty = true
			return nil
		}
		suggest.Filter.Country = country
	case "city":
		city, err := suggest.dicts.GetCity(value)
		if err != nil {
			suggest.expectEmpty = true
			return nil
		}
		suggest.Filter.City = city
	case "limit":
		ui64, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return errors.New("Invalid limit value")
		}
		suggest.limit = int(ui64)
	case "query_id":
		// suggest.queryID = value
	default:
		return errors.New("Unknown suggest param")
	}

	return nil
}

func (store *Store) Similarity(me *Account, account *Account) float64 {
	similarity := float64(0)

	meLikes := store.index.Liker.Find(me.ID)
	accountLikes := store.index.Liker.Find(account.ID)

	if len(meLikes) == 0 {
		return similarity
	}

	i := 0
	j := 0

	for i < len(meLikes) && j < len(accountLikes) {
		myTsTotal := uint64(meLikes[i].Ts)
		myTsCount := uint64(1)
		i++

		for i < len(meLikes) && meLikes[i-1].ID == meLikes[i].ID {
			myTsTotal += uint64(meLikes[i].Ts)
			myTsCount++
			i++
		}

		anotherTsTotal := uint64(0)
		anotherTsCount := uint64(0)
		for j < len(accountLikes) && accountLikes[j].ID >= meLikes[i-1].ID {
			if accountLikes[j].ID == meLikes[i-1].ID {
				anotherTsTotal += uint64(accountLikes[j].Ts)
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
