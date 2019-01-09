package main

import (
	"sort"
	"strings"
)

func (store *Store) FilterAll(filter *Filter) AccountSlice {
	accounts := make(AccountSlice, 0, 50)

	if filter.ExpectEmpty {
		return accounts
	}

	// scan all
	count := uint8(1)
	for _, ID := range store.findIds(filter) {
		account := store.accounts[ID]

		if !filter.NoFilter && !store.filterAccount(account, filter) {
			continue
		}

		accounts = append(accounts, account)

		if count >= *filter.Fields.Limit {
			break
		}
		count++
	}

	return accounts
}

func (store *Store) findIds(filter *Filter) IDS {
	fields := &filter.Fields

	if fields.LikesContains != nil {
		if len(*fields.LikesContains) == 1 {
			interest := (*fields.LikesContains)[0]
			fields.LikesContains = nil
			return store.indexLikee.Get(interest)
		}

		likersAll := make([]IDS, len(*fields.LikesContains))
		for i, likeeContains := range *fields.LikesContains {
			likersAll[i] = store.indexLikee.Get(likeeContains)
		}

		likers := make(IDS, 0)

		firstI := 0
		for _, liker := range likersAll[firstI] {
			exists := true
			for i := 0; i < len(likersAll); i++ {
				if i == firstI {
					continue
				}
				likerIndex := sort.Search(len(likersAll[i]), func(j int) bool {
					return likersAll[i][j] <= liker
				})
				if likerIndex == len(likersAll[i]) || likersAll[i][likerIndex] != liker {
					exists = false
					break
				}
			}
			if exists {
				likers = append(likers, liker)
			}
		}

		fields.LikesContains = nil
		return likers
	}

	if fields.CityEq != nil {
		city := *fields.CityEq
		fields.CityEq = nil
		return store.indexCity.Get(city)
	}

	if fields.BirthYear != nil {
		birthYear := *fields.BirthYear
		fields.BirthYear = nil
		return store.indexBirthYear.Get(birthYear)
	}

	if fields.PhoneCode != nil {
		phoneCode := *fields.PhoneCode
		fields.PhoneCode = nil
		return store.indexPhoneCode.Get(phoneCode)
	}

	if fields.InterestsContains != nil {
		if len(*fields.InterestsContains) == 1 {
			interest := (*fields.InterestsContains)[0]
			fields.InterestsContains = nil
			return store.indexInterest.Get(interest)
		}

		interestsAll := make([]IDS, len(*fields.InterestsContains))
		for i, interestContains := range *fields.InterestsContains {
			interestsAll[i] = store.indexInterest.Get(interestContains)
		}

		ids := make(IDS, 0)

		firstI := 0
		for _, ID := range interestsAll[firstI] {
			exists := true
			for i := 0; i < len(interestsAll); i++ {
				if i == firstI {
					continue
				}
				interestIndex := sort.Search(len(interestsAll[i]), func(j int) bool {
					return interestsAll[i][j] <= ID
				})
				if interestIndex == len(interestsAll[i]) || interestsAll[i][interestIndex] != ID {
					exists = false
					break
				}
			}
			if exists {
				ids = append(ids, ID)
			}
		}

		fields.InterestsContains = nil
		return ids
	}

	// if fields.InterestsAny != nil {

	// }

	if fields.CityAny != nil {
		citiesCur := make([]uint32, len(*fields.CityAny))

		citiesAny := make([]IDS, len(*fields.CityAny))
		for i, cityAny := range *fields.CityAny {
			citiesAny[i] = store.indexCity.Get(cityAny)
		}

		ids := make(IDS, 0)
		for {
			maxID := uint32(0)
			maxCity := -1
			for i, cityCur := range citiesCur {
				if cityCur < uint32(len(citiesAny[i])) && citiesAny[i][cityCur] > maxID {
					maxID = citiesAny[i][cityCur]
					maxCity = i
				}
			}

			if maxID > 0 {
				ids = append(ids, maxID)
				citiesCur[maxCity]++
			} else {
				break
			}
		}

		return ids
	}

	if fields.FnameAny != nil {
		fnamesCur := make([]uint32, len(*fields.FnameAny))

		fnamesAny := make([]IDS, len(*fields.FnameAny))
		for i, fnameAny := range *fields.FnameAny {
			fnamesAny[i] = store.indexFname.Get(fnameAny)
		}

		ids := make(IDS, 0)
		for {
			maxID := uint32(0)
			maxFname := -1
			for i, fnameCur := range fnamesCur {
				if fnameCur < uint32(len(fnamesAny[i])) && fnamesAny[i][fnameCur] > maxID {
					maxID = fnamesAny[i][fnameCur]
					maxFname = i
				}
			}

			if maxID > 0 {
				ids = append(ids, maxID)
				fnamesCur[maxFname]++
			} else {
				break
			}
		}

		return ids
	}

	if fields.CityNull != nil {
		if *fields.CityNull {
			fields.CityNull = nil
			return store.indexCity.Get(0)
		}
	}

	if fields.CountryEq != nil {
		country := *fields.CountryEq
		fields.CountryEq = nil
		return store.indexCountry.Get(country)
	}

	if fields.CountryNull != nil {
		if *fields.CountryNull {
			fields.CountryNull = nil
			return store.indexCountry.Get(0)
		}
	}

	return store.indexID.FindAll()
}

func (store *Store) filterAccount(account *Account, filter *Filter) bool {
	fields := &filter.Fields

	if fields.SexEq != nil {
		if account.Sex != *fields.SexEq {
			return false
		}
	}

	if fields.EmailDomain != nil {
		if account.Email[account.EmailDomain:] != *fields.EmailDomain {
			return false
		}
	}

	if fields.EmailGt != nil {
		if strings.Compare(account.Email, *fields.EmailGt) != +1 {
			return false
		}
	}

	if fields.EmailLt != nil {
		if strings.Compare(account.Email, *fields.EmailLt) != -1 {
			return false
		}
	}

	if fields.StatusEq != nil {
		if account.Status != *fields.StatusEq {
			return false
		}
	}

	if fields.StatusNeq != nil {
		if account.Status == *fields.StatusNeq {
			return false
		}
	}

	if fields.FnameEq != nil {
		if account.Fname == 0 {
			return false
		}

		if account.Fname != *fields.FnameEq {
			return false
		}
	}

	if fields.FnameAny != nil {
		if account.Fname == 0 {
			return false
		}

		any := false
		for _, fname := range *fields.FnameAny {
			if fname == account.Fname {
				any = true
			}
		}
		if !any {
			return false
		}
	}

	if fields.FnameNull != nil {
		if *fields.FnameNull {
			if account.Fname != 0 {
				return false
			}
		} else {
			if account.Fname == 0 {
				return false
			}
		}
	}

	if fields.SnameEq != nil {
		if account.Sname == 0 {
			return false
		}

		if account.Sname != *fields.SnameEq {
			return false
		}
	}

	if fields.SnameStarts != nil {
		if account.Sname == 0 {
			return false
		}

		snameStr, err := store.dicts.GetSnameString(account.Sname)
		if err != nil {
			return false
		}
		if !strings.HasPrefix(snameStr, *fields.SnameStarts) {
			return false
		}
	}

	if fields.SnameNull != nil {
		if *fields.SnameNull {
			if account.Sname != 0 {
				return false
			}
		} else {
			if account.Sname == 0 {
				return false
			}
		}
	}

	if fields.PhoneCode != nil {
		if account.PhoneCode == nil {
			return false
		}

		if *account.PhoneCode != *fields.PhoneCode {
			return false
		}
	}

	if fields.PhoneNull != nil {
		if *fields.PhoneNull {
			if account.Phone != nil {
				return false
			}
		} else {
			if account.Phone == nil {
				return false
			}
		}
	}

	if fields.CountryEq != nil {
		if account.Country == 0 {
			return false
		}

		if account.Country != *fields.CountryEq {
			return false
		}
	}

	if fields.CountryNull != nil {
		if *fields.CountryNull {
			if account.Country != 0 {
				return false
			}
		} else {
			if account.Country == 0 {
				return false
			}
		}
	}

	if fields.CityEq != nil {
		if account.City == 0 {
			return false
		}

		if account.City != *fields.CityEq {
			return false
		}
	}

	if fields.CityAny != nil {
		if account.City == 0 {
			return false
		}

		any := false
		for _, city := range *fields.CityAny {
			if city == account.City {
				any = true
			}
		}
		if !any {
			return false
		}
	}

	if fields.CityNull != nil {
		if *fields.CityNull {
			if account.City != 0 {
				return false
			}
		} else {
			if account.City == 0 {
				return false
			}
		}
	}

	if fields.BirthLt != nil {
		if account.Birth >= *fields.BirthLt {
			return false
		}
	}

	if fields.BirthGt != nil {
		if account.Birth <= *fields.BirthGt {
			return false
		}
	}

	if fields.BirthYear != nil {
		if account.Birth < *fields.BirthYearGte || account.Birth > *fields.BirthYearLte {
			return false
		}
	}

	if fields.InterestsContains != nil {
		if len(account.Interests) == 0 {
			return false
		}
		contains := true
		for _, interestContains := range *fields.InterestsContains {
			containsInterest := false
			for _, interest := range account.Interests {
				if interest == interestContains {
					containsInterest = true
					break
				}
			}
			if !containsInterest {
				contains = false
				break
			}
		}
		if !contains {
			return false
		}
	}

	if fields.InterestsAny != nil {
		if len(account.Interests) == 0 {
			return false
		}

		any := false
		for _, interestAny := range *fields.InterestsAny {
			for _, interest := range account.Interests {
				if interest == interestAny {
					any = true
				}
			}
		}
		if !any {
			return false
		}
	}

	if fields.LikesContains != nil {
		if len(account.Likes) == 0 {
			return false
		}

		contains := true
		for _, likeID := range *fields.LikesContains {
			containsLike := false
			for _, like := range account.Likes {
				if like.ID == likeID {
					containsLike = true
					break
				}
			}
			if !containsLike {
				contains = false
				break
			}
		}
		if !contains {
			return false
		}
	}

	if fields.PremiumNow != nil {
		if account.Premium == nil {
			return false
		}

		if account.Premium.Start > store.now || store.now > account.Premium.Finish {
			return false
		}
	}

	if fields.PremiumNull != nil {
		if *fields.PremiumNull {
			if account.Premium != nil {
				return false
			}
		} else {
			if account.Premium == nil {
				return false
			}
		}
	}

	return true
}
