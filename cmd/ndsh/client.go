package main

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
	// this is a hilarious type
	servicesJSON     map[string]map[string]map[string]map[string]map[string]string
	haveServicesJSON chan error

	// ClientURL stores client URL currently in use
	ClientURL *url.URL
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

func getClient(network string, node int) (client.ABCIClient, error) {
	if node < 0 {
		return nil, fmt.Errorf("invalid node: %d", node)
	}
	var netname string
	var err error
	switch strings.ToLower(network) {
	case "main", "mainnet":
		netname = "mainnet"
		ClientURL = nil
	case "test", "testnet":
		netname = "testnet"
		ClientURL = nil
	case "dev", "devnet":
		netname = "devnet"
		ClientURL = nil
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
		nws, ok := servicesJSON["networks"]
		if !ok {
			return nil, errors.New("networks key not in services.json")
		}
		netjs, ok := nws[netname]
		if !ok {
			return nil, fmt.Errorf("%s not found in services networks list", netname)
		}
		nodes, ok := netjs["nodes"]
		if !ok {
			return nil, fmt.Errorf("nodes key not found in %s services list", netname)
		}
		nodename := fmt.Sprintf("%s-%d", netname, node)
		svcs, ok := nodes[nodename]
		if !ok {
			return nil, fmt.Errorf("%s not found in nodes list", nodename)
		}
		rpc, ok := svcs["rpc"]
		if !ok {
			return nil, fmt.Errorf("bad services.json: rpc key not found in %s", nodename)
		}
		if !strings.HasPrefix(rpc, "http") {
			rpc = "https://" + rpc
		}
		ClientURL, err = url.Parse(rpc)
		if err != nil {
			return nil, errors.Wrap(err, "invalid services.json for "+nodename)
		}
	}

	// we have a URL object
	// ignore any path; we'll supply our own, externally
	ClientURL.Path = ""
	return client.NewHTTP(ClientURL.String(), "/websocket"), nil
}
