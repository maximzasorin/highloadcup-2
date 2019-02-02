package main

import (
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

var rawAccountsPool = sync.Pool{
	New: func() interface{} {
		return &RawAccount{
			Interests: make([]string, 0, 16),
			Likes:     make([]RawLike, 0, 32),
		}
	},
}

func BorrowRawAccount() *RawAccount {
	rawAccount := rawAccountsPool.Get().(*RawAccount)
	rawAccount.Reset()
	return rawAccount
}

func (account *RawAccount) Reset() {
	account.ID = 0
	account.Sex = 0
	account.Status = 0
	account.Sname = nil
	account.Fname = nil
	account.Country = nil
	account.City = nil
	account.EmailDomain = 0
	account.Birth = 0
	account.Joined = 0
	account.PhoneCode = nil
	account.Phone = nil
	account.Email = ""
	account.Premium = nil
	account.Interests = account.Interests[:0]
	account.Likes = account.Likes[:0]
}

func (account *RawAccount) Release() {
	rawAccountsPool.Put(account)
}

type Likes struct {
	likes []Like
	len   int
}

const maxLikesLen = 128

var likesPool = sync.Pool{
	New: func() interface{} {
		return &Likes{
			likes: make([]Like, maxLikesLen),
		}
	},
}

func BorrowLikes() *Likes {
	l := likesPool.Get().(*Likes)
	l.Reset()
	return l
}

func (likes *Likes) Truncate() {
	likes.likes = likes.likes[:likes.len]
}

func (likes *Likes) Reset() {
	likes.likes = likes.likes[:maxLikesLen]
	likes.len = 0
}

func (likes *Likes) Release() {
	likesPool.Put(likes)
}

var idsPool = sync.Pool{
	New: func() interface{} {
		ids := make(IDS, 0, 4*1024)
		return &ids
	},
}

func BorrowIDS() *IDS {
	ids := idsPool.Get().(*IDS)
	ids.Reset()
	return ids
}

func (ids *IDS) Reset() {
	*ids = (*ids)[:0]
}

func (ids *IDS) Release() {
	idsPool.Put(ids)
}
