package main

import (
	"errors"
)

type (
	Fname    uint8
	Sname    uint16
	Country  uint8
	City     uint16
	Interest uint8
)

type Dicts struct {
	fnames        map[string]Fname
	fnameStrs     map[Fname]string
	snames        map[string]Sname
	snameStrs     map[Sname]string
	countries     map[string]Country
	countryStrs   map[Country]string
	cities        map[string]City
	cityStrs      map[City]string
	countryCities map[Country][]City
	cityCountry   map[City]Country
	interests     map[string]Interest
	interestStrs  map[Interest]string
}

func NewDicts() *Dicts {
	return &Dicts{
		snames:        make(map[string]Sname),
		snameStrs:     make(map[Sname]string),
		fnames:        make(map[string]Fname),
		fnameStrs:     make(map[Fname]string),
		countries:     make(map[string]Country),
		countryStrs:   make(map[Country]string),
		cities:        make(map[string]City),
		cityStrs:      make(map[City]string),
		countryCities: make(map[Country][]City),
		cityCountry:   make(map[City]Country),
		interests:     make(map[string]Interest),
		interestStrs:  make(map[Interest]string),
	}
}

func (dicts *Dicts) AddFname(fnameStr string) Fname {
	fname, exists := dicts.fnames[fnameStr]
	if exists {
		return fname
	}

	fname = Fname(len(dicts.fnames) + 1)
	dicts.fnames[fnameStr] = fname
	dicts.fnameStrs[fname] = fnameStr

	return fname
}

func (dicts *Dicts) GetFname(fnameStr string) (Fname, error) {
	fname, exists := dicts.fnames[fnameStr]
	if !exists {
		return 0, errors.New("Cannot find fname")
	}
	return fname, nil
}

func (dicts *Dicts) GetFnameString(fname Fname) (string, error) {
	fnameStr, exists := dicts.fnameStrs[fname]
	if !exists {
		return "", errors.New("Cannot find fname string")
	}
	return fnameStr, nil
}

func (dicts *Dicts) GetFnames() map[string]Fname {
	return dicts.fnames
}

func (dicts *Dicts) AddSname(snameStr string) Sname {
	sname, exists := dicts.snames[snameStr]
	if exists {
		return sname
	}

	sname = Sname(len(dicts.snames) + 1)
	dicts.snames[snameStr] = sname
	dicts.snameStrs[sname] = snameStr

	return sname
}

func (dicts *Dicts) GetSname(snameStr string) (Sname, error) {
	sname, exists := dicts.snames[snameStr]
	if !exists {
		return 0, errors.New("Cannot find sname")
	}
	return sname, nil
}

func (dicts *Dicts) GetSnameString(sname Sname) (string, error) {
	snameStr, exists := dicts.snameStrs[sname]
	if !exists {
		return "", errors.New("Cannot find sname string")
	}
	return snameStr, nil
}

func (dicts *Dicts) GetSnames() map[string]Sname {
	return dicts.snames
}

func (dicts *Dicts) AddCountry(countryStr string) Country {
	country, exists := dicts.countries[countryStr]
	if exists {
		return country
	}

	country = Country(len(dicts.countries) + 1)
	dicts.countries[countryStr] = country
	dicts.countryStrs[country] = countryStr
	dicts.countryCities[country] = make([]City, 0, 15)

	return country
}

func (dicts *Dicts) GetCountry(countryStr string) (Country, error) {
	country, exists := dicts.countries[countryStr]
	if !exists {
		return 0, errors.New("Cannot find country")
	}
	return country, nil
}

func (dicts *Dicts) GetCountryString(country Country) (string, error) {
	countryStr, exists := dicts.countryStrs[country]
	if !exists {
		return "", errors.New("Cannot find country string")
	}
	return countryStr, nil
}

func (dicts *Dicts) GetCountries() map[string]Country {
	return dicts.countries
}

func (dicts *Dicts) AddCity(country Country, cityStr string) City {
	city, exists := dicts.cities[cityStr]
	if exists {
		return city
	}

	city = City(len(dicts.cities) + 1)
	dicts.cities[cityStr] = city
	dicts.cityStrs[city] = cityStr

	if country != 0 {
		dicts.countryCities[country] = append(dicts.countryCities[country], city)
		dicts.cityCountry[city] = country
	}

	return city
}

func (dicts *Dicts) GetCity(cityStr string) (City, error) {
	city, exists := dicts.cities[cityStr]
	if !exists {
		return 0, errors.New("Cannot find city")
	}
	return city, nil
}

func (dicts *Dicts) GetCityString(city City) (string, error) {
	cityStr, exists := dicts.cityStrs[city]
	if !exists {
		return "", errors.New("Cannot find city string")
	}
	return cityStr, nil
}

func (dicts *Dicts) GetCities() map[string]City {
	return dicts.cities
}

func (dicts *Dicts) ExistsCityInCountry(city City, country Country) bool {
	cityCountry, exists := dicts.cityCountry[city]
	if !exists {
		return false
	}
	return country == cityCountry
}

func (dicts *Dicts) UpdateCountryCities(store *Store) {
	for cityStr := range dicts.cities {
		city := dicts.cities[cityStr]
		if _, known := dicts.cityCountry[city]; known {
			continue
		}
		for _, account := range *store.GetAll() {
			if account.City == city && account.Country != 0 {
				dicts.cityCountry[city] = account.Country
				dicts.countryCities[account.Country] = append(dicts.countryCities[account.Country], account.City)
			}
		}
	}
}

func (dicts *Dicts) AddInterest(interestStr string) Interest {
	interest, exists := dicts.interests[interestStr]
	if exists {
		return interest
	}

	interest = Interest(len(dicts.interests) + 1)
	dicts.interests[interestStr] = interest
	dicts.interestStrs[interest] = interestStr

	return interest
}

func (dicts *Dicts) GetInterest(interestStr string) (Interest, error) {
	interest, exists := dicts.interests[interestStr]
	if !exists {
		return 0, errors.New("Cannot find interest")
	}
	return interest, nil
}

func (dicts *Dicts) GetInterestString(interest Interest) (string, error) {
	interestStr, exists := dicts.interestStrs[interest]
	if !exists {
		return "", errors.New("Cannot find interest string")
	}
	return interestStr, nil
}

func (dicts *Dicts) GetInterests() map[string]Interest {
	return dicts.interests
}
