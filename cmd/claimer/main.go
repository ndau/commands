package main

import (
	"fmt"
	"os"

	"github.com/oneiro-ndev/rest"
	log "github.com/sirupsen/logrus"
)

const configPathS = "config-path"

func check(err error, context string, formatters ...interface{}) {
	if err != nil {
		context += ": %s"
		formatters = append(formatters, err.Error())
		fmt.Fprintf(os.Stderr, context, formatters...)
		os.Exit(1)
	}
}

func main() {
	cf := rest.DefaultConfig()
	cf.AddString(configPathS, DefaultConfigPath)
	cf.Load()

	config, err := LoadConfig(cf.GetString(configPathS))
	check(err, "loading configuration")

	svc := &claimService{
		Logger: log.New().WithField("bin", "claimer"),
		Config: config,
	}

	server := rest.StandardSetup(cf, svc)
	if server != nil {
		rest.WatchSignals(nil, rest.FatalFunc(svc, "SIGINT"), rest.FatalFunc(svc, "SIGTERM"))
		log.Fatal(server.ListenAndServe())
	}
}
