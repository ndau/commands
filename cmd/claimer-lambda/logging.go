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
	"encoding/base64"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/ndau/o11y/pkg/honeycomb"
	log "github.com/sirupsen/logrus"
)

func redact(s string) string {
	switch {
	case s == "":
		return "<empty>"
	case len(s) < 3:
		// too short to be a real key
		return s
	default:
		// the slices below are ensure the output type is string not byte
		return s[0:1] + strings.Repeat("*", len(s)-2) + s[len(s)-1:len(s)]
	}
}

func getLogger(sess *session.Session) *log.Entry {
	// setup logger per the environment
	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{})

	var err error
	level := log.InfoLevel
	levelS := os.Getenv("LOG_LEVEL")
	if levelS != "" {
		level, err = log.ParseLevel(levelS)
		if err != nil {
			// whatever, we're not going to fail for this
			logger.WithError(err).WithFields(log.Fields{
				"bin":       "claimer",
				"LOG_LEVEL": levelS,
			}).Error("could not parse desired log level")
			level = log.InfoLevel
		}
	}
	logger.SetLevel(level)

	le := logger.WithField("bin", "claimer")

	hcDataset := os.Getenv("HONEYCOMB_DATASET")
	hcEncKey := os.Getenv("HONEYCOMB_KEY_ENCRYPTED")

	leh := le.WithFields(log.Fields{
		"HONEYCOMB_DATASET":       hcDataset,
		"HONEYCOMB_KEY_ENCRYPTED": redact(hcEncKey),
	})
	if hcDataset != "" && hcEncKey != "" {
		hcKeyData, err := base64.StdEncoding.DecodeString(hcEncKey)
		if err != nil {
			leh.WithError(err).Error("HONEYCOMB_KEY_ENCRYPTED was not valid base64 data")
			return le
		}

		// decrypt the honeycomb key...
		kmsClient := kms.New(sess)
		hckdo, err := kmsClient.Decrypt(&kms.DecryptInput{
			CiphertextBlob: hcKeyData,
		})
		if err != nil {
			leh.WithError(err).Error("KMS decryption error")
			return le
		}

		hcKey := string(hckdo.Plaintext)
		leh = leh.WithField("HONEYCOMB_KEY", redact(hcKey))

		// then set the key in the environment...
		os.Setenv("HONEYCOMB_KEY", hcKey)

		leh.Info("setting up honeycomb")
		// then finally perform normal honeycomb setup
		return honeycomb.Setup(logger).WithField("bin", "claimer")
	}
	leh.Warn("honeycomb not configured")

	return le
}
