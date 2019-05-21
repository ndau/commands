package main

import (
	"io"
	"sort"

	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// An Account is the ndsh's internal representation of an account
type Account struct {
	Address               address.Address
	Data                  *backing.AccountData
	OwnershipPublic       *signature.PublicKey
	OwnershipPrivate      *signature.PrivateKey
	PrivateValidationKeys []signature.PrivateKey
}

// NewAccount creates a new Account from an algorithm and seed
//
// It does not attempt to perform any blockchain actions.
//
// Kind should be one of the kinds defined in the address package in ndaumath.
// If it <= 0, it will default to KindUser.
func NewAccount(with signature.Algorithm, seed io.Reader, kind byte) (Account, error) {
	if kind <= 0 {
		kind = address.KindUser
	}
	publicOwn, privateOwn, err := signature.Generate(with, seed)
	if err != nil {
		return Account{}, errors.Wrap(err, "generating keys")
	}
	addr, err := address.Generate(kind, publicOwn.KeyBytes())
	if err != nil {
		return Account{}, errors.Wrap(err, "generating address")
	}
	return Account{
		Address:          addr,
		OwnershipPublic:  &publicOwn,
		OwnershipPrivate: &privateOwn,
	}, nil
}

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
func (as *Accounts) Add(a Account, nicknames ...string) {
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

	if len(as.rnames) == 0 {
		// we can skip a sort step here because there is no existing data, so
		// we can just pick the new stuff
		as.rnames = arnames
		as.accts = make([]*Account, len(arnames))
		for idx := range as.accts {
			as.accts[idx] = &a
		}
		return
	}

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
			newaccts = append(newaccts, &a)
			idxnew++
			if as.rnames[idxexisting] == as.rnames[idxnew] {
				// if they're equal, we also have to increment the existing index
				idxexisting++
			}
		}
	}
	// we've run out of items, so just append all remaining items
	newrnames = append(newrnames, as.rnames[idxexisting:]...)
	newaccts = append(newaccts, as.accts[idxexisting:]...)
	newrnames = append(newrnames, arnames[idxnew:]...)
	for i := idxnew; i < len(arnames); i++ {
		newaccts = append(newaccts, &a)
	}

	as.rnames = newrnames
	as.accts = newaccts
}
