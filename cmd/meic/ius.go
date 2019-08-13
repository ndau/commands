package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"

	"github.com/gorilla/websocket"
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
	nodeAddr      *url.URL
	selfKeys      signer.SignDevice
	issuanceKeys  []signature.PublicKey
	sales         chan TargetPriceSale
	updates       []chan UpdateOrders
	manualUpdates chan struct{}
	issueTxs      chan struct{}
	wsConn        *websocket.Conn
}

// NewIUS creates a new IUS and performs required initialization
//
// serverAddress is the external address at which this IUS can be reached.
// nodeAddress is the address (including port) to a ndau node's RPC connection
func NewIUS(
	logger *logrus.Entry,
	serverAddress string,
	selfKeys signer.SignDevice,
	issuanceKeys []signature.PublicKey,
	nodeAddress string,
) (*IssuanceUpdateSystem, error) {
	serverAddr, err := url.Parse(serverAddress)
	if err != nil {
		return nil, errors.Wrap(err, "parsing server address")
	}
	nodeAddr, err := url.Parse(nodeAddress)
	if err != nil {
		return nil, errors.Wrap(err, "parsing node address")
	}
	nodeAddr.Path = "/websocket"

	ius := IssuanceUpdateSystem{
		logger:       logger,
		serverAddr:   serverAddr,
		nodeAddr:     nodeAddr,
		selfKeys:     selfKeys,
		issuanceKeys: issuanceKeys,
		// We never want OTSs to have to block when reporting sales, so we
		// allocate a buffer in the sales channel.
		sales:         make(chan TargetPriceSale, 256),
		updates:       make([]chan UpdateOrders, 0, len(otsImpls)),
		manualUpdates: make(chan struct{}),
		// a buffer size of 1 means that the tx goroutine can just skip
		// adding to this channel if doing so would block.
		// This effectively compresses multiple redundant `Issue`s into one.
		// Doing so doesn't hurt anything, so why not.
		issueTxs: make(chan struct{}, 1),
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

// this function should _not_ be called in a goroutine.
// it handles launching the goroutines necessary to actually monitor the
// websocket connection.
func (ius *IssuanceUpdateSystem) monitorIssueTxs(stop <-chan struct{}) error {
	var err error
	ius.wsConn, _, err = websocket.DefaultDialer.Dial(ius.nodeAddr.String(), nil)
	if err != nil {
		return errors.Wrap(err, "dialing node websocket connection")
	}

	subsMsg := `{"jsonrpc":"2.0","method":"subscribe","id":"0","params":{"query":"tm.event='Tx'"}}`
	pingMsg := `{"jsonrpc":"2.0","method":"abci_info"}`

	// send the subscription message over the connection
	err = ius.wsConn.WriteMessage(websocket.TextMessage, []byte(subsMsg))
	if err != nil {
		return errors.Wrap(err, "sending subscription message")
	}

	// past this point it's all goroutines, so we have to assume that we won't
	// really have network errors ever

	// start a goroutine which keeps the TM websocket connection alive and forwards
	// issues
	go func() {
		// if there's no traffic in 30s, TM closes the websocket.
		// We therefore have to keep pinging it manually
		timeout := time.After(25 * time.Second)

		for {
			select {
			case <-stop:
				break
			case <-timeout:
				err := ius.wsConn.WriteMessage(websocket.TextMessage, []byte(pingMsg))
				if err != nil {
					// can't do much useful with a non-nil error, but we also
					// don't expect to see them very often. If this turns out
					// to bite us often, we can build up some kind of error-
					// handling infrastructure suitable for this multi-goroutine
					// situation.
					panic(errors.Wrap(err, "writing ping to TM"))
				}
				timeout = time.After(25 * time.Second)
			}
		}
	}()

	// start a goroutine which reads all traffic inbound on the TM websocket
	// connection, determines which correspond to issue txs, and forwards
	// notifications for those
	go func() {
		var msg map[string]interface{}
		for {
			// shutdown?
			select {
			case <-stop:
				break
			default:
			}

			err := ius.wsConn.ReadJSON(&msg)
			if err != nil {
				panic(errors.Wrap(err, "reading message from TM"))
			}
			// get the b64-encoded tx data
			b64data, err := getNested(msg, "result", "data", "value", "TxResult", "tx")
			if err != nil {
				if strings.HasPrefix(err.Error(), "item not found:") {
					// in that case, it was probably a pong message
					continue
				}
				panic(errors.Wrap(err, "unexpected TM ws message"))
			}

			data, err := base64.StdEncoding.DecodeString(b64data.(string))
			if err != nil {
				panic(errors.Wrap(err, "decoding tx b64"))
			}

			tx, err := metatx.Unmarshal(data, ndau.TxIDs)
			if err != nil {
				panic(errors.Wrap(err, "unmarshaling tx"))
			}

			if _, ok := tx.(*ndau.Issue); ok {
				// aha, this is what we wanted to see
				// use a non-blocking send, just in case there are a lot of
				// messages incoming
				select {
				case ius.issueTxs <- struct{}{}:
				default:
				}
			}
			// otherwise it doesn't matter, just loop again
		}
	}()

	return nil
}

// Run the issuance service
//
// This function will only ever return normally if it receives a message on
// the `stop` channel. This can be accomplished without ever sending such
// a message by closing the channel. If you don't want to ever stop it, passing
// a nil channel will do the right thing.
func (ius *IssuanceUpdateSystem) Run(stop <-chan struct{}) error {
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
		if ius.wsConn != nil {
			ius.wsConn.Close()
		}
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

	// connect to TM and listen to the TX websocket feed to get notifications of
	// Issue txs
	err := ius.monitorIssueTxs(stop)
	if err != nil {
		return errors.Wrap(err, "creating WS connection to TM to receive Issue tx notifications")
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
		case <-ius.issueTxs:
			// TODO: do exactly what we did above
		case <-timeout:
			// TODO, implement everything from the manual update, plus:
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
