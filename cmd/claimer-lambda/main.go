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

	"github.com/akrylysov/algnhsa"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	claimer "github.com/oneiro-ndev/commands/cmd/claimer/claimerlib"
	"github.com/oneiro-ndev/rest"
	log "github.com/sirupsen/logrus"
)

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
	bucket := os.Getenv("S3_CONFIG_BUCKET")
	path := os.Getenv("S3_CONFIG_PATH")

	sess, err := session.NewSession()
	check(err, "creating session")
	s3client := s3.New(sess)
	configData, err := s3client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
	})
	check(err, "fetching config data from s3")
	defer configData.Body.Close()

	config, err := claimer.LoadConfigData(configData.Body)
	check(err, "loading configuration")

	svc := claimer.NewClaimService(config, getLogger(sess))
	svc.GetLogger().WithField("node address", svc.Config.NodeRPC).Info("using RPC address")
	{
		fields := log.Fields{}
		for addr, keys := range svc.Config.Nodes {
			fields[addr] = len(keys)
		}
		svc.GetLogger().WithFields(fields).Info("qty keys known per known node")
	}

	cf := rest.DefaultConfig()
	cf.Load()
	server := rest.StandardSetup(cf, svc)
	if server != nil {
		rest.WatchSignals(nil, rest.FatalFunc(svc, "SIGINT"), rest.FatalFunc(svc, "SIGTERM"))
		algnhsa.ListenAndServe(server.Handler, nil)
	}
}
