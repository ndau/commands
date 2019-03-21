package main

import (
	"strconv"

	"github.com/attic-labs/noms/go/datas"
	nt "github.com/attic-labs/noms/go/types"
)

func metanodeheight(db datas.Database, ref nt.Ref) uint64 {
	v, err := strconv.ParseUint(
		string(valueAt(db, ref).(nt.Struct).Get("Height").(nt.String)),
		36, 64,
	)
	if err != nil {
		panic(err)
	}
	return v
}
