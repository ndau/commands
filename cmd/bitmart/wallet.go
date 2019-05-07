package bitmart

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// Wallet represents a particular wallet
type Wallet struct {
	ID        string  `json:"id"`
	Available float64 `json:"available"`
	Name      string  `json:"name"`
	Frozen    float64 `json:"frozen"`
}

// UnmarshalJSON implements json.Unmarshaler
//
// It's necessary because we can't trust bitmart to actually send numbers
// in numeric fields
func (w *Wallet) UnmarshalJSON(data []byte) error {
	var obj map[string]interface{}
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return errors.Wrap(err, "wallet")
	}

	getnum := func(name string) float64 {
		field, ok := obj[name]
		if !ok {
			err = fmt.Errorf("wallet field %s not found", name)
			return 0
		}
		switch f := field.(type) {
		case float64:
			return f
		case string:
			var fl float64
			fl, err = strconv.ParseFloat(f, 64)
			return fl
		}
		err = fmt.Errorf("unexpected type for Wallet.%s: %T", name, field)
		return 0
	}

	getstr := func(name string) string {
		field, ok := obj[name]
		if !ok {
			err = fmt.Errorf("wallet field %s not found", name)
			return ""
		}
		switch f := field.(type) {
		case string:
			return f
		}
		err = fmt.Errorf("unexpected type for Wallet.%s: %T", name, field)
		return ""
	}

	w.ID = getstr("id")
	w.Available = getnum("available")
	w.Name = getstr("name")
	w.Frozen = getnum("frozen")
	return err
}

// GetWallets retrieves the list of user wallets
func GetWallets(auth *Auth) ([]Wallet, error) {
	req, err := http.NewRequest(http.MethodGet, APIWallet, nil)
	if err != nil {
		return nil, errors.Wrap(err, "constructing wallet request")
	}
	resp, err := auth.Dispatch(req, 3*time.Second)
	if err != nil {
		return nil, errors.Wrap(err, "performing wallet request")
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading wallet response")
	}

	var wallets []Wallet
	err = json.Unmarshal(data, &wallets)
	if err != nil {
		return wallets, errors.Wrap(err, "parsing wallet response")
	}

	return wallets, nil
}
