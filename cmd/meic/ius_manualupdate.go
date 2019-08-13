package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// manual updates must be POST requests, with a JSON body containing two fields:
// - `timestamp` which is a RFC3339 string and must be within 30s of the server time
// - `signature` which is a signature.Signature which must validate with the server's keys
func (ius *IssuanceUpdateSystem) handleManualUpdate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		reqres.RespondJSON(w, reqres.NewFromErr("POST only", errors.New("method not allowed"), http.StatusMethodNotAllowed))
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewFromErr("could not read http request body", err, http.StatusBadRequest))
		return
	}

	var payload struct {
		Timestamp string              `json:"timestamp"`
		Signature signature.Signature `json:"signature"`
	}

	err = json.Unmarshal(data, &payload)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewFromErr("could not parse request payload", err, http.StatusBadRequest))
		return
	}

	ts, err := time.Parse(time.RFC3339, payload.Timestamp)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewFromErr("could not parse timestamp as RFC3339", err, http.StatusBadRequest))
		return
	}

	tsdiff := ts.Sub(time.Now())
	if tsdiff < 0 {
		tsdiff = -tsdiff
	}
	if tsdiff > 30*time.Second {
		reqres.RespondJSON(w, reqres.NewFromErr("timestamp must be within 30s of server time", errors.New("timestamp out of bounds"), http.StatusBadRequest))
		return
	}

	pubkey, err := ius.selfKeys.Public(0)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewFromErr("could not get server public key", err, http.StatusInternalServerError))
		return
	}

	if !payload.Signature.Verify([]byte(payload.Timestamp), *pubkey) {
		reqres.RespondJSON(w, reqres.NewFromErr("must have signed timestamp string with server private key", errors.New("invalid signature"), http.StatusBadRequest))
		return
	}

	// looks like everything checked out! trigger a manual update
	ius.manualUpdates <- struct{}{}
	w.WriteHeader(http.StatusNoContent)
}
