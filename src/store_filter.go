package main

import (
	"strings"
)

func (store *Store) Filter(filter *Filter, accounts *AccountsBuffer) {
	if filter.ExpectEmpty() {
		return
	}

	it := store.findIds(filter)

	for it.Cur() != 0 {
		account := store.get(it.Cur())

		if !filter.NoFilter() && !store.filterAccount(account, filter) {
			it.Next()
			continue
		}

		*accounts = append(*accounts, account)

		if len(*accounts) >= filter.Limit() {
			break
		}
		it.Next()
	}

	// for _, id := range store.findIds(filter) {
	// 	account := store.get(id)

	// 	if !filter.NoFilter && !store.filterAccount(account, filter) {
	// 		continue
	// 	}

	// 	accounts = append(accounts, account)

	// 	if len(accounts) >= filter.Limit {
	// 		break
	// 	}
	// }
}

func (store *Store) findIds(filter *Filter) IndexIterator {
	if len(filter.LikesContains) > 0 {
		if len(filter.LikesContains) == 1 {
			likee := filter.LikesContains[0]
			filter.LikesContains = filter.LikesContains[:0]
			return store.index.Likee.Iter(ID(likee))
			// return store.index.Likee.Find(ID(likee))
		}
		likersAll := make([]IndexIterator, len(filter.LikesContains))
		// likersAll := BorrowIndexIterators()
		// defer likersAll.Release()
		for i, likeeContains := range filter.LikesContains {
			// likersAll[i] = store.index.Likee.Find(ID(likeeContains))
			likersAll[i] = store.index.Likee.Iter(ID(likeeContains))
		}
		filter.LikesContains = nil
		return NewIntersectIndexIterator(likersAll...)
		// return IntersectIndexes(likersAll...)
	}
	if filter.CityEq != 0 {
		city := filter.CityEq
		filter.CityEq = 0
		// return store.index.City.Find(city)
		return store.index.City.Iter(city)
	}
	if filter.BirthYear != 0 {
		birthYear := filter.BirthYear
		filter.BirthYear = 0
		// return store.index.BirthYear.Find(birthYear)
		if filter.CityNullSet && filter.CityNull {
			filter.CityNullSet = false
			// return store.index.City.Find(0)
			return NewIntersectIndexIterator(
				store.index.BirthYear.Iter(birthYear),
				store.index.City.Iter(0),
			)
		}
		return store.index.BirthYear.Iter(birthYear)
	}
	if filter.PhoneCode != 0 {
		phoneCode := filter.PhoneCode
		filter.PhoneCode = 0
		// return store.index.PhoneCode.Find(phoneCode)
		return store.index.PhoneCode.Iter(phoneCode)
	}
	if len(filter.InterestsContains) > 0 {
		if len(filter.InterestsContains) == 1 {
			interest := filter.InterestsContains[0]
			filter.InterestsContains = filter.InterestsContains[:0]
			// return store.index.Interest.Find(interest)
			return store.index.Interest.Iter(interest)
		}
		interestsAll := make([]IndexIterator, len(filter.InterestsContains))
		// interestsAll := BorrowIndexIterators()
		// defer interestsAll.Release()
		for i, interestContains := range filter.InterestsContains {
			// interestsAll[i] = store.index.Interest.Find(interestContains)
			interestsAll[i] = store.index.Interest.Iter(interestContains)
		}
		filter.InterestsContains = filter.InterestsContains[:0]
		if filter.CountryNullSet && filter.CountryNull {
			filter.CountryNullSet = false
			interestsAll = append(interestsAll, store.index.Country.Iter(0))
		}
		return NewIntersectIndexIterator((interestsAll)...)
		// return IntersectIndexes(interestsAll...)
	}
	if len(filter.InterestsAny) > 0 {
		interestsAny := make([]IndexIterator, len(filter.InterestsAny))
		// interestsAny := BorrowIndexIterators()
		// defer interestsAny.Release()

		for i, interestAny := range filter.InterestsAny {
			interestsAny[i] = store.index.Interest.Iter(interestAny)
		}
		filter.InterestsAny = filter.InterestsAny[:0]
		interestsIt := NewUnionIndexIterator((interestsAny)...)
		if filter.CountryEq != 0 {
			country := filter.CountryEq
			filter.CountryEq = 0
			// return store.index.Country.Find(country)
			return NewIntersectIndexIterator(interestsIt, store.index.Country.Iter(country))
		}
		return interestsIt
	}
	if len(filter.CityAny) > 0 {
		citiesAny := make([]IndexIterator, len(filter.CityAny))
		// citiesAny := BorrowIndexIterators()
		// defer citiesAny.Release()

		for i, cityAny := range filter.CityAny {
			citiesAny[i] = store.index.City.Iter(cityAny)
		}
		filter.CityAny = filter.CityAny[:0]
		return NewUnionIndexIterator((citiesAny)...)
	}
	if len(filter.FnameAny) > 0 {
		// fnamesAny := make([]IDS, len(filter.FnameAny))
		fnamesAny := make([]IndexIterator, len(filter.FnameAny))
		// fnamesAny := BorrowIndexIterators()
		// defer fnamesAny.Release()

		for i, fnameAny := range filter.FnameAny {
			// fnamesAny[i] = store.index.Fname.Find(fnameAny)
			// fnamesAny[i] = store.index.Fname.Iter(fnameAny)
			fnamesAny[i] = store.index.Fname.Iter(fnameAny)
		}
		filter.FnameAny = filter.FnameAny[:0]
		if filter.CountryEq != 0 {
			country := filter.CountryEq
			filter.CountryEq = 0
			// return store.index.Country.Find(country)
			return NewIntersectIndexIterator(
				NewUnionIndexIterator((fnamesAny)...),
				store.index.Country.Iter(country),
			)
		}
		return NewUnionIndexIterator((fnamesAny)...)
		// return UnionIndexes(fnamesAny...)
	}
	if filter.CityNullSet {
		if filter.CityNull {
			filter.CityNullSet = false
			// return store.index.City.Find(0)
			return store.index.City.Iter(0)
		}
	}
	if filter.CountryEq != 0 {
		country := filter.CountryEq
		filter.CountryEq = 0
		// return store.index.Country.Find(country)
		return store.index.Country.Iter(country)
	}
	if filter.CountryNullSet {
		if filter.CountryNull {
			filter.CountryNullSet = false
			// return store.index.Country.Find(0)
			return store.index.Country.Iter(0)
		}
	}
	// return store.index.ID.FindAll()
	return store.index.ID.Iter()
}

func (store *Store) filterAccount(account *Account, filter *Filter) bool {
	if filter.SexEq != 0 {
		if account.Sex != filter.SexEq {
			return false
		}
	}
	if filter.EmailDomain != "" {
		if account.Email[account.EmailDomain:] != filter.EmailDomain {
			return false
		}
	}
	if filter.StatusEq != 0 {
		if account.Status != filter.StatusEq {
			return false
		}
	}
	if filter.StatusNeq != 0 {
		if account.Status == filter.StatusNeq {
			return false
		}
	}
	if filter.FnameEq != 0 {
		if account.Fname == 0 {
			return false
		}
		if account.Fname != filter.FnameEq {
			return false
		}
	}
	if filter.FnameNullSet {
		if filter.FnameNull {
			if account.Fname != 0 {
				return false
			}
		} else {
			if account.Fname == 0 {
				return false
			}
		}
	}
	if filter.SnameEq != 0 {
		if account.Sname == 0 {
			return false
		}
		if account.Sname != filter.SnameEq {
			return false
		}
	}
	if filter.SnameNullSet {
		if filter.SnameNull {
			if account.Sname != 0 {
				return false
			}
		} else {
			if account.Sname == 0 {
				return false
			}
		}
	}
	if filter.PhoneCode != 0 {
		if account.PhoneCode != filter.PhoneCode {
			return false
		}
	}
	if filter.PhoneNullSet {
		if filter.PhoneNull {
			if account.Phone != nil {
				return false
			}
		} else {
			if account.Phone == nil {
				return false
			}
		}
	}
	if filter.CountryEq != 0 {
		if account.Country == 0 {
			return false
		}

		if account.Country != filter.CountryEq {
			return false
		}
	}
	if filter.CountryNullSet {
		if filter.CountryNull {
			if account.Country != 0 {
				return false
			}
		} else {
			if account.Country == 0 {
				return false
			}
		}
	}
	if filter.PremiumNullSet {
		if filter.PremiumNull {
			if account.Premium != nil {
				return false
			}
		} else {
			if account.Premium == nil {
				return false
			}
		}
	}
	if filter.CityEq != 0 {
		if account.City == 0 {
			return false
		}

		if account.City != filter.CityEq {
			return false
		}
	}
	if len(filter.CityAny) > 0 {
		if account.City == 0 {
			return false
		}
		any := false
		for _, city := range filter.CityAny {
			if city == account.City {
				any = true
			}
		}
		if !any {
			return false
		}
	}
	if filter.CityNullSet {
		if filter.CityNull {
			if account.City != 0 {
				return false
			}
		} else {
			if account.City == 0 {
				return false
			}
		}
	}
	if filter.BirthLt != 0 {
		if int64(account.Birth) >= filter.BirthLt {
			return false
		}
	}
	if filter.BirthGt != 0 {
		if int64(account.Birth) <= filter.BirthGt {
			return false
		}
	}
	if filter.BirthYear != 0 {
		if int64(account.Birth) < filter.BirthYearGte || int64(account.Birth) > filter.BirthYearLte {
			return false
		}
	}
	if filter.PremiumNow {
		if !store.PremiumNow(account) {
			return false
		}
	}
	if len(filter.InterestsContains) > 0 {
		if len(account.Interests) == 0 {
			return false
		}
		contains := true
		for _, interestContains := range filter.InterestsContains {
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
	if len(filter.InterestsAny) > 0 {
		if len(account.Interests) == 0 {
			return false
		}

		any := false
		for _, interestAny := range filter.InterestsAny {
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
	if len(filter.LikesContains) != 0 {
		likes := store.index.Liker.Find(account.ID)
		if len(likes) == 0 {
			return false
		}
		contains := true
		for _, likeID := range filter.LikesContains {
			containsLike := false
			for _, like := range likes {
				if like.ID == ID(likeID) {
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
	if filter.SnameStarts != "" {
		if account.Sname == 0 {
			return false
		}
		snameStr, err := store.dicts.GetSnameString(account.Sname)
		if err != nil {
			return false
		}
		if !strings.HasPrefix(snameStr, filter.SnameStarts) {
			return false
		}
	}
	if filter.EmailGt != "" {
		if strings.Compare(account.Email, filter.EmailGt) != +1 {
			return false
		}
	}
	if filter.EmailLt != "" {
		if strings.Compare(account.Email, filter.EmailLt) != -1 {
			return false
		}
	}
	if len(filter.FnameAny) > 0 {
		if account.Fname == 0 {
			return false
		}
		any := false
		for _, fname := range filter.FnameAny {
			if fname == account.Fname {
				any = true
			}
		}
		if !any {
			return false
		}
	}
	return true
}
