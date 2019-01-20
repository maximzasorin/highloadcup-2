package main

import (
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

type RecommendFilter struct {
	Country *Country
	City    *City
}

type Recommend struct {
	store *Store
	dicts *Dicts
	// Account     *Account
	QueryID     string
	ExpectEmpty bool
	Filter      RecommendFilter
	Limit       int
}

func NewRecommend(store *Store, dicts *Dicts) *Recommend {
	return &Recommend{
		store: store,
		dicts: dicts,
	}
}

func (recommend *Recommend) Parse(query string) error {
	values, err := url.ParseQuery(query)
	if err != nil {
		return err
	}

	for param, paramValues := range values {
		if len(paramValues) != 1 || paramValues[0] == "" {
			return errors.New("Invalid recommend param value")
		}

		err := recommend.ParseParam(param, paramValues[0])
		if err != nil {
			return err
		}
	}
	if recommend.Limit == 0 {
		return errors.New("Limit should be specified")
	}
	if recommend.Limit > 20 {
		return errors.New("Limit should be less or equal 20")
	}
	return nil
}

func (recommend *Recommend) ParseParam(param string, value string) error {
	switch param {
	case "country":
		country, err := recommend.dicts.GetCountry(value)
		if err != nil {
			recommend.ExpectEmpty = true
			return nil
		}
		recommend.Filter.Country = &country
	case "city":
		city, err := recommend.dicts.GetCity(value)
		if err != nil {
			recommend.ExpectEmpty = true
			return nil
		}
		recommend.Filter.City = &city
	case "limit":
		ui64, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return errors.New("Invalid limit value")
		}
		recommend.Limit = int(ui64)
	case "query_id":
		recommend.QueryID = value
	default:
		return errors.New("Unknown recommend param")
	}

	return nil
}

func Compability(me *Account, somebody *Account) uint64 {
	compability := uint64(0)
	commonInts := uint64(0)
	for _, meInterest := range me.Interests {
		for _, somebodyInterest := range somebody.Interests {
			if meInterest == somebodyInterest {
				commonInts++
				break
			}
		}
	}
	if commonInts == 0 {
		return compability
	}

	if somebody.Status == StatusSingle {
		compability |= 3 << 40
	} else if somebody.Status == StatusComplicated {
		compability |= 2 << 40
	} else {
		compability |= 1 << 40
	}

	compability |= commonInts << 32

	diff := int64(me.Birth - somebody.Birth)
	if diff < 0 {
		diff = -diff
	}

	compability |= (1<<32 - 1) - uint64(uint32(diff))

	return compability
}

func (recommend *Recommend) Sex() bool       { return false }
func (recommend *Recommend) Status() bool    { return true }
func (recommend *Recommend) Fname() bool     { return true }
func (recommend *Recommend) Sname() bool     { return true }
func (recommend *Recommend) Phone() bool     { return false }
func (recommend *Recommend) Country() bool   { return false }
func (recommend *Recommend) City() bool      { return false }
func (recommend *Recommend) Birth() bool     { return true }
func (recommend *Recommend) Premium() bool   { return true }
func (recommend *Recommend) Interests() bool { return false }
