package main

import (
	"sync"

	"github.com/pkg/errors"
)

const (
	storePreallocCount = 1000*1000 + 330*1000
)

type Store struct {
	parser      *Parser
	dicts       *Dicts
	count       ID
	now         uint32
	rating      bool
	accountsMap map[ID]*Account
	accountsArr []Account
	emails      map[string]ID
	rwLock      sync.RWMutex
	index       *Index
}

func NewStore(dicts *Dicts, now uint32, rating bool) *Store {
	store := &Store{
		dicts:       dicts,
		now:         now,
		rating:      rating,
		accountsMap: make(map[ID]*Account),
		accountsArr: make([]Account, storePreallocCount),
		emails:      make(map[string]ID),
	}
	store.index = NewIndex(store, dicts)
	return store
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

	store.rwLock.Lock()
	if check {
		if _, ok := store.emails[rawAccount.Email]; ok {
			store.rwLock.Unlock()
			return nil, errors.New("Same email already taken")
		}
		if store.get(ID(rawAccount.ID)) != nil {
			store.rwLock.Unlock()
			return nil, errors.New("Account with same ID already exists")
		}
	}
	var account *Account
	if rawAccount.ID < storePreallocCount {
		account = &store.accountsArr[rawAccount.ID]
		account.ID = ID(rawAccount.ID)
		account.Sex = rawAccount.Sex
		account.Status = rawAccount.Status
		account.Birth = rawAccount.Birth
		account.Joined = rawAccount.Joined
		account.Premium = rawAccount.Premium
		account.Email = rawAccount.Email
		account.EmailDomain = rawAccount.EmailDomain
	} else {
		store.accountsMap[ID(rawAccount.ID)] = &Account{
			ID:          ID(rawAccount.ID),
			Sex:         rawAccount.Sex,
			Status:      rawAccount.Status,
			Birth:       rawAccount.Birth,
			Joined:      rawAccount.Joined,
			Premium:     rawAccount.Premium,
			Email:       rawAccount.Email,
			EmailDomain: rawAccount.EmailDomain,
		}
		account = store.accountsMap[ID(rawAccount.ID)]
	}
	store.emails[account.Email] = account.ID
	store.rwLock.Unlock()

	if rawAccount.Phone != nil {
		account.Phone = rawAccount.Phone
		account.PhoneCode = *rawAccount.PhoneCode
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
	// if len(rawAccount.Likes) > 0 {
	// 	for _, like := range rawAccount.Likes {
	// 		account.AppendLike(&AccountLike{
	// 			ID: ID(like.ID),
	// 			Ts: like.Ts,
	// 		})
	// 	}
	// 	account.SortLikes()
	// }
	for _, interestStr := range rawAccount.Interests {
		interest := store.dicts.AddInterest(interestStr)
		account.Interests = append(account.Interests, interest)
	}
	if updateIndexes {
		// batch := BorrowIndexBatch(store.index)
		batch := &IndexBatch{index: store.index}
		batch.Add(account)
		batch.AddInterests(account.ID, account.Status, account.Sex, account.City, account.Country, store.PremiumNow(account), account.Interests...)
		batch.AddGroupHash(CreateHashFromAccount(account), account.Interests...)
		for _, like := range rawAccount.Likes {
			batch.AddLike(account.ID, ID(like.ID), like.Ts)
		}
		store.index.worker.Add(batch.Dispatch())
	} else {
		// likes add always to index
		for _, like := range rawAccount.Likes {
			store.index.AppendLike(account.ID, ID(like.ID), like.Ts)
		}
	}
	store.count++

	return account, nil
}

func (store *Store) AddLikes(likes *Likes, updateIndexes bool) error {
	store.rwLock.RLock()
	for _, like := range likes.likes {
		if store.get(ID(like.Likee)) == nil {
			store.rwLock.RUnlock()
			return errors.New("Cannot find likee account")
		}
		if store.get(ID(like.Liker)) == nil {
			store.rwLock.RUnlock()
			return errors.New("Cannot find liker account")
		}
	}
	store.rwLock.RUnlock()

	// batch := BorrowIndexBatch(store.index)
	batch := &IndexBatch{index: store.index}
	for _, likes := range likes.likes {
		batch.AddLike(ID(likes.Liker), ID(likes.Likee), likes.Ts)
	}
	store.index.worker.Add(batch.Dispatch())

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
	account := store.get(id)
	if account == nil {
		store.rwLock.RUnlock()
		return nil, errors.New("Unknwon account for update")
	}
	store.rwLock.RUnlock()

	oldHash := CreateHashFromAccount(account)
	oldInts := make([]Interest, len(account.Interests))
	for i, interest := range account.Interests {
		oldInts[i] = interest
	}

	// batch := BorrowIndexBatch(store.index)
	batch := &IndexBatch{index: store.index}

	if rawAccount.Premium != nil {
		account.Premium = rawAccount.Premium
		if store.PremiumNow(account) {
			for _, interest := range account.Interests {
				// batch.InterestPremium.Add(interest, account.Status, account.Sex, account.City, account.Country, account.ID)
				batch.AddInterestPremium(account.ID, interest, account.Status, account.Sex, account.City, account.Country)
				if account.Status == StatusSingle {
					// batch.InterestSingle.Remove(interest, account.City, account.Country, account.ID)
					batch.RemoveInterestSingle(account.ID, interest, account.City, account.Country)
				} else if account.Status == StatusComplicated {
					// batch.InterestComplicated.Remove(interest, account.City, account.Country, account.ID)
					batch.RemoveInterestComplicated(account.ID, interest, account.City, account.Country)
				} else if account.Status == StatusRelationship {
					// batch.InterestRelationship.Remove(interest, account.City, account.Country, account.ID)
					batch.RemoveInterestRelationship(account.ID, interest, account.City, account.Country)
				}
			}
		} else {
			for _, interest := range account.Interests {
				batch.RemoveInterestPremium(account.ID, interest, account.Status, account.Sex, account.City, account.Country)
				// batch.InterestPremium.Remove(interest, account.Status, account.Sex, account.City, account.Country, account.ID)
				if account.Status == StatusSingle {
					// batch.InterestSingle.Add(interest, account.City, account.Country, account.ID)
					batch.AddInterestSingle(account.ID, interest, account.City, account.Country)
				} else if account.Status == StatusComplicated {
					// batch.InterestComplicated.Add(interest, account.City, account.Country, account.ID)
					batch.AddInterestComplicated(account.ID, interest, account.City, account.Country)
				} else if account.Status == StatusRelationship {
					// batch.InterestRelationship.Add(interest, account.City, account.Country, account.ID)
					batch.AddInterestRelationship(account.ID, interest, account.City, account.Country)
				}
			}
		}
	}
	if rawAccount.Sex != 0 {
		oldSex := account.Sex
		account.Sex = rawAccount.Sex

		batch.ReplaceSex(account.ID, oldSex, account.Sex)

		// batch.Sex.Remove(oldSex, account.ID)
		// batch.Sex.Add(account.Sex, account.ID)
	}
	if rawAccount.Status != 0 {
		oldStatus := account.Status
		account.Status = rawAccount.Status
		batch.ReplaceStatus(account.ID, oldStatus, account.Status)
		// batch.Status.Remove(oldStatus, account.ID)
		// batch.Status.Add(account.Status, account.ID)
		for _, interest := range account.Interests {
			if store.PremiumNow(account) {
				batch.ReplaceInterestPremiumStatus(account.ID, interest, account.Sex, account.City, account.Country, oldStatus, account.Status)
				// batch.InterestPremium.Remove(interest, oldStatus, account.Sex, account.City, account.Country, account.ID)
				// batch.InterestPremium.Add(interest, account.Status, account.Sex, account.City, account.Country, account.ID)
			} else {
				if oldStatus == StatusSingle {
					// batch.InterestSingle.Remove(interest, account.City, account.Country, account.ID)
					batch.RemoveInterestSingle(account.ID, interest, account.City, account.Country)
				} else if oldStatus == StatusComplicated {
					// batch.InterestComplicated.Remove(interest, account.City, account.Country, account.ID)
					batch.RemoveInterestComplicated(account.ID, interest, account.City, account.Country)
				} else if oldStatus == StatusRelationship {
					// batch.InterestRelationship.Remove(interest, account.City, account.Country, account.ID)
					batch.RemoveInterestRelationship(account.ID, interest, account.City, account.Country)
				}
				if account.Status == StatusSingle {
					// batch.InterestSingle.Add(interest, account.City, account.Country, account.ID)
					batch.AddInterestSingle(account.ID, interest, account.City, account.Country)
				} else if account.Status == StatusComplicated {
					// batch.InterestComplicated.Add(interest, account.City, account.Country, account.ID)
					batch.AddInterestComplicated(account.ID, interest, account.City, account.Country)
				} else if account.Status == StatusRelationship {
					// batch.InterestRelationship.Add(interest, account.City, account.Country, account.ID)
					batch.AddInterestRelationship(account.ID, interest, account.City, account.Country)
				}
			}
		}
	}
	if rawAccount.Birth != 0 && account.Birth != rawAccount.Birth {
		oldBirth := rawAccount.Birth
		account.Birth = rawAccount.Birth
		// batch.BirthYear.Remove(timestampToYear(oldBirth), ID(account.ID))
		// newYear := timestampToYear(rawAccount.Birth)
		// batch.BirthYear.Add(newYear, ID(account.ID))
		batch.ReplaceBirth(account.ID, timestampToYear(oldBirth), timestampToYear(rawAccount.Birth))
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
		oldPhoneCode := account.PhoneCode
		// if account.PhoneCode != 0 {
		// 	batch.PhoneCode.Remove(account.PhoneCode, account.ID)
		// }
		account.PhoneCode = *rawAccount.PhoneCode
		// batch.PhoneCode.Add(account.PhoneCode, account.ID)
		if oldPhoneCode != 0 {
			batch.ReplacePhoneCode(account.ID, oldPhoneCode, account.PhoneCode)
		} else {
			batch.AddPhoneCode(account.ID, account.PhoneCode)
		}
	}
	if rawAccount.Fname != nil {
		oldFname := account.Fname
		account.Fname = store.dicts.AddFname(*rawAccount.Fname)
		batch.ReplaceFname(account.ID, oldFname, account.Fname)
		// batch.Fname.Remove(oldFname, account.ID)
		// batch.Fname.Add(account.Fname, account.ID)
	}
	if rawAccount.Sname != nil {
		account.Sname = store.dicts.AddSname(*rawAccount.Sname)
	}
	if rawAccount.Country != nil {
		oldCountry := account.Country
		account.Country = store.dicts.AddCountry(*rawAccount.Country)
		batch.ReplaceCountry(account.ID, oldCountry, account.Country)
		// batch.Country.Remove(oldCountry, ID(account.ID))
		// batch.Country.Add(account.Country, ID(account.ID))
		for _, interest := range account.Interests {
			if store.PremiumNow(account) {
				batch.ReplaceInterestPremiumCountry(account.ID, interest, oldCountry, account.Country)
				// batch.InterestPremium.RemoveCountry(interest, oldCountry, account.ID)
				// batch.InterestPremium.AddCountry(interest, account.Country, account.ID)
			} else {
				if account.Status == StatusSingle {
					// batch.InterestSingle.RemoveCountry(interest, oldCountry, account.ID)
					// batch.InterestSingle.AddCountry(interest, account.Country, account.ID)
					batch.ReplaceInterestSingleCountry(account.ID, interest, oldCountry, account.Country)
				} else if account.Status == StatusComplicated {
					// batch.InterestComplicated.RemoveCountry(interest, oldCountry, account.ID)
					// batch.InterestComplicated.AddCountry(interest, account.Country, account.ID)
					batch.ReplaceInterestComplicatedCountry(account.ID, interest, oldCountry, account.Country)
				} else if account.Status == StatusRelationship {
					// batch.InterestRelationship.RemoveCountry(interest, oldCountry, account.ID)
					// batch.InterestRelationship.AddCountry(interest, account.Country, account.ID)
					batch.ReplaceInterestRelationshipCountry(account.ID, interest, oldCountry, account.Country)
				}
			}
		}
	}
	if rawAccount.City != nil {
		oldCity := account.City
		account.City = store.dicts.AddCity(account.Country, *rawAccount.City)
		// batch.City.Remove(oldCity, account.ID)
		// batch.City.Add(account.City, account.ID)
		batch.ReplaceCity(account.ID, oldCity, account.City)
		for _, interest := range account.Interests {
			if store.PremiumNow(account) {
				// batch.InterestPremium.RemoveCity(interest, oldCity, account.ID)
				// batch.InterestPremium.AddCity(interest, account.City, account.ID)
				batch.ReplaceInterestPremiumCity(account.ID, interest, oldCity, account.City)
			} else {
				if account.Status == StatusSingle {
					// batch.InterestSingle.RemoveCity(interest, oldCity, account.ID)
					// batch.InterestSingle.AddCity(interest, account.City, account.ID)
					batch.ReplaceInterestSingleCity(account.ID, interest, oldCity, account.City)
				} else if account.Status == StatusComplicated {
					// batch.InterestComplicated.RemoveCity(interest, oldCity, account.ID)
					// batch.InterestComplicated.AddCity(interest, account.City, account.ID)
					batch.ReplaceInterestComplicatedCity(account.ID, interest, oldCity, account.City)
				} else if account.Status == StatusRelationship {
					// batch.InterestRelationship.RemoveCity(interest, oldCity, account.ID)
					// batch.InterestRelationship.AddCity(interest, account.City, account.ID)
					batch.ReplaceInterestRelationshipCity(account.ID, interest, oldCity, account.City)
				}
			}
		}
	}
	// TODO: may be empty interests list
	if len(rawAccount.Interests) > 0 {
		for _, interest := range account.Interests {
			// batch.Interest.Remove(interest, account.ID)
			batch.RemoveInterest(account.ID, interest)
			if store.PremiumNow(account) {
				// batch.InterestPremium.Remove(interest, account.Status, account.Sex, account.City, account.Country, account.ID)
				batch.RemoveInterestPremium(account.ID, interest, account.Status, account.Sex, account.City, account.Country)
			} else {
				if account.Status == StatusSingle {
					// batch.InterestSingle.Remove(interest, account.City, account.Country, account.ID)
					batch.RemoveInterestSingle(account.ID, interest, account.City, account.Country)
				} else if account.Status == StatusComplicated {
					// batch.InterestComplicated.Remove(interest, account.City, account.Country, account.ID)
					batch.RemoveInterestComplicated(account.ID, interest, account.City, account.Country)
				} else if account.Status == StatusRelationship {
					// batch.InterestRelationship.Remove(interest, account.City, account.Country, account.ID)
					batch.RemoveInterestRelationship(account.ID, interest, account.City, account.Country)
				}
			}
		}

		account.Interests = account.Interests[:0]

		for _, interestStr := range rawAccount.Interests {
			interest := store.dicts.AddInterest(interestStr)
			account.Interests = append(account.Interests, interest)
			// batch.Interest.Add(interest, account.ID)
			batch.AddInterest(account.ID, interest)
			if store.PremiumNow(account) {
				// batch.InterestPremium.Add(interest, account.Status, account.Sex, account.City, account.Country, account.ID)
				batch.AddInterestPremium(account.ID, interest, account.Status, account.Sex, account.City, account.Country)
			} else {
				if account.Status == StatusSingle {
					// batch.InterestSingle.Add(interest, account.City, account.Country, account.ID)
					batch.AddInterestSingle(account.ID, interest, account.City, account.Country)
				} else if account.Status == StatusComplicated {
					// batch.InterestComplicated.Add(interest, account.City, account.Country, account.ID)
					batch.AddInterestComplicated(account.ID, interest, account.City, account.Country)
				} else if account.Status == StatusRelationship {
					// batch.InterestRelationship.Add(interest, account.City, account.Country, account.ID)
					batch.AddInterestRelationship(account.ID, interest, account.City, account.Country)
				}
			}
		}
	}

	batch.SubGroupHash(oldHash, oldInts...)
	batch.AddGroupHash(CreateHashFromAccount(account), account.Interests...)

	store.index.worker.Add(batch.Dispatch())

	// store.index.AddGroupHash(CreateHashFromAccount(account), account.Interests...)

	// go func() {
	// 	store.index.Group.SubHash(oldHash, oldInts...)
	// 	store.index.Group.AddHash(CreateHashFromAccount(account), account.Interests...)
	// }()
	// store.index.Group.Add(account)

	return account, nil
}

func (store *Store) PremiumNow(account *Account) bool {
	if account.Premium == nil {
		return false
	}
	return account.Premium.Start < store.now && store.now < account.Premium.Finish
}

func (store *Store) Get(id ID) *Account {
	store.rwLock.RLock()
	account := store.get(id)
	store.rwLock.RUnlock()
	return account
}

func (store *Store) Count() int {
	return int(store.count)
}

func (store *Store) get(id ID) *Account {
	if id < storePreallocCount {
		if store.accountsArr[id].ID == id {
			return &store.accountsArr[id]
		}
	} else {
		if account, ok := store.accountsMap[id]; ok {
			return account
		}
	}
	return nil
}

func (store *Store) Iterate(iterator func(*Account) bool) {
	for _, account := range store.accountsArr {
		if account.ID == 0 {
			continue
		}
		if !iterator(&account) {
			return
		}
	}
	for _, account := range store.accountsMap {
		if !iterator(account) {
			return
		}
	}
}
