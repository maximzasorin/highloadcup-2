package main

// type Recommend struct {
// 	store       *Store
// 	dicts       *Dicts
// 	Account     *Account
// 	QueryID     *string
// 	ExpectEmpty bool
// 	Country     *Country
// 	City        *City
// 	Limit       *uint8
// }

// func NewRecommend(store *Store, dicts *Dicts) *Recommend {
// 	return &Recommend{
// 		store: store,
// 		dicts: dicts,
// 	}
// }

// func (recommend *Recommend) Parse(accountID, query string) error {
// 	ui64, err := strconv.ParseUint(accountID, 10, 32)
// 	if err != nil {
// 		return errors.Wrap(err, "Invalid account ID")
// 	}
// 	recommend.Account = recommend.store.Get(uint32(ui64))

// 	values, err := url.ParseQuery(query)
// 	if err != nil {
// 		return err
// 	}

// 	for param, paramValues := range values {
// 		if len(paramValues) != 1 || paramValues[0] == "" {
// 			return errors.New("Invalid recommend param value")
// 		}

// 		err := recommend.ParseParam(param, paramValues[0])
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	if recommend.Limit == nil {
// 		return errors.New("Limit should be specified")
// 	}
// 	if *recommend.Limit > 20 {
// 		return errors.New("Limit should be less than 20")
// 	}
// 	return nil
// }

// func (recommend *Recommend) ParseParam(param string, value string) error {
// 	switch param {
// 	case "country":
// 		country, err := recommend.dicts.GetCountry(value)
// 		if err != nil {
// 			recommend.ExpectEmpty = true
// 			return nil
// 		}
// 		recommend.Country = &country
// 	case "city":
// 		city, err := recommend.dicts.GetCity(value)
// 		if err != nil {
// 			recommend.ExpectEmpty = true
// 			return nil
// 		}
// 		recommend.City = &city
// 	case "limit":
// 		ui64, err := strconv.ParseUint(value, 10, 8)
// 		if err != nil {
// 			return errors.New("Invalid limit value")
// 		}
// 		limit := uint8(ui64)
// 		recommend.Limit = &limit
// 	case "query_id":
// 		recommend.QueryID = &value
// 	default:
// 		return errors.New("Unknown recommend param")
// 	}

// 	return nil
// }
