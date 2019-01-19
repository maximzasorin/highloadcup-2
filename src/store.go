package main

import (
	"errors"
	"fmt"
	"sync"
)

type Store struct {
	parser                    *Parser
	dicts                     *Dicts
	now                       uint32
	test                      bool
	accounts                  map[ID]*Account
	emails                    map[string]ID
	rwLock                    sync.RWMutex
	indexID                   *IndexID
	indexLikee                *IndexLikee
	indexInterest             *IndexInterest
	indexCity                 *IndexCity
	indexBirthYear            *IndexYear
	indexJoinedYear           *IndexYear
	indexCountry              *IndexCountry
	indexFname                *IndexFname
	indexPhoneCode            *IndexPhoneCode
	indexGroup                *IndexGroup
	indexSex                  *IndexSex
	indexStatus               *IndexStatus
	indexInterestPremium      *IndexInterestPremium
	indexInterestSingle       *IndexInterestSingle
	indexInterestComplicated  *IndexInterestComplicated
	indexInterestRelationship *IndexInterestRelationship
}

func NewStore(parser *Parser, dicts *Dicts) *Store {
	return &Store{
		parser:                    parser,
		dicts:                     dicts,
		accounts:                  make(map[ID]*Account),
		emails:                    make(map[string]ID),
		indexID:                   NewIndexID(10000),
		indexLikee:                NewIndexLikee(),
		indexInterest:             NewIndexInterest(),
		indexCity:                 NewIndexCity(),
		indexBirthYear:            NewIndexYear(),
		indexJoinedYear:           NewIndexYear(),
		indexCountry:              NewIndexCountry(),
		indexFname:                NewIndexFname(),
		indexPhoneCode:            NewIndexPhoneCode(),
		indexGroup:                NewIndexGroup(),
		indexSex:                  NewIndexSex(),
		indexStatus:               NewIndexStatus(),
		indexInterestPremium:      NewIndexInterestPremium(),
		indexInterestSingle:       NewIndexInterestSingle(),
		indexInterestComplicated:  NewIndexInterestComplicated(),
		indexInterestRelationship: NewIndexInterestRelationship(),
	}
}

func (store *Store) SetNow(now uint32) {
	store.now = now
}

func (store *Store) GetNow() uint32 {
	return store.now
}

func (store *Store) Count() uint32 {
	return uint32(len(store.accounts))
}

func (store *Store) Add(rawAccount *RawAccount, check bool, updateIndexes bool) (*Account, error) {
	if check {
		if rawAccount.Email == "" {
			return nil, errors.New("Email field should be specified")
		}
		if rawAccount.EmailDomain == 0 {
			return nil, errors.New("Invalid email")
		}
		if rawAccount.Sex == 0 {
			return nil, errors.New("Sex field should be specified")
		}
		if rawAccount.Status == 0 {
			return nil, errors.New("Status field should be specified")
		}
		if rawAccount.Birth == 0 {
			return nil, errors.New("Birth field should be specified")
		}
		if rawAccount.Joined == 0 {
			return nil, errors.New("Joined field should be specified")
		}
		// for _, like := range rawAccount.Likes {
		// 	if _, ok := store.accounts[like.ID]; !ok {
		// 		return nil, errors.New("Like with unknown account ID")
		// 	}
		// }
		if rawAccount.ID == 0 {
			return nil, errors.New("Need account ID")
		}
	}

	account := Account{
		ID:          ID(rawAccount.ID),
		Sex:         rawAccount.Sex,
		Status:      rawAccount.Status,
		Birth:       rawAccount.Birth,
		Joined:      rawAccount.Joined,
		Premium:     rawAccount.Premium,
		Email:       rawAccount.Email,
		EmailDomain: rawAccount.EmailDomain,
	}

	store.rwLock.Lock()
	if check {
		if _, ok := store.emails[account.Email]; ok {
			store.rwLock.Unlock()
			return nil, errors.New("Same email already taken")
		}
		if _, ok := store.accounts[account.ID]; ok {
			store.rwLock.Unlock()
			return nil, errors.New("Account with same ID already exists")
		}
	}
	store.accounts[account.ID] = &account
	store.emails[account.Email] = account.ID
	store.rwLock.Unlock()

	if rawAccount.Phone != nil {
		account.Phone = rawAccount.Phone
		account.PhoneCode = rawAccount.PhoneCode
	}

	if rawAccount.Fname != nil {
		account.Fname = store.dicts.AddFname(*rawAccount.Fname)
	}

	if rawAccount.Sname != nil {
		account.Sname = store.dicts.AddSname(*rawAccount.Sname)
	}

	if rawAccount.Country != nil {
		account.Country = store.dicts.AddCountry(*rawAccount.Country)
	}

	if rawAccount.City != nil {
		account.City = store.dicts.AddCity(account.Country, *rawAccount.City)
	}

	for _, like := range rawAccount.Likes {
		store.indexLikee.Add(ID(like.ID), ID(account.ID))
	}

	for _, interestStr := range rawAccount.Interests {
		interest := store.dicts.AddInterest(interestStr)
		account.Interests = append(account.Interests, interest)
	}

	if updateIndexes {
		store.AddToIndex(&account)
	}

	return &account, nil
}

func (store *Store) AddToIndex(account *Account) {
	store.indexID.Add(account.ID)
	store.indexSex.Add(account.Sex, account.ID)
	store.indexStatus.Add(account.Status, account.ID)
	store.indexBirthYear.Add(timestampToYear(account.Birth), account.ID)
	store.indexJoinedYear.Add(timestampToYear(int64(account.Joined)), account.ID)
	if account.Phone != nil {
		store.indexPhoneCode.Add(*account.PhoneCode, account.ID)
	} else {
		store.indexPhoneCode.Add(0, account.ID)
	}
	if account.Fname != 0 {
		store.indexFname.Add(account.Fname, account.ID)
	} else {
		store.indexFname.Add(0, account.ID)
	}
	if account.Country != 0 {
		store.indexCountry.Add(account.Country, account.ID)
	} else {
		store.indexCountry.Add(0, account.ID)
	}
	if account.City != 0 {
		store.indexCity.Add(account.City, account.ID)
	} else {
		store.indexCity.Add(0, account.ID)
	}
	for _, interest := range account.Interests {
		store.indexInterest.Add(interest, account.ID)
		if store.PremiumNow(account) {
			store.indexInterestPremium.Add(interest, account.Status, account.Sex, account.City, account.Country, account.ID)
		} else {
			if account.Status == StatusSingle {
				store.indexInterestSingle.Add(interest, account.City, account.Country, account.ID)
			} else if account.Status == StatusComplicated {
				store.indexInterestComplicated.Add(interest, account.City, account.Country, account.ID)
			} else if account.Status == StatusRelationship {
				store.indexInterestRelationship.Add(interest, account.City, account.Country, account.ID)
			}
		}
	}
}

func (store *Store) AppendToIndex(account *Account) {
	store.indexID.Append(account.ID)
	store.indexSex.Append(account.Sex, account.ID)
	store.indexStatus.Append(account.Status, account.ID)
	store.indexBirthYear.Append(timestampToYear(account.Birth), account.ID)
	store.indexJoinedYear.Append(timestampToYear(int64(account.Joined)), account.ID)
	store.indexFname.Append(account.Fname, account.ID)
	store.indexCountry.Append(account.Country, account.ID)
	store.indexCity.Append(account.City, account.ID)
	if account.Phone != nil {
		store.indexPhoneCode.Append(*account.PhoneCode, account.ID)
	} else {
		store.indexPhoneCode.Append(0, account.ID)
	}
	for _, interest := range account.Interests {
		store.indexInterest.Append(interest, account.ID)
		if store.PremiumNow(account) {
			store.indexInterestPremium.Append(interest, account.Status, account.Sex, account.City, account.Country, account.ID)
		} else {
			if account.Status == StatusSingle {
				store.indexInterestSingle.Append(interest, account.City, account.Country, account.ID)
			}
			if account.Status == StatusComplicated {
				store.indexInterestComplicated.Append(interest, account.City, account.Country, account.ID)
			}
			if account.Status == StatusRelationship {
				store.indexInterestRelationship.Append(interest, account.City, account.Country, account.ID)
			}
		}
	}
	// store.indexGroup.Add(account)
}

func (store *Store) UpdateIndex() {
	store.indexID.Update()
	store.indexSex.UpdateAll()
	store.indexStatus.UpdateAll()
	store.indexBirthYear.UpdateAll()
	store.indexJoinedYear.UpdateAll()
	store.indexFname.UpdateAll()
	store.indexCountry.UpdateAll()
	store.indexCity.UpdateAll()
	store.indexPhoneCode.UpdateAll()
	store.indexInterest.UpdateAll()
	store.indexInterestPremium.UpdateAll()
	store.indexInterestSingle.UpdateAll()
	store.indexInterestComplicated.UpdateAll()
	store.indexInterestRelationship.UpdateAll()

	fmt.Println("total birth years =", store.indexBirthYear.Len())
	fmt.Println("total joined years =", store.indexJoinedYear.Len())
	fmt.Println("total fnames =", store.indexFname.Len())
	fmt.Println("total countries =", store.indexCountry.Len())
	fmt.Println("total cities =", store.indexCity.Len())
	fmt.Println("total phone codes =", store.indexPhoneCode.Len())
	fmt.Println("total interests =", store.indexInterest.Len())
	fmt.Println("total entries =", store.indexGroup.Len())

	// for fields, entries := range store.indexGroup.entries {
	// 	fmt.Println(fields, len(entries))
	// }
}

func (store *Store) AddLikes(rawLikes []*Like, updateIndexes bool) error {
	// if len(rawLikes) == 0 {
	// 	return errors.New("No likes founded")
	// }

	for _, rawLike := range rawLikes {
		store.rwLock.RLock()
		_, ok := store.accounts[ID(rawLike.Likee)]
		store.rwLock.RUnlock()
		if !ok {
			return errors.New("Cannot find likee account")
		}
		store.rwLock.RLock()
		_, ok = store.accounts[ID(rawLike.Liker)]
		store.rwLock.RUnlock()
		if !ok {
			return errors.New("Cannot find liker account")
		}
	}

	for _, rawLike := range rawLikes {
		store.indexLikee.Add(ID(rawLike.Likee), ID(rawLike.Liker))
	}

	return nil
}

func (store *Store) Update(id ID, rawAccount *RawAccount, updateIndexes bool) (*Account, error) {
	if rawAccount.Email != "" && rawAccount.EmailDomain == 0 {
		return nil, errors.New("Invalid email")
	}
	store.rwLock.RLock()
	emailID, ok := store.emails[rawAccount.Email]
	if ok && emailID != id {
		store.rwLock.RUnlock()
		return nil, errors.New("Same email already taken")
	}
	account, ok := store.accounts[id]
	store.rwLock.RUnlock()
	if !ok {
		return nil, errors.New("Unknwon account for update")
	}
	if rawAccount.Premium != nil {
		account.Premium = rawAccount.Premium
		if store.PremiumNow(account) {
			for _, interest := range account.Interests {
				store.indexInterestPremium.Add(interest, account.Status, account.Sex, account.City, account.Country, account.ID)
				if account.Status == StatusSingle {
					store.indexInterestSingle.Remove(interest, account.City, account.Country, account.ID)
				} else if account.Status == StatusComplicated {
					store.indexInterestComplicated.Remove(interest, account.City, account.Country, account.ID)
				} else if account.Status == StatusRelationship {
					store.indexInterestRelationship.Remove(interest, account.City, account.Country, account.ID)
				}
			}
		} else {
			for _, interest := range account.Interests {
				store.indexInterestPremium.Remove(interest, account.Status, account.Sex, account.City, account.Country, account.ID)
				if account.Status == StatusSingle {
					store.indexInterestSingle.Add(interest, account.City, account.Country, account.ID)
				} else if account.Status == StatusComplicated {
					store.indexInterestComplicated.Add(interest, account.City, account.Country, account.ID)
				} else if account.Status == StatusRelationship {
					store.indexInterestRelationship.Add(interest, account.City, account.Country, account.ID)
				}
			}
		}
	}
	if rawAccount.Sex != 0 {
		oldSex := account.Sex
		account.Sex = rawAccount.Sex
		store.indexSex.Remove(oldSex, account.ID)
		store.indexSex.Add(account.Sex, account.ID)
	}
	if rawAccount.Status != 0 {
		oldStatus := account.Status
		account.Status = rawAccount.Status
		store.indexStatus.Remove(oldStatus, account.ID)
		store.indexStatus.Add(account.Status, account.ID)
		for _, interest := range account.Interests {
			if store.PremiumNow(account) {
				store.indexInterestPremium.Remove(interest, oldStatus, account.Sex, account.City, account.Country, account.ID)
				store.indexInterestPremium.Add(interest, account.Status, account.Sex, account.City, account.Country, account.ID)
			} else {
				if oldStatus == StatusSingle {
					store.indexInterestSingle.Remove(interest, account.City, account.Country, account.ID)
				} else if oldStatus == StatusComplicated {
					store.indexInterestComplicated.Remove(interest, account.City, account.Country, account.ID)
				} else if oldStatus == StatusRelationship {
					store.indexInterestRelationship.Remove(interest, account.City, account.Country, account.ID)
				}
				if account.Status == StatusSingle {
					store.indexInterestSingle.Add(interest, account.City, account.Country, account.ID)
				} else if account.Status == StatusComplicated {
					store.indexInterestComplicated.Add(interest, account.City, account.Country, account.ID)
				} else if account.Status == StatusRelationship {
					store.indexInterestRelationship.Add(interest, account.City, account.Country, account.ID)
				}
			}
		}
	}
	if rawAccount.Birth != 0 && account.Birth != rawAccount.Birth {
		oldBirth := rawAccount.Birth
		account.Birth = rawAccount.Birth
		store.indexBirthYear.Remove(timestampToYear(oldBirth), ID(account.ID))
		newYear := timestampToYear(account.Birth)
		store.indexBirthYear.Add(newYear, ID(account.ID))
	}
	if rawAccount.Joined != 0 {
		oldBirth := rawAccount.Birth
		account.Birth = rawAccount.Birth
		store.indexBirthYear.Remove(timestampToYear(oldBirth), ID(account.ID))
		newYear := timestampToYear(account.Birth)
		store.indexBirthYear.Add(newYear, ID(account.ID))
	}
	if rawAccount.Email != "" && account.Email != rawAccount.Email {
		oldEmail := account.Email
		account.Email = rawAccount.Email
		account.EmailDomain = rawAccount.EmailDomain
		store.rwLock.Lock()
		delete(store.emails, oldEmail)
		store.emails[account.Email] = account.ID
		store.rwLock.Unlock()
	}
	if rawAccount.Phone != nil && (account.Phone == nil || *account.Phone != *rawAccount.Phone) {
		account.Phone = rawAccount.Phone
		if account.PhoneCode != nil {
			store.indexPhoneCode.Remove(*account.PhoneCode, account.ID)
		}
		account.PhoneCode = rawAccount.PhoneCode
		store.indexPhoneCode.Add(*account.PhoneCode, account.ID)
	}
	if len(rawAccount.Likes) > 0 {
		for _, like := range rawAccount.Likes {
			store.indexLikee.Add(ID(like.ID), account.ID)
		}
	}
	if rawAccount.Fname != nil {
		oldFname := account.Fname
		account.Fname = store.dicts.AddFname(*rawAccount.Fname)
		store.indexFname.Remove(oldFname, account.ID)
		store.indexFname.Add(account.Fname, account.ID)
	}
	if rawAccount.Sname != nil {
		account.Sname = store.dicts.AddSname(*rawAccount.Sname)
	}
	if rawAccount.Country != nil {
		oldCountry := account.Country
		account.Country = store.dicts.AddCountry(*rawAccount.Country)
		store.indexCountry.Remove(oldCountry, ID(account.ID))
		store.indexCountry.Add(account.Country, ID(account.ID))
		for _, interest := range account.Interests {
			if store.PremiumNow(account) {
				store.indexInterestPremium.RemoveCountry(interest, oldCountry, account.ID)
				store.indexInterestPremium.AddCountry(interest, account.Country, account.ID)
			} else {
				if account.Status == StatusSingle {
					store.indexInterestSingle.RemoveCountry(interest, oldCountry, account.ID)
					store.indexInterestSingle.AddCountry(interest, account.Country, account.ID)
				} else if account.Status == StatusComplicated {
					store.indexInterestComplicated.RemoveCountry(interest, oldCountry, account.ID)
					store.indexInterestComplicated.AddCountry(interest, account.Country, account.ID)
				} else if account.Status == StatusRelationship {
					store.indexInterestRelationship.RemoveCountry(interest, oldCountry, account.ID)
					store.indexInterestRelationship.AddCountry(interest, account.Country, account.ID)
				}
			}
		}
	}
	if rawAccount.City != nil {
		oldCity := account.City
		account.City = store.dicts.AddCity(account.Country, *rawAccount.City)
		store.indexCity.Remove(oldCity, account.ID)
		store.indexCity.Add(account.City, account.ID)
		for _, interest := range account.Interests {
			if store.PremiumNow(account) {
				store.indexInterestPremium.RemoveCity(interest, oldCity, account.ID)
				store.indexInterestPremium.AddCity(interest, account.City, account.ID)
			} else {
				if account.Status == StatusSingle {
					store.indexInterestSingle.RemoveCity(interest, oldCity, account.ID)
					store.indexInterestSingle.AddCity(interest, account.City, account.ID)
				} else if account.Status == StatusComplicated {
					store.indexInterestComplicated.RemoveCity(interest, oldCity, account.ID)
					store.indexInterestComplicated.AddCity(interest, account.City, account.ID)
				} else if account.Status == StatusRelationship {
					store.indexInterestRelationship.RemoveCity(interest, oldCity, account.ID)
					store.indexInterestRelationship.AddCity(interest, account.City, account.ID)
				}
			}
		}
	}
	// TODO: may be empty interests list
	if len(rawAccount.Interests) > 0 {
		for _, interest := range account.Interests {
			store.indexInterest.Remove(interest, account.ID)
			if store.PremiumNow(account) {
				store.indexInterestPremium.Remove(interest, account.Status, account.Sex, account.City, account.Country, account.ID)
			} else {
				if account.Status == StatusSingle {
					store.indexInterestSingle.Remove(interest, account.City, account.Country, account.ID)
				} else if account.Status == StatusComplicated {
					store.indexInterestComplicated.Remove(interest, account.City, account.Country, account.ID)
				} else if account.Status == StatusRelationship {
					store.indexInterestRelationship.Remove(interest, account.City, account.Country, account.ID)
				}
			}
		}
		account.Interests = account.Interests[:0]
		for _, interestStr := range rawAccount.Interests {
			interest := store.dicts.AddInterest(interestStr)
			account.Interests = append(account.Interests, interest)
			store.indexInterest.Add(interest, account.ID)
			if store.PremiumNow(account) {
				store.indexInterestPremium.Add(interest, account.Status, account.Sex, account.City, account.Country, account.ID)
			} else {
				if account.Status == StatusSingle {
					store.indexInterestSingle.Add(interest, account.City, account.Country, account.ID)
				} else if account.Status == StatusComplicated {
					store.indexInterestComplicated.Add(interest, account.City, account.Country, account.ID)
				} else if account.Status == StatusRelationship {
					store.indexInterestRelationship.Add(interest, account.City, account.Country, account.ID)
				}
			}
		}
	}

	return account, nil
}

func (store *Store) PremiumNow(account *Account) bool {
	if account.Premium == nil {
		return false
	}
	return account.Premium.Start < store.now && store.now < account.Premium.Finish
}

func (store *Store) Get(id ID) (*Account, bool) {
	store.rwLock.RLock()
	account, ok := store.accounts[id]
	store.rwLock.RUnlock()
	return account, ok
}

func (store *Store) GetAll() map[ID]*Account {
	return store.accounts
}
