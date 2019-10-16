package main

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
	"fmt"
	"os"
	"unicode/utf8"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	config "github.com/oneiro-ndev/ndau/pkg/tool.config"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
)

func orQuit(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("%v", err))
		cli.Exit(1)
	}
}

func getConfig() *config.Config {
	config, err := config.LoadDefault(config.GetConfigPath())
	orQuit(errors.Wrap(err, "Failed to load configuration"))
	return config
}

// validateBytes ensures that the submitted bytes are valid utf-8,
// by quitting if they're not
func validateBytes(bytes []byte) {
	if !utf8.Valid(bytes) {
		orQuit(fmt.Errorf("'%q' is not a valid utf-8 sequence", bytes))
	}
}

// JSG create global client conn, will get reused for server mode
var nodeHTTP *client.HTTP

// tmnode sets up a client connection to a Tendermint node
func tmnode(node string, json, compact *bool) client.ABCIClient {
	if json != nil && *json {
		return tool.NewJSONClient(!*compact)
	}

	if nodeHTTP == nil {
		nodeHTTP = client.NewHTTP(node, "/websocket")
	}
	return nodeHTTP
}

// turn a jsonable blob into pretty-printed json
func jsonify(jsonable interface{}) (string, error) {
	js, err := json.MarshalIndent(jsonable, "", "  ")
	if err != nil {
		return "", err
	}

	return string(js), nil
}

// finish a command by pretty-printing its result as json
func finish(verbose bool, result interface{}, err error, cmdName string) {
	orQuit(errors.Wrap(err, fmt.Sprintf("Main action failed in %s subcommand", cmdName)))
	if verbose {
		jsresult, err := jsonify(result)
		orQuit(errors.Wrap(err, fmt.Sprintf("jsonify failed in %s subcommand", cmdName)))
		fmt.Println(jsresult)
	}
}

// query the account to get the current sequence
func sequence(conf *config.Config, addr address.Address) uint64 {
	ad, _, err := tool.GetAccount(tmnode(conf.Node, nil, nil), addr)
	orQuit(errors.Wrap(
		err,
		fmt.Sprintf("Failed to get current sequence number for %s", addr),
	))
	return ad.Sequence + 1
}
