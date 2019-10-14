package main

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"encoding/json"
	"fmt"

	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// ErrNilStage is returned when Stage or Tx is nil
var ErrNilStage = errors.New("Staged tx is nil")

// Stage keeps track of a staged transaction
type Stage struct {
	Account *Account
	Tx      metatx.Transactable
}

// Sign the staged tx
//
// If sigs has length 0 and the account is set, appropriate signatures
// are drawn from the account.
func (s *Stage) Sign(sigs []signature.Signature) error {
	if s == nil || s.Tx == nil {
		return ErrNilStage
	}

	if s.Account != nil && len(sigs) == 0 {
		sb := s.Tx.SignableBytes()
		switch s.Tx.(type) {
		case ndau.Signable:
			for _, pvt := range s.Account.PrivateValidationKeys {
				sigs = append(sigs, pvt.Sign(sb))
			}
		case *ndau.SetValidation:
			if s.Account.OwnershipPrivate == nil {
				return errors.New("unknown ownership private key; cannot sign SetValidation")
			}
			sigs = append(sigs, s.Account.OwnershipPrivate.Sign(sb))
		}
	}

	switch v := s.Tx.(type) {
	case ndau.Signable:
		v.ExtendSignatures(sigs)
		s.Tx = v.(metatx.Transactable)
		return nil
	case *ndau.SetValidation:
		if len(sigs) > 0 {
			v.Signature = sigs[0]
		}
		s.Tx = v
		return nil
	default:
		return fmt.Errorf("cannot sign %T", v)
	}
}

// Override the specified field name with the given value
func (s *Stage) Override(k string, v interface{}) error {
	if s == nil || s.Tx == nil {
		return ErrNilStage
	}

	data, err := json.Marshal(s.Tx)
	if err != nil {
		return errors.Wrap(err, "marshaling staged tx for override")
	}
	var jsdata map[string]interface{}
	err = json.Unmarshal(data, &jsdata)
	if err != nil {
		return errors.Wrap(err, "unmarshaling tx into map for override")
	}

	// update sequence
	jsdata[k] = v

	// clear existing signatures: they'll be invalid now
	delete(jsdata, "signature")
	delete(jsdata, "signatures")

	// clear out the tx object
	txid, err := metatx.TxIDOf(s.Tx, ndau.TxIDs)
	if err != nil {
		return errors.Wrap(err, "getting txid")
	}
	s.Tx = metatx.Clone(ndau.TxIDs[txid])

	data, err = json.Marshal(jsdata)
	if err != nil {
		return errors.Wrap(err, "marshaling edited tx")
	}
	err = json.Unmarshal(data, &s.Tx)
	if err != nil {
		return errors.Wrap(err, "unmarshaling edited tx")
	}

	// re-sign with associated account
	if s.Account != nil {
		err = s.Sign(nil)
		if err != nil {
			return err
		}
	}

	return nil
}
