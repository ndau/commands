package bitmart

import (
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// CancelOrder cancels an order on bitmart
func CancelOrder(auth *Auth, entrustID uint64) error {
	url := fmt.Sprintf("%sorders/%d", auth.key.Endpoint, entrustID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return errors.Wrap(err, "constructing order request")
	}

	message := fmt.Sprintf("entrust_id=%d", entrustID)

	req.Header.Set(SignatureHeader, HMACSign(auth.key.Secret, message))
	req.Header.Set("Content-Type", "application/json")

	resp, err := auth.Dispatch(req, 3*time.Second)
	if err != nil {
		return errors.Wrap(err, "performing order request")
	}
	defer resp.Body.Close()

	return nil
}
