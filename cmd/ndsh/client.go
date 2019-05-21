package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/tendermint/tendermint/rpc/client"
)

const servicesURL = "https://s3.us-east-2.amazonaws.com/ndau-json/services.json"

var (
	// this is a hilarious type
	servicesJSON     map[string]map[string]map[string]map[string]map[string]string
	haveServicesJSON chan struct{}
)

// on init, async get the services list
// just try to read from haveServicesJson before trying to use it
func init() {
	haveServicesJSON = make(chan struct{})
	go func() {
		resp, err := http.Get(servicesURL)
		check(err, "getting services map")
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		check(err, "reading services data")
		err = json.Unmarshal(data, &servicesJSON)
		check(err, "unmarshaling services data")
		close(haveServicesJSON)
	}()
}

func getClient(network string, node int) client.ABCIClient {
	if node < 0 {
		bail(fmt.Sprintf("invalid node: %d", node))
	}
	var netname string
	var u *url.URL
	var err error
	switch strings.ToLower(network) {
	case "main", "mainnet":
		netname = "mainnet"
	case "test", "testnet":
		netname = "testnet"
	case "dev", "devnet":
		netname = "devnet"
	case "local", "localnet":
		u, err = url.Parse(fmt.Sprintf("http://localhost:%d", 26670+node))
		check(err, "bad code in net.go: couldn't parse localnet url")
	default:
		u, err = url.Parse(network)
		if err != nil {
			// suppress the actual error, but use our own
			bail(fmt.Sprintf("invalid URL: %s", network))
		}
	}

	if u == nil {
		// user specified a symbolic name, which means we have to wait for
		// the AWS query to resolve
		<-haveServicesJSON
		nws, ok := servicesJSON["networks"]
		if !ok {
			bail("networks key not in services.json")
		}
		netjs, ok := nws[netname]
		if !ok {
			bail("%s not found in services networks list", netname)
		}
		nodes, ok := netjs["nodes"]
		if !ok {
			bail("nodes key not found in %s services list", netname)
		}
		nodename := fmt.Sprintf("%s-%d", netname, node)
		svcs, ok := nodes[nodename]
		if !ok {
			bail("%s not found in nodes list", nodename)
		}
		rpc, ok := svcs["rpc"]
		if !ok {
			bail("bad services.json: rpc key not found in %s", nodename)
		}
		u, err = url.Parse(rpc)
		check(err, "invalid services.json for %s", nodename)
	}

	// we have a URL object
	// ignore any path; we'll supply our own, externally
	u.Path = ""
	return client.NewHTTP(u.String(), "/websocket")
}
