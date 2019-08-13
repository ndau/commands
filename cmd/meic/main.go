package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	cli "github.com/jawher/mow.cli"
	bitmart "github.com/oneiro-ndev/commands/cmd/meic/ots/bitmart"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/recovery/pkg/signer"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/rpc/client"
)

func main() {
	app := cli.App("bitmart", "track issuance sales and send appropriate transactions")

	var (
		apikeyPath = app.StringArg("API_KEY", "", "path to an apikey.json file")
		symbol     = app.StringOpt("S symbol", bitmart.NdauSymbol, "trade symbol to watch")
		ndauRPC    = app.StringArg("NDAU_RPC", "", "RPC address to the ndau blockchain")
		intervalI  = app.IntOpt("i interval", 60, "interval between polls to the bitmart API (seconds)")
		wsAddrS    = app.StringOpt("s serve", "ws://localhost:28260", "address to which the signature service should connect")
		pubkeyS    = app.StringArg("PUB_KEY", "", "issuance service's ndau-format public key")
		privkeyS   = app.StringOpt("p priv-key", "", "issuance service's ndau-format private key. if not set, will prompt")
		signkeysS  = app.StringsArg("SIGN_KEYS", []string{}, "ndau public keys to send to the signature server to sign the tx with")
		lastTradeI = app.IntOpt("l last-trade", 0, "ensure_id of last issued trade")
	)

	app.Spec = "API_KEY NDAU_RPC PUB_KEY [-i][-s][-S][-p][-l] SIGN_KEYS..."

	app.Action = func() {
		// set up and validate arguments
		if *privkeyS == "" {
			pks := input("ndau-format private key: ")
			privkeyS = &pks
		}
		logger := logrus.NewEntry(logrus.New())
		selfkey, err := signer.NewVirtualDevice(*pubkeyS, *privkeyS)
		check(err, "creating signer key from supplied keys")

		signkeys := make([]signature.PublicKey, 0, len(*signkeysS))
		for _, skS := range *signkeysS {
			sk, err := signature.ParsePublicKey(skS)
			check(err, "parsing signing key "+skS)
			signkeys = append(signkeys, *sk)
		}

		wsAddr, err := url.Parse(*wsAddrS)
		check(err, "parsing address on which to serve websocket connections")

		interval := time.Duration(*intervalI) * time.Second

		rpcNode := client.NewHTTP(*ndauRPC, "/websocket")
		var issuer address.Address
		err = tool.Sysvar(rpcNode, sv.ReleaseFromEndowmentAddressName, &issuer)
		check(err, "getting issuer address")

		key, err := bitmart.LoadAPIKey(*apikeyPath)
		check(err, "loading api key")
		auth := bitmart.NewAuth(key)

		lastTrade := int64(*lastTradeI)

		// ok, config is set
		// set up a signer server, and wait for a connection from the signature service
		log.Print("configuration validated; setting up websocket server...")
		sigserverman := signer.NewServerManager(logger, selfkey)
		mux := http.NewServeMux()
		httpserver := &http.Server{
			Addr:         wsAddr.Host,
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		}
		mux.HandleFunc(wsAddr.Path, sigserverman.Serve())

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

		log.Print("waiting for connection from signature service...")
		<-sigserverman.GetConnectionChan()
		log.Print("got connection from signature service!")
		// poll loop
		for {
			log.Print("polling bitmart API for new trades...")
			trades, err := bitmart.GetTradeHistoryAfter(&auth, *symbol, lastTrade)
			check(err, "getting trades")
			trades, err = bitmart.FilterSales(&auth, trades)
			check(err, "filtering sales")

			totalNewSales := math.Ndau(0)
			for _, trade := range trades {
				totalNewSales += trade.Amount
			}
			log.Printf("total new sales: %s ndau", totalNewSales)

			if totalNewSales > 0 {
				log.Print("creating tx and sending to signature server...")
				ad, _, err := tool.GetAccount(rpcNode, issuer)
				check(err, "getting issuer account")
				issue := ndau.NewIssue(totalNewSales, ad.Sequence+1)

				// get it signed
				sigs := sigserverman.SignTx(issue, signkeys)
				log.Print(*issue)
				// unpack and validate response
				issue.ExtendSignatures(sigs)
				if len(sigs) < len(signkeys) {
					bail("signature service failed to sign the tx enough times: expect %d, have %d", len(signkeys), len(sigs))
				}

				// now send it
				log.Print("sending signed Issue tx to the ndau blockchain...")
				_, err = tool.SendCommit(rpcNode, issue)
				check(err, "submitting issue tx to blockchain")

				// update the last trade so we don't double-issue
				for _, trade := range trades {
					if trade.TradeID > lastTrade {
						lastTrade = trade.TradeID
					}
				}
				log.Printf("successfully submitted tx. new last trade_id: %d", lastTrade)
			}
			log.Printf("now waiting %s", interval)
			time.Sleep(interval)
		}
	}
	app.Run(os.Args)
}
