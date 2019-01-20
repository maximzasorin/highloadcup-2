package main

import (
	"sort"
	"sync"
)

const (
	StatusSingleString            = "свободны"
	StatusComplicatedString       = "всё сложно"
	StatusRelationshipString      = "заняты"
	StatusComplicated        byte = iota + 1
	StatusRelationship
	StatusSingle
)

const (
	SexFemale = byte('f')
	SexMale   = byte('m')
)

type Premium struct {
	Start  uint32
	Finish uint32
}

type AccountLike struct {
	ID ID
	Ts uint32
}

type AccountLikes []AccountLike

type Account struct {
	ID          ID
	Sex         byte
	Status      byte
	Sname       Sname   // optional
	Fname       Fname   // optional
	Country     Country // optional
	City        City    // optional
	EmailDomain uint8
	PhoneCode   uint16 // optional
	Birth       int64
	Joined      uint32
	Phone       *string // optional
	Email       string
	Premium     *Premium // optional
	Interests   []Interest
	Likes       AccountLikes
	rwLock      sync.RWMutex
}

func (account *Account) AddLike(accountLike *AccountLike) {
	account.rwLock.Lock()
	account.Likes = append(account.Likes, *accountLike)
	sort.Sort(account.Likes)
	account.rwLock.Unlock()
}

func (account *Account) AppendLike(accountLike *AccountLike) {
	account.rwLock.Lock()
	account.Likes = append(account.Likes, *accountLike)
	account.rwLock.Unlock()
}

func (account *Account) SortLikes() {
	account.rwLock.Lock()
	sort.Sort(account.Likes)
	account.rwLock.Unlock()
}

type Like struct {
	Likee uint32
	Ts    uint32
	Liker uint32
}

type RawLike struct {
	ID uint32
	Ts uint32
}

type RawAccount struct {
	ID          uint32
	Sex         byte
	Status      byte
	Sname       *string // optional
	Fname       *string // optional
	Country     *string // optional
	City        *string // optional
	EmailDomain uint8
	Birth       int64
	Joined      uint32
	PhoneCode   *uint16 // optional
	Phone       *string // optional
	Email       string
	Premium     *Premium // optional
	Interests   []string
	Likes       []RawLike
}

func (al AccountLikes) Len() int {
	return len(al)
}

func (al AccountLikes) Swap(i, j int) {
	al[i], al[j] = al[j], al[i]
}

func (al AccountLikes) Less(i, j int) bool {
	return al[i].ID > al[j].ID
}
