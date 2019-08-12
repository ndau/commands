package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"

	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/oneiro-ndev/recovery/pkg/signer"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// An IssuanceUpdateSystem has two major responsibilities:
//
// - create issue transactions on demand and send them to the ndau blockchain
// - periodically determine what the set of Target Sales sell orders should be,
//   and instruct each OTS implementation to update its orders accordingly
type IssuanceUpdateSystem struct {
	logger        *logrus.Entry
	serverAddr    *url.URL
	selfKeys      signer.SignDevice
	issuanceKeys  []signature.PublicKey
	sales         chan TargetPriceSale
	updates       []chan UpdateOrders
	manualUpdates chan struct{}
}

// NewIUS creates a new IUS and performs required initialization
func NewIUS(
	logger *logrus.Entry,
	serverAddress string,
	selfKeys signer.SignDevice,
	issuanceKeys []signature.PublicKey,
) (*IssuanceUpdateSystem, error) {
	serverAddr, err := url.Parse(serverAddress)
	if err != nil {
		return nil, errors.Wrap(err, "parsing websocket address")
	}
	ius := IssuanceUpdateSystem{
		logger:        logger,
		serverAddr:    serverAddr,
		selfKeys:      selfKeys,
		issuanceKeys:  issuanceKeys,
		sales:         make(chan TargetPriceSale),
		updates:       make([]chan UpdateOrders, 0, len(otsImpls)),
		manualUpdates: make(chan struct{}),
	}

	for i := 0; i < len(otsImpls); i++ {
		ius.updates = append(ius.updates, make(chan UpdateOrders))
	}

	return &ius, nil
}

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

// Run the issuance service
func (ius *IssuanceUpdateSystem) Run() error {

	// set up http server: this both accepts the connection from the signature
	// server, and serves the update endpoint
	sigserverman := signer.NewServerManager(ius.logger, ius.selfKeys)
	mux := http.NewServeMux()
	httpserver := &http.Server{
		Addr:         ius.serverAddr.Host,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	mux.HandleFunc("/sigserv", sigserverman.Serve())
	mux.HandleFunc("/update", ius.handleManualUpdate)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		httpserver.ListenAndServe()
		wg.Done()
	}()

	// shutdown
	defer func() {
		sigserverman.Close()
		httpserver.Close()
		wg.Wait()
	}()

	ius.logger.Debug("waiting for connection from signature service...")
	<-sigserverman.GetConnectionChan()
	ius.logger.Info("got connection from signature service")

	return nil
}
