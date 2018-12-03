package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/pkg/errors"
)

func returnError(w http.ResponseWriter, err error) {
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
	}
}
func assignOption(opt string) *string {
	var res *string
	if opt != "" {
		res = new(string)
		*res = opt
	} else {
		res = nil
	}
	return res
}

// getKeyClosure sets the appropriate options for a command to get the key
// using a variety of argument styles.
func getKeyClosureAPI(w http.ResponseWriter, r *http.Request, allowBinary bool) func() []byte {
	var key, kf *string = nil, nil
	var bkey, bkf *string = nil, nil

	if r != nil {
		v := r.URL.Query()
		key = assignOption(v.Get("key"))
		kf = assignOption(v.Get("key-file"))
		if allowBinary {
			bkey = assignOption(v.Get("binary-key"))
			bkf = assignOption(v.Get("binary-key-file"))
		}
	}

	return func() []byte {
		if key != nil && *key != "" {
			// key must be valid utf-8
			bytes := []byte(*key)
			validateBytes(bytes)
			return bytes
		}
		if kf != nil && *kf != "" {
			k, err := ioutil.ReadFile(*kf)
			returnError(w, errors.Wrap(err, "bad key file"))
			// key must be valid utf-8
			validateBytes(k)
			return k
		}
		if allowBinary {
			if bkey != nil && *bkey != "" {
				b, err := base64.StdEncoding.DecodeString(*bkey)
				returnError(w, errors.Wrap(err, "bad binary key"))
				return b
			}
			if bkf != nil && *bkf != "" {
				k, err := ioutil.ReadFile(*bkf)
				returnError(w, errors.Wrap(err, "bad binary key file"))
				return k
			}
		}

		// we shouldn't get here; one of the above key flags should have been set
		returnError(w, errors.New("No key set"))
		return []byte{}
	}
}

// getHeightClosure sets the appropriate options for a command to get the
// appropriate moment in history
func getHeightClosureAPI(w http.ResponseWriter, r *http.Request) func() uint64 {
	var height *int

	if r != nil {
		v := r.URL.Query()
		height64, err := strconv.ParseInt(v.Get("height"), 10, 64)
		if err != nil {
			height64 = 0
		}
		height = new(int)
		*height = int(height64)
	}

	return func() uint64 {
		var h uint64
		if height != nil {
			h = uint64(*height)
		}
		return h
	}
}

/*
func getNamespaceClosureAPI(w http.ResponseWriter, r *http.Request) func(*tool.Config) (namespace []byte) {
	var sys *bool
	var name, ns *string = nil, nil

	if r != nil {
		v := r.URL.Query()
		res, err := strconv.ParseBool(v.Get("sys"))
		if err != nil {
			sys = new(bool)
			*sys = res
		}
		name = assignOption(v.Get("name"))
		ns = assignOption(v.Get("ns"))
	}

	return func(c *tool.Config) (namespace []byte) {
		if *sys {
			return cns.System
		}

		if name != nil && len(*name) > 0 {
			identity, found := c.Identities[*name]
			if !found {
				returnError(w, fmt.Errorf("Identity '%s' not found", *name))
			}
			return identity.PublicKey
		}

		if ns != nil && len(*ns) > 0 {
			nsB, err := base64.StdEncoding.DecodeString(*ns)
			returnError(w, err)

			if len(nsB) != cns.Size {
				returnError(w, fmt.Errorf(
					"namespace must have size %d; found %d",
					cns.Size, len(nsB),
				))
			}

			return nsB
		}

		returnError(w, fmt.Errorf("Namespace error in getNamespaceClosure"))
		return []byte{}
	}
}
*/
// getValueClosure sets the appropriate options for a command to get the value
// using a variety of argument styles.
func getValueClosureAPI(w http.ResponseWriter, r *http.Request) func() []byte {
	var value, vf, bval, bvf *string = nil, nil, nil, nil

	if r != nil {
		v := r.URL.Query()
		value = assignOption(v.Get("value"))
		vf = assignOption(v.Get("value-file"))
		bval = assignOption(v.Get("binary-value"))
		bvf = assignOption(v.Get("binary-value-file"))
	}
	return func() []byte {
		if value != nil && *value != "" {
			// must be valid utf-8
			bytes := []byte(*value)
			validateBytes(bytes)
			return bytes
		}
		if vf != nil && *vf != "" {
			k, err := ioutil.ReadFile(*vf)
			returnError(w, errors.Wrap(err, "Bad value file"))
			// must be valid utf-8
			validateBytes(k)
			return k
		}

		if bval != nil && *bval != "" {
			b, err := base64.StdEncoding.DecodeString(*bval)
			returnError(w, errors.Wrap(err, "Bad value string"))
			return b
		}
		if bvf != nil && *bvf != "" {
			k, err := ioutil.ReadFile(*bvf)
			returnError(w, errors.Wrap(err, "Bad value file"))
			return k
		}

		// if no value option is set, return an empty slice
		return make([]byte, 0)
	}
}

func encodeJSON(w http.ResponseWriter, result interface{}, err error) {
	if err == nil {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "    ")
		err = enc.Encode(result)
	} else {
		http.Error(w, "Failed to encode JSON result", http.StatusBadRequest)
	}
}

// getStatus swagger:route GET /status
//
// Display status
//
// Return status of current node
// This can be changed with the status flag.
//
// Responses:
// 		default: genericError
//		200:	 status
//
func getStatus(w http.ResponseWriter, r *http.Request) {
	config := getConfig()
	status, err := tool.Info(tmnode(config.Node))
	encodeJSON(w, status, err)
}

func getHealth(w http.ResponseWriter, r *http.Request) {
	config := getConfig()
	health, err := tmnode(config.Node).Health()
	encodeJSON(w, health, err)
}
func getNetInfo(w http.ResponseWriter, r *http.Request) {
	config := getConfig()
	netInfo, err := tmnode(config.Node).NetInfo()
	encodeJSON(w, netInfo, err)
}
func getGenesis(w http.ResponseWriter, r *http.Request) {
	config := getConfig()
	genesis, err := tmnode(config.Node).Genesis()
	encodeJSON(w, genesis, err)
}
func getABCIInfo(w http.ResponseWriter, r *http.Request) {
	config := getConfig()
	abciInfo, err := tmnode(config.Node).ABCIInfo()
	encodeJSON(w, abciInfo, err)
}

func getNumUnconfirmedTxs(w http.ResponseWriter, r *http.Request) {
	config := getConfig()
	txs, err := tmnode(config.Node).Health()
	encodeJSON(w, txs, err)
}

func getDumpConsensusState(w http.ResponseWriter, r *http.Request) {
	config := getConfig()
	consensusState, err := tmnode(config.Node).DumpConsensusState()
	encodeJSON(w, consensusState, err)
}

func getBlock(w http.ResponseWriter, r *http.Request) {
	config := getConfig()
	heightParam := r.URL.Query().Get("height")
	if heightParam != "" {
		height, err := strconv.ParseInt(heightParam, 10, 64)
		block, err := tmnode(config.Node).Block(&height)
		encodeJSON(w, block, err)
	}
}

func getBlockChain(w http.ResponseWriter, r *http.Request) {
	config := getConfig()
	minHeightParam := r.URL.Query().Get("min_height")
	maxHeightParam := r.URL.Query().Get("max_height")
	if minHeightParam != "" && maxHeightParam != "" {
		minHeight, err := strconv.ParseInt(minHeightParam, 10, 64)
		maxHeight, err := strconv.ParseInt(maxHeightParam, 10, 64)
		block, err := tmnode(config.Node).BlockchainInfo(minHeight, maxHeight)
		encodeJSON(w, block, err)

	}
}
