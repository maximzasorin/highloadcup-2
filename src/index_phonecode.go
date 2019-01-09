package main

import "sort"

type IndexPhoneCode struct {
	phoneCodes map[uint16]IDS
}

func NewIndexPhoneCode() *IndexPhoneCode {
	return &IndexPhoneCode{
		phoneCodes: make(map[uint16]IDS),
	}
}

func (indexPhoneCode *IndexPhoneCode) Add(phoneCode uint16, ID uint32) {
	_, ok := indexPhoneCode.phoneCodes[phoneCode]
	if !ok {
		indexPhoneCode.phoneCodes[phoneCode] = make([]uint32, 1)
		indexPhoneCode.phoneCodes[phoneCode][0] = ID
		return
	}

	indexPhoneCode.phoneCodes[phoneCode] = append(indexPhoneCode.phoneCodes[phoneCode], ID)
}

func (indexPhoneCode *IndexPhoneCode) Remove(phoneCode uint16, ID uint32) {
	_, ok := indexPhoneCode.phoneCodes[phoneCode]
	if !ok {
		return
	}
	for i, accountID := range indexPhoneCode.phoneCodes[phoneCode] {
		if accountID == ID {
			indexPhoneCode.phoneCodes[phoneCode] = append(indexPhoneCode.phoneCodes[phoneCode][:i], indexPhoneCode.phoneCodes[phoneCode][i+1:]...)
			return
		}
	}
}

func (indexPhoneCode *IndexPhoneCode) Update(phoneCode uint16) {
	if phoneCode == 0 {
		for phoneCode := range indexPhoneCode.phoneCodes {
			sort.Sort(indexPhoneCode.phoneCodes[phoneCode])
		}
		return
	}

	if _, ok := indexPhoneCode.phoneCodes[phoneCode]; ok {
		sort.Sort(indexPhoneCode.phoneCodes[phoneCode])
	}
}

func (indexPhoneCode *IndexPhoneCode) Get(phoneCode uint16) IDS {
	if _, ok := indexPhoneCode.phoneCodes[phoneCode]; ok {
		return indexPhoneCode.phoneCodes[phoneCode]
	}
	return make(IDS, 0)
}
