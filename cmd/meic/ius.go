package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/tool"
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
		logger:       logger,
		serverAddr:   serverAddr,
		selfKeys:     selfKeys,
		issuanceKeys: issuanceKeys,
		// We never want OTSs to have to block when reporting sales, so we
		// allocate a buffer in the sales channel.
		sales:         make(chan TargetPriceSale, 256),
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
//
// This function will only ever return normally if it receives a message on
// the `stop` channel. This can be accomplished without ever sending such
// a message by closing the channel. If you don't want to ever stop it, passing
// a nil channel will do the right thing.
func (ius *IssuanceUpdateSystem) Run(stop <-chan struct{}) {
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

	// start up the OTS instances now that we can possibly respond to their updates
	for idx := range otsImpls {
		go otsImpls[idx].Run(ius.logger.WithField("ots index", idx), ius.sales, ius.updates[idx])
	}

	// TODO: set up a persistent websocket connection to the ndau tx feed
	// https://tendermint.com/docs/app-dev/subscribing-to-events-via-websocket.html
	// and create a channel which receives Issue tx events

	// everything's set up, let's handle some messages!
	for {
		timeout := time.After(10 * time.Minute)
		select {
		case <-stop:
			break
		case <-ius.manualUpdates:
			// TODO: implement the following
			// 1. get the total issuance from the blockchain
			// 2. compute the current desired target sales stack
			// 3. send that stack individually to each OTS

			// TODO: handle incoming Issue txs from the blockchain
			// via this same handler
		case <-timeout:
			// TODO, implement everythin from the manual update, plus:
			timeout = time.After(10 * time.Minute)
		case sale := <-ius.sales:
			// TODO: implement the following:
			// - get the current issuance sequence from the blockchain
			// - generate the issuance tx
			// - send the tx to the signature server
			// - update the tx with the returned signatures
			// - send it to the blockchain
			// - handle errors appropriately
			tx := ndau.NewIssue(sale.Qty, 0)
			tx.Signatures = sigserverman.SignTx(tx, []signature.PublicKey{})
			_, err := tool.SendSync(nil, tx)
			if err != nil {
				panic(err)
			}
		}
	}
}
