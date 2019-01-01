package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/francoispqt/gojay"
)

const (
	StatusSingleString            = "свободны"
	StatusComplicatedString       = "всё сложно"
	StatusRelationshipString      = "заняты"
	StatusSingle             byte = iota
	StatusRelationship
	StatusComplicated
)

const (
	SexFemale = byte('f')
	SexMale   = byte('m')
)

type Premium struct {
	Start  uint32
	Finish uint32
}

type Like struct {
	ID uint32
	Ts uint32
}

type Account struct {
	ID        uint32
	Sex       byte
	Status    byte
	Birth     uint32
	Joined    uint32
	Premium   *Premium // optional
	Fname     *string  // optional
	Sname     *string  // optional
	Phone     *string  // optional
	Country   *string  // optional
	City      *string  // optional
	Email     string
	Interests []string
	Likes     []Like
}

type Store struct {
	now      uint32
	test     bool
	accounts map[uint32]*Account
}

func NewStore() *Store {
	return &Store{
		accounts: make(map[uint32]*Account),
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

func (store *Store) Add(account *Account) error {
	if _, ok := store.accounts[account.ID]; ok {
		return errors.New("Account with same ID already exists")
	}

	store.accounts[account.ID] = account

	return nil
}

func (store *Store) Get(id uint32) *Account {
	return store.accounts[id]
}

func (store *Store) FilterAll(filter *Filter, writer http.ResponseWriter) {
	// for ID, account := range store.accounts {

	// }

	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("{\"accounts\":[]}"))
}

func (account *Account) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "id":
		return dec.Uint32(&account.ID)
	case "sex":
		var sex string
		err := readString(dec, &sex, false)
		if err != nil {
			return err
		}
		if len(sex) != 1 || (sex[0] != SexFemale && sex[0] != SexMale) {
			return errors.New("Invalid sex value")
		}
		account.Sex = sex[0]
		return nil
	case "status":
		var status string
		err := readString(dec, &status, true)
		if err != nil {
			return err
		}
		switch status {
		case StatusSingleString:
			account.Status = StatusSingle
		case StatusComplicatedString:
			account.Status = StatusComplicated
		case StatusRelationshipString:
			account.Status = StatusRelationship
		default:
			return errors.New("Unknown account status")
		}
		return nil
	case "birth":
		return dec.Uint32(&account.Birth)
	case "joined":
		return dec.Uint32(&account.Joined)
	case "premium":
		err := dec.ObjectNull(&account.Premium)
		if err != nil {
			return err
		}
		return nil
	case "email":
		return readString(dec, &account.Email, false)
	case "phone":
		return readStringNull(dec, &account.Phone, false)
	case "fname":
		return readStringNull(dec, &account.Fname, true)
	case "sname":
		return readStringNull(dec, &account.Sname, true)
	case "country":
		return readStringNull(dec, &account.Country, true)
	case "city":
		return readStringNull(dec, &account.City, true)
	case "likes":
		return dec.Array(gojay.DecodeArrayFunc(func(dec *gojay.Decoder) error {
			like := Like{}
			err := dec.Object(&like)
			if err != nil {
				return err
			}
			account.Likes = append(account.Likes, like)
			return nil
		}))
	case "interests":
		return dec.Array(gojay.DecodeArrayFunc(func(dec *gojay.Decoder) error {
			var interest string
			err := readString(dec, &interest, true)
			if err != nil {
				return err
			}
			account.Interests = append(account.Interests, interest)
			return nil
		}))
	}
	return errors.New(`Unknown account field "` + key + `"`)
}

func (account *Account) NKeys() int {
	return 14
}

func (premium *Premium) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "start":
		return dec.Uint32(&premium.Start)
	case "finish":
		return dec.Uint32(&premium.Finish)
	}
	return errors.New("Unknown premium field")
}

func (premium *Premium) NKeys() int {
	return 2
}

func (like *Like) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "id":
		return dec.Uint32(&like.ID)
	case "ts":
		return dec.Uint32(&like.Ts)
	}
	return errors.New("Unknown like field")
}

func (like *Like) NKeys() int {
	return 2
}

func readString(dec *gojay.Decoder, str *string, unquote bool) error {
	var buf []byte
	err := dec.EmbeddedJSON((*gojay.EmbeddedJSON)(&buf))
	if err != nil {
		return err
	}
	if unquote {
		*str, err = strconv.Unquote(string(buf))
		if err != nil {
			return err
		}
	} else {
		s := string(buf)
		*str = s[1 : len(s)-1]
	}

	return nil
}

func readStringNull(dec *gojay.Decoder, str **string, unquote bool) error {
	var buf []byte
	err := dec.EmbeddedJSON((*gojay.EmbeddedJSON)(&buf))
	if err != nil {
		return err
	}
	if buf[0] == 'n' {
		*str = nil
		return nil
	}
	// if unquote {
	s, err := strconv.Unquote(string(buf))
	*str = &s
	if err != nil {
		return err
	}
	// } else {
	// 	s := string(buf)
	// 	unq := s[1 : len(s)-1]
	// 	*str = &unq
	// }

	return nil
}
