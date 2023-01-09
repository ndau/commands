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
	"reflect"
	"strconv"

	"github.com/ndau/noms/go/datas"
	nt "github.com/ndau/noms/go/types"
	log "github.com/sirupsen/logrus"
)

func metanodeheight(db datas.Database, ref nt.Ref) uint64 {
	metastate, ok := valueAt(db, ref).(nt.Struct)
	if !ok {
		log.WithField("metastate type", reflect.TypeOf(valueAt(db, ref)).String()).
			Fatal("expected metastate to be a nt.Struct")
	}
	heightv, ok := metastate.MaybeGet("Height")
	if !ok {
		log.Fatal("metastate did not have a .Height field")
	}
	heights, ok := heightv.(nt.String)
	if !ok {
		log.WithField(".Height type", reflect.TypeOf(heightv).String()).
			Fatal("expected .Height to be stored as a nt.String")
	}
	v, err := strconv.ParseUint(
		string(heights),
		36, 64,
	)
	if err != nil {
		log.WithError(err).Fatal("node height not a base36 string")
	}
	return v
}
