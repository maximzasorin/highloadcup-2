package main

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

type Like struct {
	ID uint32
	Ts uint32
}

type RawLike struct {
	Likee uint32
	Ts    uint32
	Liker uint32
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
	Likes       []Like
}

type Account struct {
	ID          uint32
	Sex         byte
	Status      byte
	Sname       Sname   // optional
	Fname       Fname   // optional
	Country     Country // optional
	City        City    // optional
	EmailDomain uint8
	Birth       int64
	Joined      uint32
	PhoneCode   *uint16 // optional
	Phone       *string // optional
	Email       string
	Premium     *Premium // optional
	Interests   []Interest
	Likes       []Like
}
