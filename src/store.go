package main

import (
	"errors"
)

type Accounts map[uint32]*Account

type AccountSlice []*Account

type Store struct {
	parser         *Parser
	dicts          *Dicts
	now            uint32
	test           bool
	withPremium    uint32
	countLikes     uint64
	accounts       Accounts
	indexID        *IndexID
	indexEmail     *IndexEmail
	indexLikee     *IndexLikee
	indexInterest  *IndexInterest
	indexCity      *IndexCity
	indexBirthYear *IndexYear
	indexCountry   *IndexCountry
	indexFname     *IndexFname
	indexPhoneCode *IndexPhoneCode
}

func NewStore(parser *Parser, dicts *Dicts) *Store {
	return &Store{
		parser:         parser,
		dicts:          dicts,
		accounts:       make(Accounts),
		indexID:        NewIndexID(),
		indexEmail:     NewIndexEmail(),
		indexLikee:     NewIndexLikee(),
		indexInterest:  NewIndexInterest(),
		indexCity:      NewIndexCity(),
		indexBirthYear: NewIndexYear(),
		indexCountry:   NewIndexCountry(),
		indexFname:     NewIndexFname(),
		indexPhoneCode: NewIndexPhoneCode(),
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
		if _, ok := store.accounts[rawAccount.ID]; ok {
			return nil, errors.New("Account with same ID already exists")
		}
		if store.indexEmail.Has(rawAccount.Email) {
			return nil, errors.New("Same email already taken")
		}
	}

	account := Account{
		ID:          rawAccount.ID,
		Sex:         rawAccount.Sex,
		Status:      rawAccount.Status,
		Birth:       rawAccount.Birth,
		Joined:      rawAccount.Joined,
		Premium:     rawAccount.Premium,
		Email:       rawAccount.Email,
		EmailDomain: rawAccount.EmailDomain,
		// Phone:       rawAccount.Phone,
		// PhoneCode:   rawAccount.PhoneCode,
		// Likes:       rawAccount.Likes,
	}

	if rawAccount.Phone != nil {
		account.Phone = rawAccount.Phone
		account.PhoneCode = rawAccount.PhoneCode
		store.indexPhoneCode.Add(*account.PhoneCode, account.ID)
	} else {
		store.indexPhoneCode.Add(0, account.ID)
	}

	if rawAccount.Fname != nil {
		account.Fname = store.dicts.AddFname(*rawAccount.Fname)
		store.indexFname.Add(account.Fname, account.ID)
	} else {
		store.indexFname.Add(0, account.ID)
	}

	if rawAccount.Sname != nil {
		account.Sname = store.dicts.AddSname(*rawAccount.Sname)
	}

	if rawAccount.Country != nil {
		account.Country = store.dicts.AddCountry(*rawAccount.Country)
		store.indexCountry.Add(account.Country, account.ID)
	} else {
		store.indexCountry.Add(0, account.ID)
	}

	if rawAccount.City != nil {
		account.City = store.dicts.AddCity(account.Country, *rawAccount.City)
		store.indexCity.Add(account.City, account.ID)
	} else {
		store.indexCity.Add(0, account.ID)
	}

	for _, like := range rawAccount.Likes {
		store.indexLikee.Add(like.ID, account.ID)
		if updateIndexes {
			store.indexLikee.Update(like.ID)
		}
	}

	for _, interestStr := range rawAccount.Interests {
		interest := store.dicts.AddInterest(interestStr)
		account.Interests = append(account.Interests, interest)
		store.indexInterest.Add(interest, account.ID)
		if updateIndexes {
			store.indexInterest.Update(interest)
		}
	}

	store.accounts[account.ID] = &account
	store.indexEmail.Add(account.Email, account.ID)

	store.indexID.Add(account.ID)
	store.indexBirthYear.Add(timestampToYear(account.Birth), account.ID)
	if updateIndexes {
		store.indexID.Update()
	}

	// if account.Premium != nil {
	// 	store.withPremium++
	// }

	// store.countLikes += uint64(len(account.Likes))

	return &account, nil
}

func (store *Store) AddLikes(rawLikes []*RawLike) error {
	// if len(rawLikes) == 0 {
	// 	return errors.New("No likes founded")
	// }

	for _, rawLike := range rawLikes {
		if _, ok := store.accounts[rawLike.Likee]; !ok {
			return errors.New("Cannot find likee account")
		}
		if _, ok := store.accounts[rawLike.Liker]; !ok {
			return errors.New("Cannot find liker account")
		}
	}

	for _, rawLike := range rawLikes {
		store.indexLikee.Add(rawLike.Likee, rawLike.Liker)
		// store.indexLikee.Update(rawLike.Likee)
	}

	return nil
}

func (store *Store) Update(ID uint32, rawAccount *RawAccount) (*Account, error) {
	if rawAccount.Email != "" && rawAccount.EmailDomain == 0 {
		return nil, errors.New("Invalid email")
	}
	emailID, err := store.indexEmail.Get(rawAccount.Email)
	if err == nil && emailID != ID {
		return nil, errors.New("Same email already taken")
	}

	account := store.Get(ID)
	if account == nil {
		return nil, errors.New("Unknwon account for update")
	}
	if rawAccount.Sex != 0 {
		account.Sex = rawAccount.Sex
	}
	if rawAccount.Status != 0 {
		account.Status = rawAccount.Status
	}
	if rawAccount.Birth != 0 {
		account.Birth = rawAccount.Birth
	}
	if rawAccount.Joined != 0 {
		account.Joined = rawAccount.Joined
	}
	if rawAccount.Premium != nil {
		account.Premium = rawAccount.Premium
	}
	if rawAccount.Email != "" && account.Email != rawAccount.Email {
		oldEmail := account.Email
		account.Email = rawAccount.Email
		account.EmailDomain = rawAccount.EmailDomain
		store.indexEmail.Remove(oldEmail)
		store.indexEmail.Add(account.Email, account.ID)
	}
	if rawAccount.Phone != nil {
		account.Phone = rawAccount.Phone
		account.PhoneCode = rawAccount.PhoneCode
	}

	if len(rawAccount.Likes) > 0 {
		for _, like := range rawAccount.Likes {
			account.Likes = append(account.Likes, like)
			store.indexLikee.Add(like.ID, account.ID)
			// store.indexLikee.Update(like.ID)
		}
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
	// TODO: may be empty interests list
	if len(rawAccount.Interests) > 0 {
		// for _, interest := range account.Interests {
		// 	store.indexInterest.Remove(interest, ID)
		// }
		account.Interests = account.Interests[:0]
		for _, interestStr := range rawAccount.Interests {
			interest := store.dicts.AddInterest(interestStr)
			account.Interests = append(account.Interests, interest)
			// store.indexInterest.Add(interest, account.ID)
			// store.indexInterest.Update(interest)
		}
	}

	return account, nil
}

func (store *Store) UpdateIndex() {
	store.indexID.Update()
	store.indexLikee.Update(0)
	store.indexInterest.Update(0)
	store.indexCity.Update(0)
	store.indexBirthYear.Update(0)
	store.indexCountry.Update(0)
	store.indexFname.Update(0)
	store.indexPhoneCode.Update(0)
}

func (store *Store) Get(id uint32) *Account {
	return store.accounts[id]
}

func (store *Store) GetAll() *Accounts {
	return &store.accounts
}
