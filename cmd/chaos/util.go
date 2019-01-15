package main

import (
	"encoding/json"
	"fmt"
	"os"
	"unicode/utf8"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/chaos/pkg/tool"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
)

func orQuit(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("%v", err))
		cli.Exit(1)
	}
}

func getConfig() *tool.Config {
	config, err := tool.Load()
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

// tmnode sets up a client connection to a Tendermint node
func tmnode(node string) *client.HTTP {
	return client.NewHTTP(node, "/websocket")
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
func finish(*verbose bool, result interface{}, err error, cmdName string) {
	orQuit(errors.Wrap(err, fmt.Sprintf("Main action failed in %s subcommand", cmdName)))
	if *verbose {
		jsresult, err := jsonify(result)
		orQuit(errors.Wrap(err, fmt.Sprintf("jsonify failed in %s subcommand", cmdName)))
		fmt.Println(jsresult)
	}
}
