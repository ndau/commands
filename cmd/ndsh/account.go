package main

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/key"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// An Account is the ndsh's internal representation of an account
type Account struct {
	Path                  string
	root                  *key.ExtendedKey
	OwnershipPrivate      *signature.PrivateKey
	OwnershipPublic       *signature.PublicKey
	Address               address.Address
	Data                  *backing.AccountData
	PrivateValidationKeys []signature.PrivateKey
}

// NewAccount creates a new Account from a seed and derivation path
//
// It does not attempt to perform any blockchain actions.
//
// If seed is nil, the system's best source of cryptographic randomness will
// be used.
//
// If path is empty, the default /44'/20036'/100/1 is used.
//
// Kind should be one of the kinds defined in the address package in ndaumath.
// If it <= 0, it will default to KindUser.
func NewAccount(seed []byte, path string, kind byte) (Account, error) {
	var err error
	if seed == nil {
		seed = make([]byte, key.RecommendedSeedLen)
		_, err = rand.Read(seed)
	}
	if path == "" {
		path = "/44'/20036'/100/1"
	}
	if kind <= 0 {
		kind = address.KindUser
	}

	root := new(key.ExtendedKey)
	root, err = key.NewMaster(seed)
	if err != nil {
		return Account{}, errors.Wrap(err, "generating root keys")
	}

	return newAccountFromRoot(root, path, kind)
}

func newAccountFromRoot(root *key.ExtendedKey, path string, kind byte) (Account, error) {
	a := Account{
		Path: path,
		root: root,
	}

	ownPvt, err := a.root.DeriveFrom("/", path)
	if err != nil {
		return a, errors.Wrap(err, "deriving ownership key")
	}
	a.OwnershipPrivate, err = ownPvt.SPrivKey()
	if err != nil {
		return a, errors.Wrap(err, "converting ownership pvt key to ndau fmt")
	}

	ownPub, err := ownPvt.Public()
	if err != nil {
		return a, errors.Wrap(err, "deriving public ownership key")
	}
	a.OwnershipPublic, err = ownPub.SPubKey()
	if err != nil {
		return a, errors.Wrap(err, "converting ownership pub key to ndau fmt")
	}

	a.Address, err = address.Generate(kind, a.OwnershipPublic.KeyBytes())
	if err != nil {
		return a, errors.Wrap(err, "generating address")
	}

	return a, nil
}

// AccountDoesNotExist is returned from Update if the account does not exist
type AccountDoesNotExist struct {
	address address.Address
}

var _ error = (*AccountDoesNotExist)(nil)

func (a AccountDoesNotExist) Error() string {
	return fmt.Sprintf("%s does not exist", a.address)
}

// IsAccountDoesNotExist identifies AccountDoesNotExist errors
func IsAccountDoesNotExist(err error) bool {
	_, ok := err.(AccountDoesNotExist)
	return ok
}

// Update this account with current data from the blockchain
//
// Writes debug data if the print function is non-nil and sh.Verbose is true.
// It is safe to pass a nil print function.
func (acct *Account) Update(sh *Shell, print func(format string, args ...interface{})) (err error) {
	ad, resp, err := tool.GetAccount(sh.Node, acct.Address)
	if err != nil {
		if sh.Verbose && print != nil {
			print("    getting account: %s", err.Error())
		}
		return err
	}
	exists := false
	_, err = fmt.Sscanf(resp.Response.Info, query.AccountInfoFmt, &exists)
	if sh.Verbose && print != nil {
		print("    exists: %t", exists)
	}
	if err != nil {
		if sh.Verbose && print != nil {
			print("    err determing whether acct exists: %s", err.Error())
		}
		return err
	}
	if !exists {
		return AccountDoesNotExist{acct.Address}
	}

	acct.Data = ad

	return nil
}

func (acct *Account) display(sh *Shell, nicknames []string) {
	sh.Write("%s (%s): %s", acct.Address, acct.Path, strings.Join(nicknames, " "))
}
