package main

import (
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

type Suggest struct {
	store       *Store
	dicts       *Dicts
	Account     *Account
	QueryID     *string
	ExpectEmpty bool
	Country     *Country
	City        *City
	Limit       *uint8
}

func NewSuggest(store *Store, dicts *Dicts) *Suggest {
	return &Suggest{
		store: store,
		dicts: dicts,
	}
}

func (suggest *Suggest) Parse(accountID, query string) error {
	ui64, err := strconv.ParseUint(accountID, 10, 32)
	if err != nil {
		return errors.Wrap(err, "Invalid account ID")
	}
	suggest.Account = suggest.store.Get(uint32(ui64))

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
	if suggest.Limit == nil {
		return errors.New("Limit should be specified")
	}
	if *suggest.Limit > 20 {
		return errors.New("Limit should be less than 20")
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
		suggest.Country = &country
	case "city":
		city, err := suggest.dicts.GetCity(value)
		if err != nil {
			suggest.ExpectEmpty = true
			return nil
		}
		suggest.City = &city
	case "limit":
		ui64, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return errors.New("Invalid limit value")
		}
		limit := uint8(ui64)
		suggest.Limit = &limit
	case "query_id":
		suggest.QueryID = &value
	default:
		return errors.New("Unknown suggest param")
	}

	return nil
}
