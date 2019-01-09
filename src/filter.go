package main

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type Year uint16

type Filter struct {
	parser      *Parser
	dicts       *Dicts
	QueryID     *string
	NoFilter    bool
	ExpectEmpty bool
	Fields      struct {
		Sex       bool
		Email     bool
		Status    bool
		Fname     bool
		Sname     bool
		Phone     bool
		Country   bool
		City      bool
		Birth     bool
		Interests bool
		Likes     bool
		Premium   bool

		Limit             *uint8
		SexEq             *byte
		EmailDomain       *string
		EmailLt           *string
		EmailGt           *string
		StatusEq          *byte
		StatusNeq         *byte
		FnameEq           *Fname
		FnameAny          *[]Fname
		FnameNull         *bool
		SnameEq           *Sname
		SnameStarts       *string
		SnameNull         *bool
		PhoneCode         *uint16
		PhoneNull         *bool
		CountryEq         *Country
		CountryNull       *bool
		CityEq            *City
		CityAny           *[]City
		CityNull          *bool
		BirthLt           *int64
		BirthGt           *int64
		BirthYear         *Year
		BirthYearGte      *int64
		BirthYearLte      *int64
		InterestsContains *[]Interest
		InterestsAny      *[]Interest
		LikesContains     *[]uint32
		PremiumNow        *bool
		PremiumNull       *bool
	}
}

func NewFilter(parser *Parser, dicts *Dicts) *Filter {
	return &Filter{
		parser:   parser,
		dicts:    dicts,
		NoFilter: true,
	}
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

	fields := &filter.Fields

	if fields.Limit == nil {
		return errors.New("Limit should be specified")
	}

	filter.NoFilter = !fields.Sex &&
		!fields.Email &&
		!fields.Status &&
		!fields.Fname &&
		!fields.Sname &&
		!fields.Phone &&
		!fields.Country &&
		!fields.City &&
		!fields.Birth &&
		!fields.Interests &&
		!fields.Likes &&
		!fields.Premium

	if fields.StatusEq != nil && fields.StatusNeq != nil && *fields.StatusEq == *fields.StatusNeq {
		filter.ExpectEmpty = true
	}

	if fields.CountryEq != nil {
		if fields.CityEq != nil {
			if !filter.dicts.ExistsCityInCountry(*fields.CityEq, *fields.CountryEq) {
				filter.ExpectEmpty = true
			}
		}

		if fields.CityAny != nil {
			anyExists := false
			for _, city := range *fields.CityAny {
				if filter.dicts.ExistsCityInCountry(city, *fields.CountryEq) {
					anyExists = true
				}
			}
			if !anyExists {
				filter.ExpectEmpty = true
			}
		}
	}

	return nil
}

func (filter *Filter) ParseParam(param string, value string) error {
	fields := &filter.Fields

	switch param {
	case "sex_eq":
		sex, err := filter.parser.ParseSex(value)
		if err != nil {
			return err
		}
		fields.SexEq = &sex
		fields.Sex = true
		// filter.NoFilter = false
	case "email_domain":
		fields.EmailDomain = &value
		fields.Email = true
		// filter.NoFilter = false
	case "email_lt":
		fields.EmailLt = &value
		fields.Email = true
		// filter.NoFilter = false
	case "email_gt":
		fields.EmailGt = &value
		fields.Email = true
		// filter.NoFilter = false
	case "status_eq":
		status, err := filter.parser.ParseStatus(value)
		if err != nil {
			return err
		}
		fields.StatusEq = &status
		fields.Status = true
		// filter.NoFilter = false
	case "status_neq":
		status, err := filter.parser.ParseStatus(value)
		if err != nil {
			return err
		}
		fields.StatusNeq = &status
		fields.Status = true
		// filter.NoFilter = false
	case "fname_eq":
		fname, err := filter.dicts.GetFname(value)
		if err != nil {
			filter.ExpectEmpty = true
			return nil
		}
		fields.FnameEq = &fname
		fields.Fname = true
		// filter.NoFilter = false
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
			filter.ExpectEmpty = true
			return nil
		}
		filter.Fields.FnameAny = &fnameAny
		fields.Fname = true
		// filter.NoFilter = false
	case "fname_null":
		fnameNull := value == "1"
		fields.FnameNull = &fnameNull
		fields.Fname = true
		// filter.NoFilter = false
	case "sname_eq":
		sname, err := filter.dicts.GetSname(value)
		if err != nil {
			filter.ExpectEmpty = true
			return nil
		}
		fields.SnameEq = &sname
		fields.Sname = true
		// filter.NoFilter = false
	case "sname_starts":
		fields.SnameStarts = &value
		fields.Sname = true
		// filter.NoFilter = false
	case "sname_null":
		snameNull := value == "1"
		fields.SnameNull = &snameNull
		fields.Sname = true
		// filter.NoFilter = false
	case "phone_code":
		ui64, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return err
		}
		phoneCode := uint16(ui64)
		fields.PhoneCode = &phoneCode
		fields.Phone = true
		// filter.NoFilter = false
	case "phone_null":
		phoneNull := value == "1"
		fields.PhoneNull = &phoneNull
		fields.Phone = true
		// filter.NoFilter = false
	case "country_eq":
		country, err := filter.dicts.GetCountry(value)
		if err != nil {
			filter.ExpectEmpty = true
			return nil
		}
		fields.CountryEq = &country
		fields.Country = true
		// filter.NoFilter = false
	case "country_null":
		countryNull := value == "1"
		fields.CountryNull = &countryNull
		fields.Country = true
		// filter.NoFilter = false
	case "city_eq":
		city, err := filter.dicts.GetCity(value)
		if err != nil {
			filter.ExpectEmpty = true
			return nil
		}
		fields.CityEq = &city
		fields.City = true
		// filter.NoFilter = false
	case "city_any":
		cityAny := make([]City, 0)
		cityStrs := strings.Split(value, ",")
		if len(cityStrs) == 0 {
			return nil
		}
		for _, cityStr := range cityStrs {
			city, err := filter.dicts.GetCity(cityStr)
			if err != nil {
				continue
			}
			cityAny = append(cityAny, city)
		}
		if len(cityAny) == 0 {
			filter.ExpectEmpty = true
			return nil
		}
		filter.Fields.CityAny = &cityAny
		fields.City = true
		// filter.NoFilter = false
	case "city_null":
		cityNull := value == "1"
		fields.CityNull = &cityNull
		fields.City = true
	case "birth_lt":
		ts, err := parseTimestamp(value)
		if err != nil {
			return err
		}
		fields.BirthLt = &ts
		fields.Birth = true
		// filter.NoFilter = false
	case "birth_gt":
		ts, err := parseTimestamp(value)
		if err != nil {
			return err
		}
		fields.BirthGt = &ts
		fields.Birth = true
		// filter.NoFilter = false
	case "birth_year":
		ui64, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return err
		}
		birthYear := Year(ui64)
		fields.BirthYear = &birthYear
		fields.Birth = true

		birthYearGte, birthYearLte := YearToTimestamp(birthYear)
		fields.BirthYearGte = &birthYearGte
		fields.BirthYearLte = &birthYearLte

		// filter.NoFilter = false
	case "interests_contains":
		interestsContains := make([]Interest, 0)
		interestsContainsStr := strings.Split(value, ",")
		for _, interestStr := range interestsContainsStr {
			interest, err := filter.dicts.GetInterest(interestStr)
			if err != nil {
				continue
			}
			interestsContains = append(interestsContains, interest)
		}
		if len(interestsContains) != len(interestsContainsStr) {
			filter.ExpectEmpty = true
			return nil
		}
		fields.InterestsContains = &interestsContains
		fields.Interests = true
		// filter.NoFilter = false
		// interestsContains := strings.Split(value, ",")
		// fields.InterestsContains = &interestsContains
	case "interests_any":
		interestsAny := make([]Interest, 0)
		interestsAnyStr := strings.Split(value, ",")
		for _, interestStr := range interestsAnyStr {
			interest, err := filter.dicts.GetInterest(interestStr)
			if err != nil {
				continue
			}
			interestsAny = append(interestsAny, interest)
		}
		if len(interestsAny) == 0 {
			filter.ExpectEmpty = true
			return nil
		}
		fields.InterestsAny = &interestsAny
		fields.Interests = true
		// filter.NoFilter = false
		// interesetsAny := strings.Split(value, ",")
		// fields.InterestsAny = &interesetsAny
	case "likes_contains":
		likeContains := make([]uint32, 0)
		likes := strings.Split(value, ",")
		for _, like := range likes {
			ui64, err := strconv.ParseUint(like, 10, 32)
			if err != nil {
				return err
			}
			likeContains = append(likeContains, uint32(ui64))
		}
		fields.LikesContains = &likeContains
		fields.Likes = true
		// filter.NoFilter = false
	case "premium_now":
		premiumNow := true
		fields.PremiumNow = &premiumNow
		fields.Premium = true
		// filter.NoFilter = false
	case "premium_null":
		premiumNull := value == "1"
		fields.PremiumNull = &premiumNull
		fields.Premium = true
		// filter.NoFilter = false
	case "limit":
		ui64, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return errors.New("Invalid limit value")
		}
		limit := uint8(ui64)
		fields.Limit = &limit
	case "query_id":
		filter.QueryID = &value
	default:
		return errors.New("Unknown filter param")
	}

	return nil
}

func parseTimestamp(timestamp string) (int64, error) {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return 0, err
	}
	return ts, nil
}

func YearToTimestamp(year Year) (gte, lte int64) {
	gte = time.Date(int(year), 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	lte = time.Date(int(year)+1, 1, 1, 0, 0, 0, 0, time.UTC).Unix() - 1

	return
}

func timestampToYear(timestamp int64) Year {
	return Year(time.Unix(timestamp, 0).Year())
}

func (filter *Filter) Sex() bool       { return filter.Fields.Sex }
func (filter *Filter) Status() bool    { return filter.Fields.Status }
func (filter *Filter) Fname() bool     { return filter.Fields.Fname }
func (filter *Filter) Sname() bool     { return filter.Fields.Sname }
func (filter *Filter) Phone() bool     { return filter.Fields.Phone }
func (filter *Filter) Country() bool   { return filter.Fields.Country }
func (filter *Filter) City() bool      { return filter.Fields.City }
func (filter *Filter) Birth() bool     { return filter.Fields.Birth }
func (filter *Filter) Premium() bool   { return filter.Fields.Premium }
func (filter *Filter) Interests() bool { return filter.Fields.Interests }
