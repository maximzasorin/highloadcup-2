package main

import "errors"

type IndexEmail struct {
	emails map[string]uint32
}

func NewIndexEmail() *IndexEmail {
	return &IndexEmail{
		emails: make(map[string]uint32),
	}
}

func (indexEmail *IndexEmail) Add(email string, ID uint32) {
	indexEmail.emails[email] = ID
}

func (indexEmail *IndexEmail) Remove(email string) {
	if indexEmail.Has(email) {
		delete(indexEmail.emails, email)
	}
}

func (indexEmail *IndexEmail) Get(email string) (uint32, error) {
	ID, has := indexEmail.emails[email]
	if has {
		return ID, nil
	} else {
		return 0, errors.New("Can not find ID by email")
	}
}

func (indexEmail *IndexEmail) Has(email string) bool {
	_, has := indexEmail.emails[email]
	return has
}
