package claimer

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	sdk "github.com/ndau/ndau/pkg/api_sdk"
	"github.com/ndau/ndau/pkg/ndau"
	"github.com/ndau/ndau/pkg/ndauapi/reqres"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/signature"
	log "github.com/sirupsen/logrus"
)

// intended to be launched in a goroutine. Create a claim node reward tx
// and dispatch it to the configured node.
func dispatch(logger *log.Entry, node string, addr address.Address, keys []signature.PrivateKey) {
	logger = logger.WithFields(log.Fields{
		"nodeURL":       node,
		"winnerAddress": addr.String(),
	})

	client, err := sdk.NewClient(node)
	if err != nil {
		logger.WithError(err).Error("could not connect to ndau node")
		return
	}

	ad, err := client.GetAccount(addr)
	if err != nil {
		logger.WithError(err).Error("could not get account data")
		return
	}

	tx := ndau.NewClaimNodeReward(addr, ad.Sequence+1, keys...)
	_, err = client.Send(tx)
	if err != nil {
		logger.WithError(err).Error("could not send claim tx")
	} else {
		logger.Info("successfully claimed node reward")
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
			if config.SyncMode != nil && *config.SyncMode {
				dispatch(logger, config.NodeAPI, payload.Winner, keys)
			} else {
				go dispatch(logger, config.NodeAPI, payload.Winner, keys)
			}
		} else {
			logger.WithField("winnerAddress", payload.Winner).Info("winner was not among configured nodes")
		}

		reqres.RespondJSON(w, reqres.OKResponse(struct {
			Dispatched bool `json:"dispatched"`
		}{exists}))
	}
}
