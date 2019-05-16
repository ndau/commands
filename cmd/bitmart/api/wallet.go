package bitmart

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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

	// attempts to get all fields. err remains nil if everything succeeded
	w.ID, err = getStr(obj, "id")
	w.Available, err = getFloat(obj, "available")
	w.Name, err = getStr(obj, "name")
	w.Frozen, err = getFloat(obj, "frozen")
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
		fmt.Fprintln(os.Stderr, string(data))
		return wallets, errors.Wrap(err, "parsing wallet response")
	}

	return wallets, nil
}
