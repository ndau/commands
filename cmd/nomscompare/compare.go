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
	"bytes"
	"encoding/base64"
	"fmt"
	"reflect"
	"sort"

	nt "github.com/attic-labs/noms/go/types"
	util "github.com/ndau/noms-util"
	log "github.com/sirupsen/logrus"
)

func compare(a, b nt.Value, path string, rlogger log.FieldLogger) {
	// logger is the logger within this function; it has fields appropriate
	// to this node, but not to child nodes.
	// rlogger is the logger passed to recursive child calls.
	dotpath := path
	if path == "" {
		dotpath = "."
	}
	logger := rlogger.WithField("path", dotpath)

	at := reflect.TypeOf(a)
	bt := reflect.TypeOf(b)
	if at != bt {
		logger.WithFields(log.Fields{
			"a type": at.String(),
			"b type": bt.String(),
		}).Info("type mismatch")
		return
	}
	logger = logger.WithField("type", at.String())
	logger.Debug("comparing")

	errs := func(erra, errb error, context string) bool {
		anerr := erra != nil || errb != nil
		if anerr {
			logger.WithFields(log.Fields{
				"a err": erra,
				"b err": errb,
			}).Error(context)
		}
		return anerr
	}

	// now walk the keys lists, comparing items with equal keys,
	// making note of items present in only one list
	walkcompare := func(
		akeys, bkeys []string,
		a, b nt.Value,
		itemtype string,
		getsubitem func(value nt.Value, key string) nt.Value,
		subpath func(key string) string,
	) {
		ai := 0
		bi := 0
		for ai < len(akeys) || bi < len(bkeys) {
			switch {
			case ai >= len(akeys): // || bkeys[bi] < akeys[ai]:
				logger.WithField(itemtype, bkeys[bi]).Info(itemtype + " present in b and not a")
				bi++
			case bi >= len(bkeys): // || akeys[ai] < bkeys[bi]:
				logger.WithField(itemtype, akeys[ai]).Info(itemtype + " present in a and not b")
				ai++
			default:
				if getsubitem != nil && subpath != nil {
					compare(
						getsubitem(a, akeys[ai]),
						getsubitem(b, bkeys[bi]),
						subpath(akeys[ai]),
						rlogger,
					)
				}
				ai++
				bi++
			}
		}
	}

	// because we know the types are equal, we can get away with a type-switch
	// and type assertion to keep them in sync without too much boilerplate
	switch av := a.(type) {
	case nt.Blob:
		bv := b.(nt.Blob)
		aby, erra := util.Unblob(av)
		bby, errb := util.Unblob(bv)
		if errs(erra, errb, "unblobbing") {
			return
		}
		if !bytes.Equal(aby, bby) {
			logger.WithFields(log.Fields{
				"a value": base64.StdEncoding.EncodeToString(aby),
				"b value": base64.StdEncoding.EncodeToString(bby),
			}).Info("mismatch")
		}

	case nt.Bool:
		bv := b.(nt.Bool)
		if bool(av) != bool(bv) {
			logger.WithFields(log.Fields{
				"a value": bool(av),
				"b value": bool(bv),
			}).Info("mismatch")
		}

	case nt.List:
		bv := b.(nt.List)

		alen := av.Edit().Len()
		blen := bv.Edit().Len()
		if alen != blen {
			logger.WithFields(log.Fields{
				"a len": alen,
				"b len": blen,
			}).Info("mismatch")
			return
		}

		for idx := uint64(0); idx < alen; idx++ {
			compare(av.Get(idx), bv.Get(idx), fmt.Sprintf("%s[%d]", path, idx), rlogger)
		}

	case nt.Map:
		bv := b.(nt.Map)

		akeys, erra := mapKeys(av)
		bkeys, errb := mapKeys(bv)
		if errs(erra, errb, "converting map keys to strings") {
			return
		}

		if len(akeys) != len(bkeys) {
			logger.WithFields(log.Fields{
				"a len": len(akeys),
				"b len": len(bkeys),
			}).Info("mismatch")
		}

		walkcompare(
			akeys, bkeys,
			a, b,
			"key",
			func(v nt.Value, k string) nt.Value {
				return v.(nt.Map).Get(nt.String(k))
			},
			func(k string) string {
				return fmt.Sprintf("%s[\"%s\"]", path, k)
			},
		)

	case nt.Number:
		bv := b.(nt.Number)

		if float64(av) != float64(bv) {
			logger.WithFields(log.Fields{
				"a value": float64(av),
				"b value": float64(bv),
			}).Info("mismatch")
		}

	case nt.Set:
		bv := b.(nt.Set)

		aitems, erra := setItems(av)
		bitems, errb := setItems(bv)
		if errs(erra, errb, "getting items from set") {
			return
		}

		if len(aitems) != len(bitems) {
			logger.WithFields(log.Fields{
				"a len": len(aitems),
				"b len": len(bitems),
			}).Info("mismatch")
		}

		walkcompare(aitems, bitems, a, b, "item", nil, nil)

	case nt.String:
		bv := b.(nt.String)

		if string(av) != string(bv) {
			logger.WithFields(log.Fields{
				"a value": string(av),
				"b value": string(bv),
			}).Info("mismatch")
		} else {
			//			logger.WithField("string_value", string(av)).Info("string_value")
		}

	case nt.Struct:
		bv := b.(nt.Struct)

		afields := structFields(av)
		bfields := structFields(bv)

		if len(afields) != len(bfields) {
			logger.WithFields(log.Fields{
				"a len": len(afields),
				"b len": len(bfields),
			}).Info("mismatch")
		}

		walkcompare(
			afields, bfields,
			a, b,
			"field",
			func(v nt.Value, k string) nt.Value {
				return v.(nt.Struct).Get(k)
			},
			func(s string) string {
				return fmt.Sprintf("%s.%s", path, s)
			},
		)

	default:
		logger.Error("unknown type")
	}
}

func mapKeys(m nt.Map) (keys []string, err error) {
	m.Iter(func(k, v nt.Value) (stop bool) {
		ks, ok := k.(nt.String)
		if !ok {
			stop = true
			err = fmt.Errorf(
				"found non-string key type '%s' in map",
				reflect.TypeOf(k),
			)
			return
		}
		keys = append(keys, string(ks))
		return
	})
	sort.Strings(keys)
	return
}

func setItems(s nt.Set) (items []string, err error) {
	s.Iter(func(i nt.Value) (stop bool) {
		is, ok := i.(nt.String)
		if !ok {
			stop = true
			err = fmt.Errorf(
				"found non-string item type '%s' in set",
				reflect.TypeOf(i),
			)
			return
		}
		items = append(items, string(is))
		return
	})
	sort.Strings(items)
	return
}

func structFields(s nt.Struct) (fields []string) {
	s.IterFields(func(name string, v nt.Value) (stop bool) {
		fields = append(fields, name)
		return false
	})
	sort.Strings(fields)
	return
}
