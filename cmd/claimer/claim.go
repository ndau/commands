package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	log "github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/rpc/client"
)

// intended to be launched in a goroutine. Create a claim node reward tx
// and dispatch it to the configured node.
func dispatch(logger *log.Entry, node string, addr address.Address, keys []signature.PrivateKey) {
	logger = logger.WithFields(log.Fields{
		"node url":       node,
		"winner address": addr.String(),
	})

	rpc := client.NewHTTP(node, "/websocket")

	ad, _, err := tool.GetAccount(rpc, addr)
	if err != nil {
		logger.WithError(err).Error("could not get account data")
		return
	}

	tx := ndau.NewClaimNodeReward(addr, ad.Sequence+1, keys...)
	_, err = tool.SendCommit(rpc, tx)
	if err != nil {
		logger.WithError(err).Error("could not send claim tx")
	}
}

// Claim returns a HandlerFunc which claims a node reward
func Claim(config *Config, logger *log.Entry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("could not read request body", err, http.StatusBadRequest))
			return
		}
		payload := struct {
			Random int64           `json:"random"`
			Winner address.Address `json:"winner"`
		}{}
		err = json.Unmarshal(body, &payload)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("could not parse payload JSON", err, http.StatusBadRequest))
			return
		}

		keys, exists, err := config.GetKeys(payload.Winner)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("misconfiguration", err, http.StatusInternalServerError))
			return
		}

		if exists {
			go dispatch(logger, config.NodeRPC, payload.Winner, keys)
		}

		reqres.RespondJSON(w, reqres.OKResponse(struct {
			Dispatched bool `json:"dispatched"`
		}{exists}))
	}
}
