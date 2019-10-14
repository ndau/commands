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
	"io"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
)

// Config configures the claimer
type Config struct {
	// URL to the RPC address of a node
	NodeRPC string `toml:"node_rpc"`
	// The Nodes map actually maps an address to a list of private keys
	Nodes map[string][]string `toml:"nodes"`
	// If set and true, operate in synchronous mode: wait for the blockchain
	// to return before returning from a request.
	SyncMode *bool `toml:"sync_mode"`
}

// DefaultConfigPath is the default expected path for the claimer's configuration
//
// This is a relative path, and can be prepended as necessary
const DefaultConfigPath = "claimer_conf.toml"

// LoadConfig loads the configuration data
func LoadConfig(path string) (*Config, error) {
	config := new(Config)
	_, err := toml.DecodeFile(path, config)
	if os.ExpandEnv("$CLAIMER_SYNC_MODE") == "1" {
		tru := true
		config.SyncMode = &tru
	}
	return config, err
}

// LoadConfigData loads the configuration data from a streaming reader
//
// This is useful if for example the data is not stored in the local filesystem
func LoadConfigData(data io.Reader) (*Config, error) {
	config := new(Config)
	_, err := toml.DecodeReader(data, config)
	return config, err
}

// GetKeys gets the configured keys for a given address, and whether they are configured at all
func (c Config) GetKeys(addr address.Address) (keys []signature.PrivateKey, exists bool, err error) {
	var ks []string
	ks, exists = c.Nodes[addr.String()]
	keys = make([]signature.PrivateKey, 0, len(ks))
	var pk *signature.PrivateKey
	for _, k := range ks {
		pk, err = signature.ParsePrivateKey(k)
		if err != nil {
			return
		}
		keys = append(keys, *pk)
	}
	return
}
