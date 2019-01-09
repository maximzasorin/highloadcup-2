package main

import "sort"

type IndexFname struct {
	fnames map[Fname]IDS
}

func NewIndexFname() *IndexFname {
	return &IndexFname{
		fnames: make(map[Fname]IDS),
	}
}

func (indexFname *IndexFname) Add(fname Fname, ID uint32) {
	_, ok := indexFname.fnames[fname]
	if !ok {
		indexFname.fnames[fname] = make(IDS, 1)
		indexFname.fnames[fname][0] = ID
		return
	}

	indexFname.fnames[fname] = append(indexFname.fnames[fname], ID)
}

func (indexFname *IndexFname) Remove(fname Fname, ID uint32) {
	_, ok := indexFname.fnames[fname]
	if !ok {
		return
	}
	for i, accountID := range indexFname.fnames[fname] {
		if accountID == ID {
			indexFname.fnames[fname] = append(indexFname.fnames[fname][:i], indexFname.fnames[fname][i+1:]...)
			return
		}
	}
}

func (indexFname *IndexFname) Update(fname Fname) {
	if fname == 0 {
		for fname := range indexFname.fnames {
			sort.Sort(indexFname.fnames[fname])
		}
		return
	}

	if _, ok := indexFname.fnames[fname]; ok {
		sort.Sort(indexFname.fnames[fname])
	}
}

func (indexFname *IndexFname) Get(fname Fname) IDS {
	if _, ok := indexFname.fnames[fname]; ok {
		return indexFname.fnames[fname]
	}
	return make(IDS, 0)
}
