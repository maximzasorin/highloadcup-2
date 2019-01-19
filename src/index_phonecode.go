package main

import (
	"sync"
)

type IndexPhoneCode struct {
	rwLock     sync.RWMutex
	phoneCodes map[uint16]*IndexID
}

func NewIndexPhoneCode() *IndexPhoneCode {
	return &IndexPhoneCode{
		phoneCodes: make(map[uint16]*IndexID),
	}
}

func (index *IndexPhoneCode) Add(phoneCode uint16, id ID) {
	index.rwLock.RLock()
	_, ok := index.phoneCodes[phoneCode]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.phoneCodes[phoneCode] = NewIndexID(64)
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.phoneCodes[phoneCode].Add(id)
		index.rwLock.RUnlock()
		return
	}
	index.phoneCodes[phoneCode].Add(id)
	index.rwLock.RUnlock()
}

func (index *IndexPhoneCode) Append(phoneCode uint16, id ID) {
	index.rwLock.RLock()
	_, ok := index.phoneCodes[phoneCode]
	if !ok {
		index.rwLock.RUnlock()
		index.rwLock.Lock()
		index.phoneCodes[phoneCode] = NewIndexID(64)
		index.rwLock.Unlock()
		index.rwLock.RLock()
		index.phoneCodes[phoneCode].Append(id)
		index.rwLock.RUnlock()
		return
	}
	index.phoneCodes[phoneCode].Append(id)
	index.rwLock.RUnlock()
}

func (index *IndexPhoneCode) Update(phoneCode uint16) {
	index.rwLock.Lock()
	_, ok := index.phoneCodes[phoneCode]
	if !ok {
		index.rwLock.Unlock()
		return
	}
	index.phoneCodes[phoneCode].Update()
	index.rwLock.Unlock()
}

func (index *IndexPhoneCode) UpdateAll() {
	index.rwLock.Lock()
	for phoneCode := range index.phoneCodes {
		index.phoneCodes[phoneCode].Update()
	}
	index.rwLock.Unlock()
}

func (index *IndexPhoneCode) Remove(phoneCode uint16, id ID) {
	index.rwLock.RLock()
	_, ok := index.phoneCodes[phoneCode]
	if !ok {
		index.rwLock.RUnlock()
		return
	}
	index.phoneCodes[phoneCode].Remove(id)
	index.rwLock.RUnlock()
}

func (index *IndexPhoneCode) Find(phoneCode uint16) IDS {
	index.rwLock.RLock()
	if _, ok := index.phoneCodes[phoneCode]; ok {
		ids := index.phoneCodes[phoneCode].FindAll()
		index.rwLock.RUnlock()
		return ids
	}
	index.rwLock.RUnlock()
	return make(IDS, 0)
}

func (index *IndexPhoneCode) Len() int {
	index.rwLock.RLock()
	phoneCodesLen := len(index.phoneCodes)
	index.rwLock.RUnlock()
	return phoneCodesLen
}
