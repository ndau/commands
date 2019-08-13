package main

import (
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
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

// Run the issuance service
//
// This function will only ever return normally without error if it receives
// a message on the `stop` channel. This can be accomplished without ever
// sending such a message by closing the channel. If you don't want to ever
// stop it, passing a nil channel will do the right thing.
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

	// everything's set up, let's handle some messages!
	for {
		timeout := time.After(10 * time.Minute)
		select {
		case <-stop:
			break
		case <-ius.manualUpdates:
			ius.updateOTSs()
		case <-ius.issueTxs:
			ius.updateOTSs()
		case <-timeout:
			ius.updateOTSs()
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
