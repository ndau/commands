package main

import (
	"reflect"
	"strconv"

	"github.com/attic-labs/noms/go/datas"
	nt "github.com/attic-labs/noms/go/types"
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
