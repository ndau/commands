package main

import (
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
