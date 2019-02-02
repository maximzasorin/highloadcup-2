package main

import (
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type Year uint16

type Filter struct {
	parser *Parser
	dicts  *Dicts

	// queryID     string
	noFilter    bool
	expectEmpty bool
	limit       int

	sex       bool
	email     bool
	status    bool
	fname     bool
	sname     bool
	phone     bool
	country   bool
	city      bool
	birth     bool
	interests bool
	likes     bool
	premium   bool

	SexEq             byte
	EmailDomain       string
	EmailLt           string
	EmailGt           string
	StatusEq          byte
	StatusNeq         byte
	FnameEq           Fname
	FnameAny          []Fname
	FnameNull         bool
	FnameNullSet      bool
	SnameEq           Sname
	SnameStarts       string
	SnameNull         bool
	SnameNullSet      bool
	PhoneCode         uint16
	PhoneNull         bool
	PhoneNullSet      bool
	CountryEq         Country
	CountryNull       bool
	CountryNullSet    bool
	CityEq            City
	CityAny           []City
	CityNull          bool
	CityNullSet       bool
	BirthLt           int64
	BirthGt           int64
	BirthYear         Year
	BirthYearGte      int64
	BirthYearLte      int64
	InterestsContains []Interest
	InterestsAny      []Interest
	LikesContains     []uint32
	PremiumNow        bool
	PremiumNull       bool
	PremiumNullSet    bool
}

var filtersPool = sync.Pool{
	New: func() interface{} {
		return &Filter{
			FnameAny:          make([]Fname, 0, 16),
			CityAny:           make([]City, 0, 16),
			InterestsContains: make([]Interest, 0, 16),
			InterestsAny:      make([]Interest, 0, 16),
		}
	},
}

// type FiltersPool struct {
// 	parser *Parser
// 	dicts  *Dicts
// 	pool   chan *Filter
// }

// func NewFiltersPool(parser *Parser, dicts *Dicts) *FiltersPool {
// 	return &FiltersPool{
// 		parser: parser,
// 		dicts:  dicts,
// 		pool:   make(chan *Filter, 100),
// 	}
// }

func BorrowFilter(parser *Parser, dicts *Dicts) *Filter {
	f := filtersPool.Get().(*Filter)
	f.Reset()
	f.parser = parser
	f.dicts = dicts
	return f
}

func NewFilter(parser *Parser, dicts *Dicts) *Filter {
	return &Filter{
		parser: parser,
		dicts:  dicts,
	}
}

func (filter *Filter) Release() {
	filtersPool.Put(filter)
}

func (filter *Filter) Reset() {
	// filter.queryID = ""
	filter.noFilter = true
	filter.expectEmpty = false
	filter.limit = 0

	filter.sex = false
	filter.email = false
	filter.status = false
	filter.fname = false
	filter.sname = false
	filter.phone = false
	filter.country = false
	filter.city = false
	filter.birth = false
	filter.interests = false
	filter.likes = false
	filter.premium = false

	filter.SexEq = 0
	filter.EmailDomain = ""
	filter.EmailLt = ""
	filter.EmailGt = ""
	filter.StatusEq = 0
	filter.StatusNeq = 0
	filter.FnameEq = 0
	filter.FnameAny = filter.FnameAny[:0]
	filter.FnameNull = false
	filter.FnameNullSet = false
	filter.SnameEq = 0
	filter.SnameStarts = ""
	filter.SnameNull = false
	filter.SnameNullSet = false
	filter.PhoneCode = 0
	filter.PhoneNull = false
	filter.PhoneNullSet = false
	filter.CountryEq = 0
	filter.CountryNull = false
	filter.CountryNullSet = false
	filter.CityEq = 0
	filter.CityAny = filter.CityAny[:0]
	filter.CityNull = false
	filter.CityNullSet = false
	filter.BirthLt = 0
	filter.BirthGt = 0
	filter.BirthYear = 0
	filter.BirthYearGte = 0
	filter.BirthYearLte = 0
	filter.InterestsContains = filter.InterestsContains[:0]
	filter.InterestsAny = filter.InterestsAny[:0]
	filter.LikesContains = filter.LikesContains[:0]
	filter.PremiumNow = false
	filter.PremiumNull = false
	filter.PremiumNullSet = false
}

func (filter *Filter) ExpectEmpty() bool {
	return filter.expectEmpty
}

func (filter *Filter) NoFilter() bool {
	return filter.noFilter
}

func (filter *Filter) Limit() int {
	return filter.limit
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
			return errors.Wrap(err, "Invalid filter param")
		}
	}

	if filter.limit == 0 {
		return errors.New("Limit should be specified")
	}

	filter.noFilter = !filter.sex &&
		!filter.email &&
		!filter.status &&
		!filter.fname &&
		!filter.sname &&
		!filter.phone &&
		!filter.country &&
		!filter.city &&
		!filter.birth &&
		!filter.interests &&
		!filter.likes &&
		!filter.premium

	if filter.StatusEq != 0 && filter.StatusNeq != 0 && filter.StatusEq == filter.StatusNeq {
		filter.expectEmpty = true
	}

	if filter.CountryEq != 0 {
		if filter.CityEq != 0 {
			if !filter.dicts.ExistsCityInCountry(filter.CityEq, filter.CountryEq) {
				filter.expectEmpty = true
			}
		}
		if len(filter.CityAny) > 0 {
			anyExists := false
			for _, city := range filter.CityAny {
				if filter.dicts.ExistsCityInCountry(city, filter.CountryEq) {
					anyExists = true
				}
			}
			if !anyExists {
				filter.expectEmpty = true
			}
		}
	}
	return nil
}

func (filter *Filter) ParseParam(param string, value string) error {
	switch param {
	case "sex_eq":
		sex, err := filter.parser.ParseSex(value)
		if err != nil {
			return err
		}
		filter.SexEq = sex
		filter.sex = true
	case "email_domain":
		filter.EmailDomain = value
		filter.email = true
	case "email_lt":
		filter.EmailLt = value
		filter.email = true
	case "email_gt":
		filter.EmailGt = value
		filter.email = true
	case "status_eq":
		status, err := filter.parser.ParseStatus(value)
		if err != nil {
			return err
		}
		filter.StatusEq = status
		filter.status = true
	case "status_neq":
		status, err := filter.parser.ParseStatus(value)
		if err != nil {
			return err
		}
		filter.StatusNeq = status
		filter.status = true
	case "fname_eq":
		fname, err := filter.dicts.GetFname(value)
		if err != nil {
			filter.expectEmpty = true
			return nil
		}
		filter.FnameEq = fname
		filter.fname = true
	case "fname_any":
		fnameAny := make([]Fname, 0)
		for _, fnameStr := range strings.Split(value, ",") {
			fname, err := filter.dicts.GetFname(fnameStr)
			if err != nil {
				continue
			}
			fnameAny = append(fnameAny, fname)
		}
		if len(fnameAny) == 0 {
			filter.expectEmpty = true
			return nil
		}
		filter.FnameAny = fnameAny
		filter.fname = true
	case "fname_null":
		filter.FnameNull = value == "1"
		filter.FnameNullSet = true
		filter.fname = true
	case "sname_eq":
		sname, err := filter.dicts.GetSname(value)
		if err != nil {
			filter.expectEmpty = true
			return nil
		}
		filter.SnameEq = sname
		filter.sname = true
	case "sname_starts":
		filter.SnameStarts = value
		filter.sname = true
	case "sname_null":
		filter.SnameNull = value == "1"
		filter.SnameNullSet = true
		filter.sname = true
	case "phone_code":
		ui64, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return err
		}
		filter.PhoneCode = uint16(ui64)
		filter.phone = true
	case "phone_null":
		filter.PhoneNull = value == "1"
		filter.PhoneNullSet = true
		filter.phone = true
	case "country_eq":
		country, err := filter.dicts.GetCountry(value)
		if err != nil {
			filter.expectEmpty = true
			return nil
		}
		filter.CountryEq = country
		filter.country = true
	case "country_null":
		filter.CountryNull = value == "1"
		filter.CountryNullSet = true
		filter.country = true
	case "city_eq":
		city, err := filter.dicts.GetCity(value)
		if err != nil {
			filter.expectEmpty = true
			return nil
		}
		filter.CityEq = city
		filter.city = true
	case "city_any":
		cityStrs := strings.Split(value, ",")
		if len(cityStrs) == 0 {
			return nil
		}
		for _, cityStr := range cityStrs {
			city, err := filter.dicts.GetCity(cityStr)
			if err != nil {
				continue
			}
			filter.CityAny = append(filter.CityAny, city)
		}
		if len(filter.CityAny) == 0 {
			filter.expectEmpty = true
			return nil
		}
		filter.city = true
	case "city_null":
		filter.CityNull = value == "1"
		filter.CityNullSet = true
		filter.city = true
	case "birth_lt":
		ts, err := parseTimestamp(value)
		if err != nil {
			return err
		}
		filter.BirthLt = ts
		filter.birth = true
	case "birth_gt":
		ts, err := parseTimestamp(value)
		if err != nil {
			return err
		}
		filter.BirthGt = ts
		filter.birth = true
	case "birth_year":
		ui64, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return err
		}
		filter.BirthYear = Year(ui64)
		filter.birth = true

		birthYearGte, birthYearLte := YearToTimestamp(filter.BirthYear)
		filter.BirthYearGte = birthYearGte
		filter.BirthYearLte = birthYearLte
	case "interests_contains":
		interestsContainsStr := strings.Split(value, ",")
		for _, interestStr := range interestsContainsStr {
			interest, err := filter.dicts.GetInterest(interestStr)
			if err != nil {
				continue
			}
			filter.InterestsContains = append(filter.InterestsContains, interest)
		}
		if len(filter.InterestsContains) != len(interestsContainsStr) {
			filter.expectEmpty = true
			return nil
		}
		filter.interests = true
	case "interests_any":
		interestsAnyStr := strings.Split(value, ",")
		for _, interestStr := range interestsAnyStr {
			interest, err := filter.dicts.GetInterest(interestStr)
			if err != nil {
				continue
			}
			filter.InterestsAny = append(filter.InterestsAny, interest)
		}
		if len(filter.InterestsAny) == 0 {
			filter.expectEmpty = true
			return nil
		}
		filter.interests = true
	case "likes_contains":
		likes := strings.Split(value, ",")
		for _, like := range likes {
			ui64, err := strconv.ParseUint(like, 10, 32)
			if err != nil {
				return err
			}
			filter.LikesContains = append(filter.LikesContains, uint32(ui64))
		}
		filter.likes = true
	case "premium_now":
		filter.PremiumNow = true
		filter.premium = true
	case "premium_null":
		filter.PremiumNull = value == "1"
		filter.PremiumNullSet = true
		filter.premium = true
	case "limit":
		ui64, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return errors.New("Invalid limit value")
		}
		filter.limit = int(ui64)
	case "query_id":
		// filter.queryID = value
	default:
		return errors.New("Unknown filter param")
	}

	return nil
}

func YearToTimestamp(year Year) (gte, lte int64) {
	gte = time.Date(int(year), 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	lte = time.Date(int(year)+1, 1, 1, 0, 0, 0, 0, time.UTC).Unix() - 1

	return
}

func parseTimestamp(timestamp string) (int64, error) {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return 0, err
	}
	return ts, nil
}
func timestampToYear(timestamp int64) Year {
	return Year(time.Unix(timestamp, 0).Year())
}

func (filter *Filter) Sex() bool       { return filter.sex }
func (filter *Filter) Status() bool    { return filter.status }
func (filter *Filter) Fname() bool     { return filter.fname }
func (filter *Filter) Sname() bool     { return filter.sname }
func (filter *Filter) Phone() bool     { return filter.phone }
func (filter *Filter) Country() bool   { return filter.country }
func (filter *Filter) City() bool      { return filter.city }
func (filter *Filter) Birth() bool     { return filter.birth }
func (filter *Filter) Premium() bool   { return filter.premium }
func (filter *Filter) Interests() bool { return filter.interests }
