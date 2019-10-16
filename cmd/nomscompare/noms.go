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
	"encoding/hex"

	"github.com/attic-labs/noms/go/datas"
	"github.com/attic-labs/noms/go/hash"
	nt "github.com/attic-labs/noms/go/types"
	log "github.com/sirupsen/logrus"
)

func parentOf(db datas.Database, ref nt.Ref, logger log.FieldLogger) nt.Ref {
	parents := ref.TargetValue(db).(nt.Struct).Get(datas.ParentsField).(nt.Set)
	firstParent := parents.First()
	if firstParent == nil {
		logger.WithField("height", ref.Height()).Fatal("ref needed parent but had none")
	}

	psize := setSize(parents)
	if psize != 1 {
		logger.WithFields(log.Fields{
			"qty parents": psize,
			"height":      ref.Height(),
		}).Fatal("ref had too many parents")
	}

	return firstParent.(nt.Ref)
}

// setSizeEq1 returns true if the set contains exactly one element
//
// This is a dumb function to have to write, but apparently nt.Set doesn't have
// a Size() or Len() function, so we have to do this
func setSize(set nt.Set) int {
	size := 0
	set.Iter(func(nt.Value) bool {
		size++
		return false
	})
	return size
}

func seekHeight(
	target uint64,
	db datas.Database,
	ref nt.Ref,
	getHeight func(datas.Database, nt.Ref) uint64,
	logger log.FieldLogger,
) nt.Ref {
	if getHeight(db, ref) < target {
		logger.WithField("height", getHeight(db, ref)).Fatal("not tall enough")
	}

	for getHeight(db, ref) > target {
		ref = parentOf(db, ref, logger)
	}
	return ref
}

func valueAt(db datas.Database, ref nt.Ref) nt.Value {
	return ref.TargetValue(db).(nt.Struct).Get(datas.ValueField)
}

func apphash(ref nt.Ref) string {
	h := [hash.ByteLen]byte(ref.Hash())
	return hex.EncodeToString(h[:])
}
