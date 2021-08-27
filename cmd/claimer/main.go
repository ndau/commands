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
	"fmt"
	"os"

	claimer "github.com/ndau/commands/cmd/claimer/claimerlib"
	"github.com/ndau/rest"
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
	svc.GetLogger().WithField("node address", svc.Config.NodeAPI).Info("using API address")
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
