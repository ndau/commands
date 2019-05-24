package main

import (
	"fmt"
	"sort"
	"strings"
)

func rev(s string) string {
	// this won't work for wide characters, but when we need wide character
	// support, we can rewrite it. until then, it's easier to not attempt
	// to parse the unicode
	i := []byte(s)
	o := make([]byte, len(i))
	for idx, c := range i {
		o[len(o)-idx-1] = c
	}
	return string(o)
}

// Accounts manage tracking and querying a list of accounts
//
// There's a lot of internal complexity here, most of which is to support a
// simple feature: you can refer to an account by any number of nicknames,
// or its address, any unique suffix of any of those.
type Accounts struct {
	rnames []string   // reversed names
	accts  []*Account // pointers so updates to one update all nicknames
}

// NewAccounts creates a new Accounts container
func NewAccounts() *Accounts {
	return &Accounts{
		rnames: make([]string, 0),
		accts:  make([]*Account, 0),
	}
}

// Add an account to the accounts list
//
// Note: if the account address duplicates an existing address, or any nicknames
// duplicate existing nicknames, this will overwrite the existing data. However,
// if the overwritten account retains names not overwritten, it will still be
// accessable via those names.
//
// This behavior means that, in order to add new nicknames, it is safe to just
// call this function again with the new nicknames.
func (as *Accounts) Add(a *Account, nicknames ...string) {
	// construct a sorted list of reversed names by which to refer to this account
	arnames := make([]string, 0, 1+len(nicknames))
	arnames = append(arnames, rev(a.Address.String()))
	for _, n := range nicknames {
		arnames = append(arnames, rev(n))
	}
	sort.Strings(arnames)

	// now we have two lists of reversed names:
	// - as.rnames: the existing data from the accounts list
	// - arnames: the nicknames and account name of this account
	//
	// we also have a list of pointers to Accounts: as.accts
	//
	// the goal: update as.rnames and as.accts with the following properties
	// - as.rnames remains a sorted list of reversed names/nicknames
	// - for every element in as.rnames, using its index as an index into
	//   as.accts gives the correct struct.

	// next up: construct new storage for output rnames and accts lists with
	// sufficient capacity for all elements.
	newrnames := make([]string, 0, len(as.rnames)+len(arnames))
	newaccts := make([]*Account, 0, len(as.accts)+len(arnames))

	// merge-sort them such that we end
	// up with a list containing all elements from both lists.
	idxexisting := 0
	idxnew := 0
	for idxexisting < len(as.rnames) && idxnew < len(arnames) {
		if as.rnames[idxexisting] < arnames[idxnew] {
			newrnames = append(newrnames, as.rnames[idxexisting])
			newaccts = append(newaccts, as.accts[idxexisting])
			idxexisting++
		} else {
			// this prefers the new data if the new name >= the old name
			newrnames = append(newrnames, arnames[idxnew])
			newaccts = append(newaccts, a)
			if as.rnames[idxexisting] == arnames[idxnew] {
				// if they're equal, we also have to increment the existing index
				idxexisting++
			}
			// only increment this after we index using it
			idxnew++
		}
	}
	// we've run out of items, so just append all remaining items
	newrnames = append(newrnames, as.rnames[idxexisting:]...)
	newaccts = append(newaccts, as.accts[idxexisting:]...)
	newrnames = append(newrnames, arnames[idxnew:]...)
	for i := idxnew; i < len(arnames); i++ {
		newaccts = append(newaccts, a)
	}

	as.rnames = newrnames
	as.accts = newaccts
}

// NotUniqueSuffix is returned when someone tries to Get an account with an
// insufficient suffix to distinguish it from other candidates.
type NotUniqueSuffix struct {
	want string
	dym  []string
}

func (nup NotUniqueSuffix) Error() string {
	return fmt.Sprintf(
		"'%s' is not a unique suffix. Candidates: %s",
		nup.want,
		strings.Join(nup.dym, ", "),
	)
}

var _ error = (*NotUniqueSuffix)(nil)

// IsNotUniqueSuffix is true when the provided error is NotUniqueSuffix
func IsNotUniqueSuffix(err error) bool {
	_, ok := err.(NotUniqueSuffix)
	return ok
}

// NoMatch is returned when someone tries to Get an account, but there are no
// matching names.
type NoMatch struct {
	want string
}

func (nm NoMatch) Error() string {
	return "No match found for account " + nm.want
}

var _ error = (*NoMatch)(nil)

// IsNoMatch is true when the provided error is NoMatch
func IsNoMatch(err error) bool {
	_, ok := err.(NoMatch)
	return ok
}

// Get gets the account named
//
// The name can be the account's address, or any of its nicknames. Further, any
// unique suffix is sufficient.
func (as *Accounts) Get(name string) (*Account, error) {
	// special case: if name is blank but there is exactly one account known,
	// just return that. This improves the CLI use case so a user only operating
	// on one account doesn't have to name it all the time.
	if name == "" && len(as.accts) == 1 {
		return as.accts[0], nil
	}

	rname := rev(name)
	// start by using a binary search to locate candidates, if any exist
	start := 0
	end := len(as.rnames) - 1
	for start <= end {
		median := (start + end) / 2
		if as.rnames[median] < rname {
			start = median + 1
		} else {
			end = median - 1
		}
	}

	if start == len(as.rnames) {
		return nil, NoMatch{name}
	}
	matches := make([]string, 0, 1)
	for idx := start; idx < len(as.rnames) && strings.HasPrefix(as.rnames[idx], rname); idx++ {
		matches = append(matches, rev(as.rnames[idx]))
	}

	if len(matches) == 0 {
		return nil, NoMatch{name}
	}
	// if we have supplied a full identifier, it must succeed even if there
	// are other identifiers which have this as a suffix
	if matches[0] == name {
		return as.accts[start], nil
	}
	if len(matches) > 1 {
		return nil, NotUniqueSuffix{name, matches}
	}
	return as.accts[start], nil
}

// AppendNicknames is a shorthand for combining Get and Add to add nicknames to an account
func (as *Accounts) AppendNicknames(name string, nicknames ...string) error {
	acct, err := as.Get(name)
	if err != nil {
		return err
	}
	as.Add(acct, nicknames...)
	return nil
}

// Reverse returns a map of account data to the list of names refering to it
func (as *Accounts) Reverse() map[*Account][]string {
	out := make(map[*Account][]string)
	for idx, rname := range as.rnames {
		name := rev(rname)
		if name != as.accts[idx].Address.String() {
			out[as.accts[idx]] = append(out[as.accts[idx]], name)
		} else if _, ok := out[as.accts[idx]]; !ok {
			// ensure the account appears in the map even if it has no nicknames
			out[as.accts[idx]] = make([]string, 0)
		}
	}
	return out
}
