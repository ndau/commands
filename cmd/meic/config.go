package main

import (
	"github.com/BurntSushi/toml"
	"github.com/oneiro-ndev/commands/cmd/meic/ots/bitmart"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
)

// Config defines the configuration that the MEIC can have
type Config struct {
	// DepthOverrides allows specification of a custom depth level for each
	// OTS. Keys in this table are 0-based indices within the `otsImpls` list.
	// OTSs which do not appear in this list get the default level from the
	// command line.
	DepthOverrides map[uint]uint
}

// LoadConfig loads configuration data from a given path
func LoadConfig(confPath string) (*Config, error) {
	var c Config
	_, err := toml.DecodeFile(confPath, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

type args struct {
	bitmart.BMArgs

	ServerAddr        string               `arg:"-s,--server-addr,required"`
	ServerPubKey      signature.PublicKey  `arg:"-p,--server-pub-key,required"`
	ServerPvtKey      signature.PrivateKey `arg:"-P,--server-pvt-key,required"`
	NodeAddr          string               `arg:"-n,--node-addr,required"`
	DefaultStackDepth uint                 `arg:"-d,--default-stack-depth"`
	ConfPath          string               `arg:"-c,--config-path"`
}

func (a args) GetBMArgs() bitmart.BMArgs {
	return a.BMArgs
}

var _ bitmart.HasBMArgs = (*args)(nil)
