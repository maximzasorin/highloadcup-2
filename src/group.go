package main

import (
	"errors"
	"net/url"
)

type Group struct {
	QueryID string
	Limit   string
	Order   string

	// Filter
	Sex       string
	Email     string
	Status    string
	Fname     string
	Sname     string
	Phone     string
	Country   string
	City      string
	Birth     uint8
	Interests string
	Likes     uint32
	Premium   bool

	// Group
	Keys string
}

func NewGroup() *Group {
	return &Group{QueryID: "<unknown>"}
}

func (group *Group) ParseGroup(query string) error {
	values, err := url.ParseQuery(query)
	if err != nil {
		return err
	}

	for param, paramValues := range values {
		if len(paramValues) != 1 || paramValues[0] == "" {
			return errors.New("Invalid group param value")
		}

		err := group.ParseParam(param, paramValues[0])
		if err != nil {
			return err
		}
	}

	return nil
}

func (group *Group) ParseParam(param string, value string) error {
	switch param {
	case "query_id":
		group.QueryID = value
	default:
		return errors.New("Unknown group param")
	}

	return nil
}
