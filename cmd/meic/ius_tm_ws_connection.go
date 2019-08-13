package main

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/pkg/errors"
)

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
					check(err, "writing ping to TM")
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
				check(err, "reading message from TM")
			}
			// get the b64-encoded tx data
			b64data, err := getNested(msg, "result", "data", "value", "TxResult", "tx")
			if err != nil {
				if strings.HasPrefix(err.Error(), "item not found:") {
					// in that case, it was probably a pong message
					continue
				}
				check(err, "unexpected TM ws message")
			}

			data, err := base64.StdEncoding.DecodeString(b64data.(string))
			if err != nil {
				check(err, "decoding tx b64")
			}

			tx, err := metatx.Unmarshal(data, ndau.TxIDs)
			if err != nil {
				check(err, "unmarshaling tx")
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
