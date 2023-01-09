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
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
)

const servicesURL = "https://s3.us-east-2.amazonaws.com/ndau-json/services.json"

var (
	servicesJSON     map[string]interface{}
	haveServicesJSON chan error

	// ClientURL stores client URL currently in use
	ClientURL *url.URL

	// RecoveryURL stores the recovery URL currently in use
	RecoveryURL *url.URL
)

// on init, async get the services list
// just try to read from haveServicesJson before trying to use it
func init() {
	haveServicesJSON = make(chan error)
	go func() {
		defer close(haveServicesJSON)
		resp, err := http.Get(servicesURL)
		if err != nil {
			haveServicesJSON <- errors.Wrap(err, "getting services map")
			return
		}

		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			haveServicesJSON <- errors.Wrap(err, "reading services data")
			return
		}

		err = json.Unmarshal(data, &servicesJSON)
		if err != nil {
			haveServicesJSON <- errors.Wrap(err, "unmarshaling services data")
			fmt.Fprintln(os.Stderr, "unmarshaling services data:", err)
			return
		}
	}()
}

func getNested(dict map[string]interface{}, path ...string) (interface{}, error) {
	return getNestedInner(nil, dict, path)
}

func getNestedInner(breadcrumbs []string, dict map[string]interface{}, path []string) (interface{}, error) {
	if len(path) == 0 {
		return dict, nil
	}

	head := path[0]
	rest := path[1:]

	makeerr := func(msg string, context ...interface{}) error {
		prefix := " @ map"
		for _, b := range breadcrumbs {
			prefix += fmt.Sprintf("[%s]", b)
		}
		prefix += fmt.Sprintf("[%s]", head)
		return fmt.Errorf(msg+prefix, context...)
	}

	inner, ok := dict[head]
	if !ok {
		return nil, makeerr("item not found: %s", head)
	}

	if len(rest) == 0 {
		return inner, nil
	}

	imap, ok := inner.(map[string]interface{})
	if !ok {
		return nil, makeerr("unexpected item type: %T", inner)
	}

	return getNestedInner(append(breadcrumbs, head), imap, rest)
}

func getService(name, nodename string, path ...string) (*url.URL, error) {
	svci, err := getNested(servicesJSON, path...)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("retrieving %s for %s", name, nodename))
	}

	svc, ok := svci.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected %s type %T in %s", name, svci, nodename)
	}

	if !strings.HasPrefix(svc, "http") {
		svc = "https://" + svc
	}
	return url.Parse(svc)
}

func getClient(network string, node int) (client.ABCIClient, error) {
	if node < 0 {
		return nil, fmt.Errorf("invalid node: %d", node)
	}

	ClientURL = nil
	RecoveryURL = nil

	var netname string
	var err error

	switch strings.ToLower(network) {
	case "main", "mainnet":
		netname = "mainnet"
	case "test", "testnet":
		netname = "testnet"
	case "dev", "devnet":
		netname = "devnet"
	case "local", "localnet":
		ClientURL, err = url.Parse(fmt.Sprintf("http://localhost:%d", 26670+node))
		if err != nil {
			return nil, errors.New("bad code in net.go: couldn't parse localnet url")
		}
	default:
		ClientURL, err = url.Parse(network)
		if err != nil {
			// suppress the actual error, but use our own
			return nil, fmt.Errorf("invalid URL: %s", network)
		}
	}

	if ClientURL == nil {
		// user specified a symbolic name, which means we have to wait for
		// the AWS query to resolve
		select {
		case <-time.After(10 * time.Second):
			err = errors.New("timeout after 10 s")
		case err = <-haveServicesJSON:
			// probably success, but we check it anyway
		}
		check(err, "fetching services.json")
		nodename := fmt.Sprintf("%s-%d", netname, node)

		ClientURL, err = getService("rpc addr", nodename,
			"networks",
			netname,
			"nodes",
			nodename,
			"rpc",
		)
		if err != nil {
			return nil, errors.Wrap(err, "invalid services.json for "+nodename)
		}

		node0name := fmt.Sprintf("%s-0", netname)
		RecoveryURL, err = getService("recovery addr", node0name,
			"recovery",
			netname,
			"nodes",
			node0name,
			"api",
		)
	}

	// we have a URL object
	// ignore any path; we'll supply our own, externally
	ClientURL.Path = ""

	// Note - Vle: Undocumented breaking changes from tendermint v0.32 -> v0.33
	//             return type in v0.33 is value of type (*client.HTTP, error)
	// return client.NewHTTP(ClientURL.String(), "/websocket"), nil
	return client.NewHTTP(ClientURL.String(), "/websocket")
}
