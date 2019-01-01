package main

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

type Year uint16

type Filter struct {
	QueryID           string
	Limit             uint8
	SexEq             string
	EmailDomain       string
	EmailLt           string
	EmailGt           string
	StatusEq          string
	StatusNeq         string
	FnameEq           string
	FnameAny          []string
	FnameSet          bool
	FnameNotSet       bool
	SnameEq           string
	SnameStarts       string
	SnameSet          bool
	SnameNotSet       bool
	PhoneCode         string
	PhoneSet          bool
	PhoneNotSet       bool
	CountryEq         string
	CountrySet        bool
	CountryNotSet     bool
	CityEq            string
	CityAny           []string
	CitySet           bool
	CityNotSet        bool
	BirthLt           uint32
	BirthGt           uint32
	BirthYear         Year
	InterestsContains []string
	InterestsAny      []string
	LikesContains     []uint32
	PremiumNow        bool
	PremiumSet        bool
	PremiumNotSet     bool
}

func NewFilter() *Filter {
	return &Filter{QueryID: "<unknown>"}
}

func (filter *Filter) Parse(query string) error {
	values, err := url.ParseQuery(query)
	if err != nil {
		return err
	}

	for param, paramValues := range values {
		if len(paramValues) != 1 || paramValues[0] == "" {
			return errors.New("Invalid filter param value")
		}

		err := filter.ParseParam(param, paramValues[0])
		if err != nil {
			return err
		}
	}

	return nil
}

func (filter *Filter) ParseParam(param string, value string) error {
	switch param {
	case "sex_eq":
		filter.SexEq = value
	case "email_domain":
		filter.EmailDomain = value
	case "email_lt":
		filter.EmailLt = value
	case "email_gt":
		filter.EmailGt = value
	case "status_eq":
		filter.StatusEq = value
	case "status_neq":
		filter.StatusNeq = value
	case "fname_eq":
		filter.FnameEq = value
	case "fname_any":
		filter.FnameAny = strings.Split(value, ",")
	case "fname_null":
		filter.FnameSet = value == "1"
		filter.FnameNotSet = !filter.FnameSet
	case "sname_eq":
		filter.FnameEq = value
	case "sname_starts":
		filter.SnameStarts = value
	case "sname_null":
		filter.SnameSet = value == "1"
		filter.SnameNotSet = !filter.SnameSet
	case "phone_code":
		filter.PhoneCode = value
	case "phone_null":
		filter.PhoneSet = value == "1"
		filter.PhoneNotSet = !filter.PhoneSet
	case "country_eq":
		filter.CountryEq = value
	case "country_null":
		filter.CountrySet = value == "1"
		filter.CountryNotSet = !filter.CountrySet
	case "city_eq":
		filter.CityEq = value
	case "city_any":
		filter.CityAny = strings.Split(value, ",")
	case "city_null":
		filter.CitySet = value == "1"
		filter.CityNotSet = !filter.CitySet
	case "birth_lt":
		ts, err := parseTimestamp(value)
		if err != nil {
			return err
		}
		filter.BirthLt = ts
	case "birth_gt":
		ts, err := parseTimestamp(value)
		if err != nil {
			return err
		}
		filter.BirthGt = ts
	case "birth_year":
		ui64, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return err
		}
		filter.BirthYear = Year(ui64)
	case "interests_contains":
		filter.InterestsContains = strings.Split(value, ",")
	case "interests_any":
		filter.InterestsAny = strings.Split(value, ",")
	case "likes_contains":
		likes := strings.Split(value, ",")
		for _, like := range likes {
			ui64, err := strconv.ParseUint(like, 10, 32)
			if err != nil {
				return err
			}
			filter.LikesContains = append(filter.LikesContains, uint32(ui64))
		}
	case "premium_now":
		filter.PremiumNow = true
	case "premium_null":
		filter.PremiumSet = value == "1"
		filter.PremiumNotSet = !filter.PremiumSet
	case "limit":
		ui64, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return errors.New("Invalid limit value")
		}
		filter.Limit = uint8(ui64)
	case "query_id":
		filter.QueryID = value
	default:
		return errors.New("Unknown filter param")
	}

	return nil
}

func parseTimestamp(timestamp string) (uint32, error) {
	ui64, err := strconv.ParseUint(timestamp, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(ui64), nil
}
