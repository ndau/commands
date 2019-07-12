package main

import (
	"github.com/BurntSushi/toml"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
)

// Config configures the claimer
type Config struct {
	// The Nodes map actually maps an address to a list of private keys
	Nodes map[string][]string

	// URL to the RPC address of a node
	NodeRPC string
}

// DefaultConfigPath is the default expected path for the claimer's configuration
//
// This is a relative path, and can be prepended as necessary
const DefaultConfigPath = "claimer_conf.toml"

// LoadConfig loads the configuration data
func LoadConfig(path string) (*Config, error) {
	config := new(Config)
	_, err := toml.DecodeFile(path, config)
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
