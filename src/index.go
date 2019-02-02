package main

import (
	"fmt"
	"sync"
)

type Index struct {
	store                *Store
	worker               *IndexWorker
	ID                   *IndexReverseID
	Likee                *IndexLikee
	Liker                *IndexLiker
	Interest             *IndexInterest
	City                 *IndexCity
	BirthYear            *IndexYear
	JoinedYear           *IndexYear
	Country              *IndexCountry
	Fname                *IndexFname
	PhoneCode            *IndexPhoneCode
	Group                *IndexGroup
	Sex                  *IndexSex
	Status               *IndexStatus
	InterestPremium      *IndexInterestPremium
	InterestSingle       *IndexInterestSingle
	InterestComplicated  *IndexInterestComplicated
	InterestRelationship *IndexInterestRelationship
}

func NewIndex(store *Store, dicts *Dicts) *Index {
	return &Index{
		store:                store,
		worker:               NewIndexWorker(),
		ID:                   NewIndexReverseID(storePreallocCount),
		Likee:                NewIndexLikee(),
		Liker:                NewIndexLiker(),
		Interest:             NewIndexInterest(),
		City:                 NewIndexCity(),
		BirthYear:            NewIndexYear(),
		JoinedYear:           NewIndexYear(),
		Country:              NewIndexCountry(),
		Fname:                NewIndexFname(),
		PhoneCode:            NewIndexPhoneCode(),
		Group:                NewIndexGroup(dicts),
		Sex:                  NewIndexSex(),
		Status:               NewIndexStatus(),
		InterestPremium:      NewIndexInterestPremium(),
		InterestSingle:       NewIndexInterestSingle(),
		InterestComplicated:  NewIndexInterestComplicated(),
		InterestRelationship: NewIndexInterestRelationship(),
	}
}

/// -------

func (index *Index) WorkerLen() int {
	return index.worker.Len()
}

func (index *Index) Append(account *Account) {
	index.ID.Append(account.ID)
	index.Sex.Append(account.Sex, account.ID)
	index.Status.Append(account.Status, account.ID)
	index.BirthYear.Append(timestampToYear(account.Birth), account.ID)
	index.JoinedYear.Append(timestampToYear(int64(account.Joined)), account.ID)
	index.Fname.Append(account.Fname, account.ID)
	index.Country.Append(account.Country, account.ID)
	index.City.Append(account.City, account.ID)
	if account.Phone != nil {
		index.PhoneCode.Append(account.PhoneCode, account.ID)
	} else {
		index.PhoneCode.Append(0, account.ID)
	}
}

func (index *Index) AppendInterests(account *Account, premium bool, interests ...Interest) {
	for _, interest := range account.Interests {
		index.Interest.Append(interest, account.ID)
		if premium {
			index.InterestPremium.Append(interest, account.Status, account.Sex, account.City, account.Country, account.ID)
		} else {
			if account.Status == StatusSingle {
				index.InterestSingle.Append(interest, account.City, account.Country, account.ID)
			}
			if account.Status == StatusComplicated {
				index.InterestComplicated.Append(interest, account.City, account.Country, account.ID)
			}
			if account.Status == StatusRelationship {
				index.InterestRelationship.Append(interest, account.City, account.Country, account.ID)
			}
		}
	}
	index.Group.AppendHash(CreateHashFromAccount(account), account.Interests...)
}

func (index *Index) AppendLike(liker ID, likee ID, ts uint32) {
	index.Liker.Append(liker, likee, ts)
	index.Likee.Add(likee, liker)
}

// func (index *Index) NewBatch() *IndexBatch {
// 	return &IndexBatch{index: index}
// }

/// -------

type IndexBatch struct {
	index *Index
	jobs  []func()
}

func NewBatchJob(jobs ...func()) func() {
	return func() {
		for _, job := range jobs {
			job()
		}
	}
}

func (batch *IndexBatch) Dispatch() func() {
	defer batch.Release()
	return NewBatchJob(batch.jobs...)
}

func (batch *IndexBatch) Add(account *Account) {
	batch.AddID(account.ID)
	batch.AddSex(account.ID, account.Sex)
	batch.AddStatus(account.ID, account.Status)
	batch.AddBirth(account.ID, timestampToYear(account.Birth))
	batch.AddJoined(account.ID, timestampToYear(int64(account.Joined)))
	if account.Phone != nil {
		batch.AddPhoneCode(account.ID, account.PhoneCode)
	} else {
		batch.AddPhoneCode(account.ID, 0)
	}
	batch.AddCountry(account.ID, account.Country)
	batch.AddCity(account.ID, account.City)
	batch.AddFname(account.ID, account.Fname)
}

func (batch *IndexBatch) AddID(id ID) {
	batch.jobs = append(batch.jobs, func() {
		batch.addID(id)
	})
}

func (batch *IndexBatch) addID(id ID) {
	batch.index.ID.Add(id)
}

func (batch *IndexBatch) AddSex(id ID, sex byte) {
	batch.jobs = append(batch.jobs, func() {
		batch.addSex(id, sex)
	})
}

func (batch *IndexBatch) addSex(id ID, sex byte) {
	batch.index.Sex.Add(sex, id)
}

func (batch *IndexBatch) AddStatus(id ID, status byte) {
	batch.jobs = append(batch.jobs, func() {
		batch.addStatus(id, status)
	})
}

func (batch *IndexBatch) addStatus(id ID, status byte) {
	batch.index.Status.Add(status, id)
}

func (batch *IndexBatch) AddBirth(id ID, birth Year) {
	batch.jobs = append(batch.jobs, func() {
		batch.addBirth(id, birth)
	})
}

func (batch *IndexBatch) addBirth(id ID, birth Year) {
	batch.index.BirthYear.Add(birth, id)
}

func (batch *IndexBatch) AddJoined(id ID, joined Year) {
	batch.jobs = append(batch.jobs, func() {
		batch.addJoined(id, joined)
	})
}

func (batch *IndexBatch) addJoined(id ID, joined Year) {
	batch.index.JoinedYear.Add(joined, id)
}

func (batch *IndexBatch) AddCity(id ID, city City) {
	batch.jobs = append(batch.jobs, func() {
		batch.addCity(id, city)
	})
}

func (batch *IndexBatch) addCity(id ID, city City) {
	batch.index.City.Add(city, id)
}

func (batch *IndexBatch) AddCountry(id ID, country Country) {
	batch.jobs = append(batch.jobs, func() {
		batch.addCountry(id, country)
	})
}

func (batch *IndexBatch) addCountry(id ID, country Country) {
	batch.index.Country.Add(country, id)
}

func (batch *IndexBatch) AddFname(id ID, fname Fname) {
	batch.jobs = append(batch.jobs, func() {
		batch.addFname(id, fname)
	})
}

func (batch *IndexBatch) addFname(id ID, fname Fname) {
	batch.index.Fname.Add(fname, id)
}

// func (batch *IndexBatch) add(account Account) {
// 	batch.index.ID.Add(account.ID)
// 	batch.index.Sex.Add(account.Sex, account.ID)
// 	batch.index.Status.Add(account.Status, account.ID)
// 	batch.index.BirthYear.Add(timestampToYear(account.Birth), account.ID)
// 	batch.index.JoinedYear.Add(timestampToYear(int64(account.Joined)), account.ID)
// 	if account.Phone != nil {
// 		batch.index.PhoneCode.Add(account.PhoneCode, account.ID)
// 	} else {
// 		batch.index.PhoneCode.Add(0, account.ID)
// 	}
// 	batch.index.Country.Add(account.Country, account.ID)
// 	batch.index.City.Add(account.City, account.ID)
// 	batch.index.Fname.Add(account.Fname, account.ID)
// }

func (batch *IndexBatch) AddInterests(id ID, status, sex byte, city City, country Country, premium bool, interests ...Interest) {
	batch.jobs = append(batch.jobs, func() {
		for _, interest := range interests {
			batch.addInterest(id, interest)
			if premium {
				batch.addInterestPremium(id, interest, status, sex, city, country)
			} else {
				if status == StatusSingle {
					batch.addInterestSingle(id, interest, city, country)
				} else if status == StatusComplicated {
					batch.addInterestComplicated(id, interest, city, country)
				} else if status == StatusRelationship {
					batch.addInterestRelationship(id, interest, city, country)
				}
			}
		}
	})
}

func (batch *IndexBatch) RemoveInterests(id ID, status, sex byte, city City, country Country, premium bool, interests ...Interest) {
	batch.jobs = append(batch.jobs, func() {
		for _, interest := range interests {
			batch.removeInterest(id, interest)
			if premium {
				batch.removeInterestPremium(id, interest, status, sex, city, country)
			} else {
				if status == StatusSingle {
					batch.removeInterestSingle(id, interest, city, country)
				} else if status == StatusComplicated {
					batch.removeInterestComplicated(id, interest, city, country)
				} else if status == StatusRelationship {
					batch.removeInterestRelationship(id, interest, city, country)
				}
			}
		}
	})
}

// func (batch *IndexBatch) addInterests(id ID, sex, status byte, city City, country Country, premium bool, hash GroupHash, interests ...Interest) {
// 	for _, interest := range interests {
// 		batch.index.Interest.Add(interest, id)
// 		if premium {
// 			batch.index.InterestPremium.Add(interest, status, sex, city, country, id)
// 		} else {
// 			if status == StatusSingle {
// 				batch.index.InterestSingle.Add(interest, city, country, id)
// 			} else if status == StatusComplicated {
// 				batch.index.InterestComplicated.Add(interest, city, country, id)
// 			} else if status == StatusRelationship {
// 				batch.index.InterestRelationship.Add(interest, city, country, id)
// 			}
// 		}
// 	}
// 	batch.index.Group.AddHash(, interests...)
// }

// func (batch *IndexBatch) addInterests(account Account, premium bool, interests ...Interest) {
// 	for _, interest := range interests {
// 		batch.index.Interest.Add(interest, account.ID)
// 		if premium {
// 			batch.index.InterestPremium.Add(interest, account.Status, account.Sex, account.City, account.Country, account.ID)
// 		} else {
// 			if account.Status == StatusSingle {
// 				batch.index.InterestSingle.Add(interest, account.City, account.Country, account.ID)
// 			} else if account.Status == StatusComplicated {
// 				batch.index.InterestComplicated.Add(interest, account.City, account.Country, account.ID)
// 			} else if account.Status == StatusRelationship {
// 				batch.index.InterestRelationship.Add(interest, account.City, account.Country, account.ID)
// 			}
// 		}
// 	}
// 	batch.index.Group.AddHash(CreateHashFromAccount(&account), interests...)
// }

func (batch *IndexBatch) AddLike(liker ID, likee ID, ts uint32) {
	batch.jobs = append(batch.jobs, func() {
		batch.addLike(liker, likee, ts)
	})
}

func (batch *IndexBatch) addLike(liker ID, likee ID, ts uint32) {
	batch.index.Liker.Add(liker, likee, ts)
	batch.index.Likee.Add(likee, liker)
}

func (batch *IndexBatch) ReplaceSex(id ID, oldSex byte, newSex byte) {
	batch.jobs = append(batch.jobs, func() {
		batch.replaceSex(id, oldSex, newSex)
	})
}

func (batch *IndexBatch) replaceSex(id ID, oldSex byte, newSex byte) {
	batch.index.Sex.Remove(oldSex, id)
	batch.index.Sex.Add(newSex, id)
}

func (batch *IndexBatch) ReplaceStatus(id ID, oldStatus byte, newStatus byte) {
	batch.jobs = append(batch.jobs, func() {
		batch.replaceStatus(id, oldStatus, newStatus)
	})
}

func (batch *IndexBatch) replaceStatus(id ID, oldStatus byte, newStatus byte) {
	batch.index.Status.Remove(oldStatus, id)
	batch.index.Status.Add(newStatus, id)
}

func (batch *IndexBatch) ReplaceBirth(id ID, oldBirth Year, newBirth Year) {
	batch.jobs = append(batch.jobs, func() {
		batch.replaceBirth(id, oldBirth, newBirth)
	})
}

func (batch *IndexBatch) replaceBirth(id ID, oldBirth Year, newBirth Year) {
	batch.index.BirthYear.Remove(oldBirth, id)
	batch.index.BirthYear.Add(newBirth, id)
}

func (batch *IndexBatch) AddPhoneCode(id ID, phoneCode uint16) {
	batch.jobs = append(batch.jobs, func() {
		batch.addPhoneCode(id, phoneCode)
	})
}

func (batch *IndexBatch) addPhoneCode(id ID, phoneCode uint16) {
	batch.index.PhoneCode.Add(phoneCode, id)
}

func (batch *IndexBatch) ReplacePhoneCode(id ID, oldPhoneCode uint16, newPhoneCode uint16) {
	batch.jobs = append(batch.jobs, func() {
		batch.replacePhoneCode(id, oldPhoneCode, newPhoneCode)
	})
}

func (batch *IndexBatch) replacePhoneCode(id ID, oldPhoneCode uint16, newPhoneCode uint16) {
	batch.index.PhoneCode.Remove(oldPhoneCode, id)
	batch.index.PhoneCode.Add(newPhoneCode, id)
}

func (batch *IndexBatch) ReplaceFname(id ID, oldFname Fname, newFname Fname) {
	batch.jobs = append(batch.jobs, func() {
		batch.replaceFname(id, oldFname, newFname)
	})
}

func (batch *IndexBatch) replaceFname(id ID, oldFname Fname, newFname Fname) {
	batch.index.Fname.Remove(oldFname, id)
	batch.index.Fname.Add(newFname, id)
}

func (batch *IndexBatch) ReplaceCountry(id ID, oldCountry Country, newCountry Country) {
	batch.jobs = append(batch.jobs, func() {
		batch.replaceCountry(id, oldCountry, newCountry)
	})
}

func (batch *IndexBatch) replaceCountry(id ID, oldCountry Country, newCountry Country) {
	batch.index.Country.Remove(oldCountry, id)
	batch.index.Country.Add(newCountry, id)
}

func (batch *IndexBatch) ReplaceCity(id ID, oldCity City, newCity City) {
	batch.jobs = append(batch.jobs, func() {
		batch.replaceCity(id, oldCity, newCity)
	})
}

func (batch *IndexBatch) replaceCity(id ID, oldCity City, newCity City) {
	batch.index.City.Remove(oldCity, id)
	batch.index.City.Add(newCity, id)
}

func (batch *IndexBatch) SubGroupHash(hash GroupHash, interests ...Interest) {
	batch.jobs = append(batch.jobs, func() {
		batch.subGroupHash(hash, interests...)
	})
}

func (batch *IndexBatch) subGroupHash(hash GroupHash, interests ...Interest) {
	batch.index.Group.SubHash(hash, interests...)
}

func (batch *IndexBatch) AddGroupHash(hash GroupHash, interests ...Interest) {
	batch.jobs = append(batch.jobs, func() {
		batch.addGroupHash(hash, interests...)
	})
}

func (batch *IndexBatch) addGroupHash(hash GroupHash, interests ...Interest) {
	batch.index.Group.AddHash(hash, interests...)
}

func (batch *IndexBatch) AddInterest(id ID, interest Interest) {
	batch.jobs = append(batch.jobs, func() {
		batch.addInterest(id, interest)
	})
}

func (batch *IndexBatch) addInterest(id ID, interest Interest) {
	batch.index.Interest.Add(interest, id)
}

func (batch *IndexBatch) RemoveInterest(id ID, interest Interest) {
	batch.jobs = append(batch.jobs, func() {
		batch.removeInterest(id, interest)
	})
}

func (batch *IndexBatch) removeInterest(id ID, interest Interest) {
	batch.index.Interest.Remove(interest, id)
}

/// --------

func (batch *IndexBatch) ReplaceInterestPremiumCountry(id ID, interest Interest, oldCountry Country, newCountry Country) {
	batch.jobs = append(batch.jobs, func() {
		batch.replaceInterestPremiumCountry(id, interest, oldCountry, newCountry)
	})
}

func (batch *IndexBatch) replaceInterestPremiumCountry(id ID, interest Interest, oldCountry Country, newCountry Country) {
	batch.index.InterestPremium.RemoveCountry(interest, oldCountry, id)
	batch.index.InterestPremium.AddCountry(interest, newCountry, id)
}

func (batch *IndexBatch) ReplaceInterestPremiumCity(id ID, interest Interest, oldCity City, newCity City) {
	batch.jobs = append(batch.jobs, func() {
		batch.replaceInterestPremiumCity(id, interest, oldCity, newCity)
	})
}

func (batch *IndexBatch) replaceInterestPremiumCity(id ID, interest Interest, oldCity City, newCity City) {
	batch.index.InterestPremium.RemoveCity(interest, oldCity, id)
	batch.index.InterestPremium.AddCity(interest, newCity, id)
}

// -----

func (batch *IndexBatch) ReplaceInterestPremiumStatus(id ID, interest Interest, sex byte, city City, country Country, oldStatus, newStatus byte) {
	batch.jobs = append(batch.jobs, func() {
		batch.replaceInterestPremiumStatus(id, interest, sex, city, country, oldStatus, newStatus)
	})
}

func (batch *IndexBatch) replaceInterestPremiumStatus(id ID, interest Interest, sex byte, city City, country Country, oldStatus, newStatus byte) {
	batch.index.InterestPremium.Remove(interest, oldStatus, sex, city, country, id)
	batch.index.InterestPremium.Add(interest, newStatus, sex, city, country, id)
}

func (batch *IndexBatch) AddInterestPremium(id ID, interest Interest, status, sex byte, city City, country Country) {
	batch.jobs = append(batch.jobs, func() {
		batch.addInterestPremium(id, interest, status, sex, city, country)
	})
}

func (batch *IndexBatch) addInterestPremium(id ID, interest Interest, status, sex byte, city City, country Country) {
	batch.index.InterestPremium.Add(interest, status, sex, city, country, id)
}

func (batch *IndexBatch) RemoveInterestPremium(id ID, interest Interest, status, sex byte, city City, country Country) {
	batch.jobs = append(batch.jobs, func() {
		batch.removeInterestPremium(id, interest, status, sex, city, country)
	})
}

func (batch *IndexBatch) removeInterestPremium(id ID, interest Interest, status, sex byte, city City, country Country) {
	batch.index.InterestPremium.Remove(interest, status, sex, city, country, id)
}

// ******

func (batch *IndexBatch) AddInterestSingle(id ID, interest Interest, city City, country Country) {
	batch.jobs = append(batch.jobs, func() {
		batch.addInterestSingle(id, interest, city, country)
	})
}

func (batch *IndexBatch) addInterestSingle(id ID, interest Interest, city City, country Country) {
	batch.index.InterestSingle.Add(interest, city, country, id)
}

func (batch *IndexBatch) RemoveInterestSingle(id ID, interest Interest, city City, country Country) {
	batch.jobs = append(batch.jobs, func() {
		batch.removeInterestSingle(id, interest, city, country)
	})
}

func (batch *IndexBatch) removeInterestSingle(id ID, interest Interest, city City, country Country) {
	batch.index.InterestSingle.Remove(interest, city, country, id)
}

func (batch *IndexBatch) ReplaceInterestSingleCountry(id ID, interest Interest, oldCountry, newCountry Country) {
	batch.jobs = append(batch.jobs, func() {
		batch.replaceInterestSingleCountry(id, interest, oldCountry, newCountry)
	})
}

func (batch *IndexBatch) replaceInterestSingleCountry(id ID, interest Interest, oldCountry, newCountry Country) {
	batch.index.InterestSingle.RemoveCountry(interest, oldCountry, id)
	batch.index.InterestSingle.AddCountry(interest, newCountry, id)
}

func (batch *IndexBatch) ReplaceInterestSingleCity(id ID, interest Interest, oldCity, newCity City) {
	batch.jobs = append(batch.jobs, func() {
		batch.replaceInterestSingleCity(id, interest, oldCity, newCity)
	})
}

func (batch *IndexBatch) replaceInterestSingleCity(id ID, interest Interest, oldCity, newCity City) {
	batch.index.InterestSingle.RemoveCity(interest, oldCity, id)
	batch.index.InterestSingle.AddCity(interest, newCity, id)
}

/// ******

func (batch *IndexBatch) AddInterestComplicated(id ID, interest Interest, city City, country Country) {
	batch.jobs = append(batch.jobs, func() {
		batch.addInterestComplicated(id, interest, city, country)
	})
}

func (batch *IndexBatch) addInterestComplicated(id ID, interest Interest, city City, country Country) {
	batch.index.InterestComplicated.Add(interest, city, country, id)
}

func (batch *IndexBatch) RemoveInterestComplicated(id ID, interest Interest, city City, country Country) {
	batch.jobs = append(batch.jobs, func() {
		batch.removeInterestComplicated(id, interest, city, country)
	})
}

func (batch *IndexBatch) removeInterestComplicated(id ID, interest Interest, city City, country Country) {
	batch.index.InterestComplicated.Remove(interest, city, country, id)
}

func (batch *IndexBatch) ReplaceInterestComplicatedCountry(id ID, interest Interest, oldCountry, newCountry Country) {
	batch.jobs = append(batch.jobs, func() {
		batch.replaceInterestComplicatedCountry(id, interest, oldCountry, newCountry)
	})
}

func (batch *IndexBatch) replaceInterestComplicatedCountry(id ID, interest Interest, oldCountry, newCountry Country) {
	batch.index.InterestComplicated.RemoveCountry(interest, oldCountry, id)
	batch.index.InterestComplicated.AddCountry(interest, newCountry, id)
}

func (batch *IndexBatch) ReplaceInterestComplicatedCity(id ID, interest Interest, oldCity, newCity City) {
	batch.jobs = append(batch.jobs, func() {
		batch.replaceInterestComplicatedCity(id, interest, oldCity, newCity)
	})
}

func (batch *IndexBatch) replaceInterestComplicatedCity(id ID, interest Interest, oldCity, newCity City) {
	batch.index.InterestComplicated.RemoveCity(interest, oldCity, id)
	batch.index.InterestComplicated.AddCity(interest, newCity, id)
}

/// ******

func (batch *IndexBatch) AddInterestRelationship(id ID, interest Interest, city City, country Country) {
	batch.jobs = append(batch.jobs, func() {
		batch.addInterestRelationship(id, interest, city, country)
	})
}

func (batch *IndexBatch) addInterestRelationship(id ID, interest Interest, city City, country Country) {
	batch.index.InterestRelationship.Add(interest, city, country, id)
}

func (batch *IndexBatch) RemoveInterestRelationship(id ID, interest Interest, city City, country Country) {
	batch.jobs = append(batch.jobs, func() {
		batch.removeInterestRelationship(id, interest, city, country)
	})
}

func (batch *IndexBatch) removeInterestRelationship(id ID, interest Interest, city City, country Country) {
	batch.index.InterestRelationship.Remove(interest, city, country, id)
}

func (batch *IndexBatch) ReplaceInterestRelationshipCountry(id ID, interest Interest, oldCountry, newCountry Country) {
	batch.jobs = append(batch.jobs, func() {
		batch.replaceInterestRelationshipCountry(id, interest, oldCountry, newCountry)
	})
}

func (batch *IndexBatch) replaceInterestRelationshipCountry(id ID, interest Interest, oldCountry, newCountry Country) {
	batch.index.InterestRelationship.RemoveCountry(interest, oldCountry, id)
	batch.index.InterestRelationship.AddCountry(interest, newCountry, id)
}

func (batch *IndexBatch) ReplaceInterestRelationshipCity(id ID, interest Interest, oldCity, newCity City) {
	batch.jobs = append(batch.jobs, func() {
		batch.replaceInterestRelationshipCity(id, interest, oldCity, newCity)
	})
}

func (batch *IndexBatch) replaceInterestRelationshipCity(id ID, interest Interest, oldCity, newCity City) {
	batch.index.InterestRelationship.RemoveCity(interest, oldCity, id)
	batch.index.InterestRelationship.AddCity(interest, newCity, id)
}

type IndexWorker struct {
	jobs chan func()
}

func NewIndexWorker() *IndexWorker {
	return &IndexWorker{
		jobs: make(chan func(), 256*1000),
	}
}

func (worker *IndexWorker) Add(job func()) {
	worker.jobs <- job
}

func (worker *IndexWorker) Run() {
	for job := range worker.jobs {
		job()
		// fmt.Println("process index job")
	}
}

func (worker *IndexWorker) Len() int {
	return len(worker.jobs)
}

// ----

func (index *Index) RunWorker() {
	for i := 1; i <= 2; i++ {
		go index.worker.Run()
	}
}

func (index *Index) Update() {
	index.ID.Update()
	index.Liker.UpdateAll()
	index.Sex.UpdateAll()
	index.Status.UpdateAll()
	index.BirthYear.UpdateAll()
	index.JoinedYear.UpdateAll()
	index.Fname.UpdateAll()
	index.Country.UpdateAll()
	index.City.UpdateAll()
	index.PhoneCode.UpdateAll()
	index.Interest.UpdateAll()
	index.InterestPremium.UpdateAll()
	index.InterestSingle.UpdateAll()
	index.InterestComplicated.UpdateAll()
	index.InterestRelationship.UpdateAll()
	index.Group.UpdateAll()

	fmt.Println("total birth years =", index.BirthYear.Len())
	fmt.Println("total joined years =", index.JoinedYear.Len())
	fmt.Println("total fnames =", index.Fname.Len())
	fmt.Println("total countries =", index.Country.Len())
	fmt.Println("total cities =", index.City.Len())
	fmt.Println("total phone codes =", index.PhoneCode.Len())
	fmt.Println("total interests =", index.Interest.Len())

	// fmt.Println("total group entries =", len(index.Group.entries))
	// for filter, groups := range index.Group.entries {
	// 	for group := range groups {
	// 		for filterVal, entries := range groups[group] {
	// 			fmt.Printf("%b x %b x %b: %d\n", filter, group, filterVal, entries.Len())
	// 		}
	// 	}
	// }
}

// ---

type IndexIterators []IndexIterator

var indexIteratorsPool = sync.Pool{
	New: func() interface{} {
		// fmt.Println("new index iterator")
		ii := make(IndexIterators, 0, 8)
		return &ii
	},
}

func BorrowIndexIterators() *IndexIterators {
	ii := indexIteratorsPool.Get().(*IndexIterators)
	ii.Reset()
	return ii
}

func (iterators *IndexIterators) Release() {
	indexIteratorsPool.Put(iterators)
}

func (iterators *IndexIterators) Reset() {
	fmt.Println("reset index iterators")
	fmt.Println(len(*iterators))
	*iterators = (*iterators)[:0]
	fmt.Println(len(*iterators))
}

// ---

var indexBatchPool = sync.Pool{
	New: func() interface{} {
		// fmt.Println("new index batch")
		return &IndexBatch{
			jobs: make([]func(), 0, 128),
		}
	},
}

func BorrowIndexBatch(index *Index) *IndexBatch {
	batch := indexBatchPool.Get().(*IndexBatch)
	batch.Reset()
	batch.index = index
	return batch
}

func (batch *IndexBatch) Reset() {
	batch.jobs = batch.jobs[:0]
}

func (batch *IndexBatch) Release() {
	indexBatchPool.Put(batch)
}
