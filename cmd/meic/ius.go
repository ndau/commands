package main

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/oneiro-ndev/commands/cmd/meic/ots"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/oneiro-ndev/recovery/pkg/signer"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	tmclient "github.com/tendermint/tendermint/rpc/client"
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
	sales         chan ots.TargetPriceSale
	updates       []chan ots.UpdateOrders
	manualUpdates chan struct{}
	tmNode        *tmclient.HTTP
	stackGen      uint
	stackDefault  uint
	config        *Config
	errs          chan error
}

// NewIUS creates a new IUS and performs required initialization
//
// serverAddress is the external address at which this IUS can be reached.
// nodeAddress is the address (including port) to a ndau node's RPC connection
func NewIUS(
	logger *logrus.Entry,
	serverAddress string,
	selfKeys signer.SignDevice,
	nodeAddress string,
	stackDefault uint,
	config *Config,
) (*IssuanceUpdateSystem, error) {
	serverAddr, err := url.Parse(serverAddress)
	if err != nil {
		return nil, errors.Wrap(err, "parsing server address")
	}
	nodeAddr, err := url.Parse(nodeAddress)
	if err != nil {
		return nil, errors.Wrap(err, "parsing node address")
	}
	nodeAddr.Path = ""
	tmNode := tool.Client(nodeAddr.String())
	nodeAddr.Path = "/websocket"

	ius := IssuanceUpdateSystem{
		logger:     logger,
		serverAddr: serverAddr,
		nodeAddr:   nodeAddr,
		selfKeys:   selfKeys,
		// We never want OTSs to have to block when reporting sales, so we
		// allocate a buffer in the sales channel.
		sales:         make(chan ots.TargetPriceSale, 256),
		updates:       make([]chan ots.UpdateOrders, 0, len(otsImpls)),
		manualUpdates: make(chan struct{}),
		tmNode:        tmNode,
		stackGen:      stackDefault,
		stackDefault:  stackDefault,
		config:        config,
		errs:          make(chan error),
	}

	if ius.config != nil {
		for _, o := range ius.config.DepthOverrides {
			if o > ius.stackGen {
				ius.stackGen = o
			}
		}
	}

	fmt.Println("Impls = ", otsImpls)
	for i := 0; i < len(otsImpls); i++ {
		err := otsImpls[i].Init(ius.logger.WithField("method", "init"))
		check(err, fmt.Sprintf("initializing ots %d", i))
		ius.updates = append(ius.updates, make(chan ots.UpdateOrders))
	}

	fmt.Println("Impls = ", otsImpls)
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
	sigserv := signer.NewServerManager(ius.logger, ius.selfKeys)
	mux := http.NewServeMux()
	httpserver := &http.Server{
		Addr:         ius.serverAddr.Host,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	mux.HandleFunc("/signer", sigserv.Serve())
	mux.HandleFunc("/update", ius.handleManualUpdate)
	fmt.Println("starting IUS run")
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		httpserver.ListenAndServe()
		wg.Done()
	}()

	// shutdown
	defer func() {
		sigserv.Close()
		httpserver.Close()
		wg.Wait()
	}()

	ius.logger.Debug("waiting for connection from signature service...")
	fmt.Println("before GetConnection")
	<-sigserv.GetConnectionChan()
	fmt.Println("after GetCOnnection")
	ius.logger.Info("got connection from signature service")

	// start up the OTS instances now that we can possibly respond to their updates
	for idx := range otsImpls {
		go otsImpls[idx].Run(
			ius.logger.WithField("ots index", idx),
			ius.sales,
			ius.updates[idx],
			ius.errs,
		)
	}
	// sync up the OTSs with the current state of issuance on the blockchain
	ius.updateOTSs()
	// everything's set up, let's handle some messages!
	for {
		timeout := time.After(10 * time.Minute)
		select {
		case <-stop:
			break
		case <-ius.manualUpdates:
			ius.updateOTSs()
		case <-timeout:
			ius.updateOTSs()
		case sale := <-ius.sales:
			ius.handleSale(sale, sigserv)
			fmt.Println("sale = ", sale)
			// wait for Issue TX to settle on blockchain
			time.Sleep(1 * time.Second)
			ius.updateOTSs()
		case err := <-ius.errs:
			// for now, just quit and depend on procmon to restart us after a
			// delay if any OTS fails. we may wish to revisit this policy later.
			check(err, "goroutine error")
		}
	}
}
