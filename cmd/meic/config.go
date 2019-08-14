package main

import (
	"github.com/BurntSushi/toml"
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
