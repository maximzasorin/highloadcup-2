package main

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/francoispqt/gojay"
)

type Parser struct {
	dicts *Dicts
}

type SerializeFields interface {
	Sex() bool
	Status() bool
	Fname() bool
	Sname() bool
	Phone() bool
	Country() bool
	City() bool
	Birth() bool
	Premium() bool
	Interests() bool
}

func NewParser(dicts *Dicts) *Parser {
	return &Parser{dicts}
}

func (parser *Parser) DecodeAccount(reader io.Reader, update bool) (*RawAccount, error) {
	dec := gojay.BorrowDecoder(reader)
	defer dec.Release()

	rawAccount := &RawAccount{}
	err := dec.Decode(parser.AccountDecodeFunc(rawAccount, update))
	if err != nil {
		return nil, err
	}

	return rawAccount, nil
}

func (parser *Parser) DecodeAccounts(reader io.Reader) ([]*RawAccount, error) {
	dec := gojay.BorrowDecoder(reader)
	defer dec.Release()

	rawAccounts := make([]*RawAccount, 0, 10000)

	err := dec.Decode(gojay.DecodeObjectFunc(func(dec *gojay.Decoder, key string) error {
		switch key {
		case "accounts":
			return dec.Array(gojay.DecodeArrayFunc(func(dec *gojay.Decoder) error {
				rawAccount := &RawAccount{}
				err := dec.Object(parser.AccountDecodeFunc(rawAccount, false))
				if err != nil {
					return err
				}
				rawAccounts = append(rawAccounts, rawAccount)
				return nil
			}))
		}
		return errors.New("Unknown key in accounts file")
	}))

	if err != nil {
		return make([]*RawAccount, 0), err
	}

	return rawAccounts, nil
}

func (parser *Parser) DecodeLikes(reader io.Reader) ([]*Like, error) {
	dec := gojay.BorrowDecoder(reader)
	defer dec.Release()

	likes := make([]*Like, 0)
	err := dec.Decode(gojay.DecodeObjectFunc(func(dec *gojay.Decoder, key string) error {
		switch key {
		case "likes":
			return dec.Array(gojay.DecodeArrayFunc(func(dec *gojay.Decoder) error {
				like := &Like{}
				err := dec.Object(parser.LikeDecodeFunc(like))
				if err != nil {
					return err
				}
				likes = append(likes, like)
				return nil
			}))
		}
		return errors.New(`Unknown likes field "` + key + `"`)
	}))

	if err != nil {
		return make([]*Like, 0), err
	}

	return likes, nil
}

func (parser *Parser) EncodeAccounts(accounts []*Account, fields SerializeFields) []byte {
	buffer := bytes.NewBuffer([]byte(``))

	enc := gojay.BorrowEncoder(buffer)
	defer enc.Release()

	enc.Encode(gojay.EncodeObjectFunc(func(enc *gojay.Encoder) {
		enc.AddArrayKey("accounts", gojay.EncodeArrayFunc(func(enc *gojay.Encoder) {
			for _, account := range accounts {
				enc.Object(parser.AccountEncodeFunc(account, fields))
			}
		}))
	}))

	return buffer.Bytes()
}

func (parser *Parser) EncodeGroupEntries(groupEntries []*GroupEntry, orderAsc bool) []byte {
	buffer := bytes.NewBuffer([]byte(``))

	enc := gojay.NewEncoder(buffer)
	defer enc.Release()

	enc.Encode(gojay.EncodeObjectFunc(func(enc *gojay.Encoder) {
		enc.AddArrayKey("groups", gojay.EncodeArrayFunc(func(enc *gojay.Encoder) {
			if orderAsc {
				for _, groupEntry := range groupEntries {
					enc.Object(parser.EncodeGroupFunc(groupEntry))
				}
			} else {
				for i := len(groupEntries) - 1; i >= 0; i-- {
					enc.Object(parser.EncodeGroupFunc(groupEntries[i]))
				}
			}
		}))
	}))

	return buffer.Bytes()
}

func (parser *Parser) AccountDecodeFunc(rawAccount *RawAccount, update bool) gojay.DecodeObjectFunc {
	return gojay.DecodeObjectFunc(func(dec *gojay.Decoder, key string) error {
		switch key {
		case "id":
			if update {
				return errors.New("Unknown ID field for update")
			}
			return dec.Uint32(&rawAccount.ID)
		case "sex":
			var sexStr string
			err := readString(dec, &sexStr, false)
			if err != nil {
				return err
			}
			sex, err := parser.ParseSex(sexStr)
			if err != nil {
				return err
			}
			rawAccount.Sex = sex
			return nil
		case "status":
			var statusStr string
			err := readString(dec, &statusStr, true)
			if err != nil {
				return err
			}
			status, err := parser.ParseStatus(statusStr)
			if err != nil {
				return err
			}
			rawAccount.Status = status
			return nil
		case "birth":
			return dec.Int64(&rawAccount.Birth)
		case "joined":
			return dec.Uint32(&rawAccount.Joined)
		case "premium":
			premium := &Premium{}
			err := dec.Object(gojay.DecodeObjectFunc(func(dec *gojay.Decoder, key string) error {
				switch key {
				case "start":
					return dec.Uint32(&premium.Start)
				case "finish":
					return dec.Uint32(&premium.Finish)
				}
				return errors.New("Unknown premium field")
			}))
			if err != nil {
				return err
			}
			rawAccount.Premium = premium
			return nil
		case "email":
			err := readString(dec, &rawAccount.Email, false)
			if err != nil {
				return err
			}
			at := strings.Index(rawAccount.Email, "@")
			if at != -1 {
				rawAccount.EmailDomain = uint8(at) + 1
			}
			return nil
		case "phone":
			err := readStringNull(dec, &rawAccount.Phone, false)
			if err != nil {
				return err
			}
			if rawAccount.Phone != nil {
				s := strings.Index(*rawAccount.Phone, "(")
				e := strings.Index(*rawAccount.Phone, ")")
				ui64, err := strconv.ParseUint((*rawAccount.Phone)[s+1:e], 10, 16)
				if err != nil {
					return err
				}
				phoneCode := uint16(ui64)
				rawAccount.PhoneCode = &phoneCode
			}
			return nil
		case "fname":
			return readStringNull(dec, &rawAccount.Fname, true)
		case "sname":
			return readStringNull(dec, &rawAccount.Sname, true)
		case "country":
			return readStringNull(dec, &rawAccount.Country, true)
		case "city":
			return readStringNull(dec, &rawAccount.City, true)
		case "likes":
			return dec.Array(gojay.DecodeArrayFunc(func(dec *gojay.Decoder) error {
				like := RawLike{}
				err := dec.Object(gojay.DecodeObjectFunc(func(dec *gojay.Decoder, key string) error {
					switch key {
					case "id":
						return dec.Uint32(&like.ID)
					case "ts":
						return dec.Uint32(&like.Ts)
					}
					return errors.New(`Unknown like key "` + key + `"`)
				}))
				if err != nil {
					return err
				}
				rawAccount.Likes = append(rawAccount.Likes, like)
				return nil
			}))
		case "interests":
			return dec.Array(gojay.DecodeArrayFunc(func(dec *gojay.Decoder) error {
				var interest string
				err := readString(dec, &interest, true)
				if err != nil {
					return err
				}
				rawAccount.Interests = append(rawAccount.Interests, interest)
				return nil
			}))
		}
		return errors.New(`Unknown account field "` + key + `"`)
	})
}

func (parser *Parser) LikeDecodeFunc(like *Like) gojay.DecodeObjectFunc {
	return gojay.DecodeObjectFunc(func(dec *gojay.Decoder, key string) error {
		switch key {
		case "likee":
			return dec.Uint32(&like.Likee)
		case "ts":
			return dec.Uint32(&like.Ts)
		case "liker":
			return dec.Uint32(&like.Liker)
		}
		return errors.New(`Unknown like field "` + key + `"`)
	})
}

func (parser *Parser) AccountEncodeFunc(account *Account, fields SerializeFields) gojay.EncodeObjectFunc {
	return gojay.EncodeObjectFunc(func(enc *gojay.Encoder) {
		enc.AddUint32Key("id", uint32(account.ID))
		enc.AddStringKey("email", account.Email)

		if fields.Sex() {
			if account.Sex == SexFemale {
				enc.AddStringKey("sex", "f")
			} else {
				enc.AddStringKey("sex", "m")
			}
		}

		if fields.Status() {
			switch account.Status {
			case StatusSingle:
				enc.AddStringKey("status", StatusSingleString)
			case StatusRelationship:
				enc.AddStringKey("status", StatusRelationshipString)
			case StatusComplicated:
				enc.AddStringKey("status", StatusComplicatedString)
			}
		}

		if fields.Fname() {
			if account.Fname == 0 {
				// enc.AddNullKey("fname")
			} else {
				fnameStr, err := parser.dicts.GetFnameString(account.Fname)
				if err != nil {
					// enc.AddNullKey("fname")
				} else {
					enc.AddStringKey("fname", fnameStr)
				}
			}
		}

		if fields.Sname() {
			if account.Sname == 0 {
				// enc.AddNullKey("sname")
			} else {
				snameStr, err := parser.dicts.GetSnameString(account.Sname)
				if err != nil {
					// enc.AddNullKey("sname")
				} else {
					enc.AddStringKey("sname", snameStr)
				}
			}
		}

		if fields.Phone() {
			if account.Phone == nil {
				// enc.AddNullKey("phone")
			} else {
				enc.AddStringKey("phone", *account.Phone)
			}
		}

		if fields.Country() {
			if account.Country == 0 {
				// enc.AddNullKey("country")
			} else {
				countryStr, err := parser.dicts.GetCountryString(account.Country)
				if err != nil {
					// enc.AddNullKey("country")
				} else {
					enc.AddStringKey("country", countryStr)
				}
			}
		}

		if fields.City() {
			if account.City == 0 {
				// enc.AddNullKey("city")
			} else {
				cityStr, err := parser.dicts.GetCityString(account.City)
				if err != nil {
					// enc.AddNullKey("city")
				} else {
					enc.AddStringKey("city", cityStr)
				}
			}
		}

		if fields.Birth() {
			enc.AddInt64Key("birth", account.Birth)
		}

		if fields.Premium() {
			if account.Premium != nil {
				enc.AddObjectKey("premium", gojay.EncodeObjectFunc(func(enc *gojay.Encoder) {
					enc.AddUint32Key("start", account.Premium.Start)
					enc.AddUint32Key("finish", account.Premium.Finish)
				}))
			} else {
				// enc.AddNullKey("premium")
			}
		}
	})
}

func (parser *Parser) EncodeGroupFunc(groupEntry *GroupEntry) gojay.EncodeObjectFunc {
	return gojay.EncodeObjectFunc(func(enc *gojay.Encoder) {
		if groupEntry.GetSex() != 0 {
			if groupEntry.GetSex() == SexFemale {
				enc.AddStringKey("sex", "f")
			} else {
				enc.AddStringKey("sex", "m")
			}
		}

		if groupEntry.GetStatus() != 0 {
			switch groupEntry.GetStatus() {
			case StatusSingle:
				enc.AddStringKey("status", StatusSingleString)
			case StatusRelationship:
				enc.AddStringKey("status", StatusRelationshipString)
			case StatusComplicated:
				enc.AddStringKey("status", StatusComplicatedString)
			}
		}

		if groupEntry.GetInterest() != 0 {
			interestStr, err := parser.dicts.GetInterestString(groupEntry.GetInterest())
			if err != nil {
				// enc.AddNullKey("interests")
			} else {
				enc.AddStringKey("interests", interestStr)
			}
		}

		if groupEntry.GetCountry() != 0 {
			countryStr, err := parser.dicts.GetCountryString(groupEntry.GetCountry())
			if err != nil {
				// enc.AddNullKey("country")
			} else {
				enc.AddStringKey("country", countryStr)
			}
		}

		if groupEntry.GetCity() != 0 {
			cityStr, err := parser.dicts.GetCityString(groupEntry.GetCity())
			if err != nil {
				// enc.AddNullKey("city")
			} else {
				enc.AddStringKey("city", cityStr)
			}
		}

		enc.AddUint32Key("count", groupEntry.Count)
	})
}

func (parser *Parser) ParseStatus(status string) (byte, error) {
	switch status {
	case StatusSingleString:
		return StatusSingle, nil
	case StatusComplicatedString:
		return StatusComplicated, nil
	case StatusRelationshipString:
		return StatusRelationship, nil
	}
	return 0, errors.New("Unknown account status")
}

func (parser *Parser) ParseSex(sex string) (byte, error) {
	if sex[0] != SexFemale && sex[0] != SexMale {
		return 0, errors.New("Invalid account sex")
	}
	return sex[0], nil
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
	if unquote {
		s, err := strconv.Unquote(string(buf))
		*str = &s
		if err != nil {
			return err
		}
	} else {
		s := string(buf)
		unq := s[1 : len(s)-1]
		*str = &unq
	}

	return nil
}
