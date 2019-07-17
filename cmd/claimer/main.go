package main

import (
	"fmt"
	"os"

	claimer "github.com/oneiro-ndev/commands/cmd/claimer/claimerlib"
	"github.com/oneiro-ndev/rest"
	log "github.com/sirupsen/logrus"
)

const configPathS = "config-path"

func check(err error, context string, formatters ...interface{}) {
	if err != nil {
		if context[len(context)-1] == '\n' {
			context = context[:len(context)-1]
		}
		context += ": %s\n"
		formatters = append(formatters, err.Error())
		fmt.Fprintf(os.Stderr, context, formatters...)
		os.Exit(1)
	}
}

func main() {
	cf := rest.DefaultConfig()
	cf.AddString(configPathS, claimer.DefaultConfigPath)
	cf.Load()

	config, err := claimer.LoadConfig(cf.GetString(configPathS))
	check(err, "loading configuration")

	svc := claimer.NewClaimService(config, log.New().WithField("bin", "claimer"))
	svc.GetLogger().WithField("node address", svc.Config.NodeRPC).Info("using RPC address")
	{
		fields := log.Fields{}
		for addr, keys := range svc.Config.Nodes {
			fields[addr] = len(keys)
		}
		svc.GetLogger().WithFields(fields).Info("qty keys known per known node")
	}

	server := rest.StandardSetup(cf, svc)
	if server != nil {
		rest.WatchSignals(nil, rest.FatalFunc(svc, "SIGINT"), rest.FatalFunc(svc, "SIGTERM"))
		svc.GetLogger().Fatal(server.ListenAndServe())
	}
}
