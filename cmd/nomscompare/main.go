package main

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/attic-labs/noms/go/spec"
	nt "github.com/attic-labs/noms/go/types"
	cli "github.com/jawher/mow.cli"
	util "github.com/oneiro-ndev/noms-util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		cli.Exit(1)
	}
}

func checkc(err error, context string) {
	check(errors.Wrap(err, context))
}

func bail(err string) {
	check(errors.New(err))
}

func main() {
	app := cli.App("nomscompare", "compare two noms datasets")
	app.LongDesc = `
Recursively compares data.

For help specifying your datasets, see
https://github.com/attic-labs/noms/blob/master/doc/spelling.md`

	ds1 := app.StringArg("DATASET1", "", "first dataset")
	ds2 := app.StringArg("DATASET2", "", "second dataset")
	verbose := app.BoolOpt("v verbose", false, "emit additional output")

	app.Action = func() {
		log.SetLevel(log.InfoLevel)
		if *verbose {
			log.SetLevel(log.DebugLevel)
		}
		compareDS(*ds1, *ds2)
		log.Info("done")
	}
	app.Run(os.Args)
}

func validateInput(ds1, ds2 string) {
	var errs []string
	if ds1 == "" {
		errs = append(errs, "DATASET1 must be set")
	}
	if ds2 == "" {
		errs = append(errs, "DATASET2 must be set")
	}
	if ds1 == ds2 {
		errs = append(errs, "There are never any diffs when DATASET1 == DATASET2")
	}
	if len(errs) > 0 {
		bail(strings.Join(errs, "\n"))
	}
}

func compareDS(ds1, ds2 string) {
	validateInput(ds1, ds2)

	spec1, err := spec.ForDataset(ds1)
	checkc(err, ds1)
	defer spec1.Close()
	spec2, err := spec.ForDataset(ds2)
	checkc(err, ds2)
	defer spec2.Close()

	headref1, ok1 := spec1.GetDataset().MaybeHeadRef()
	headref2, ok2 := spec2.GetDataset().MaybeHeadRef()

	log.SetFormatter(new(log.JSONFormatter))
	logger := log.WithFields(log.Fields{
		"dataset a": ds1,
		"dataset b": ds2,
	})

	if !(ok1 && ok2) {
		logger.WithFields(log.Fields{
			"a has head": ok1,
			"b has head": ok2,
		}).Fatal("not both datasets have heads")
	}

	if headref1.Height() != headref2.Height() {
		logger.WithFields(log.Fields{
			"a height": headref1.Height(),
			"b height": headref2.Height(),
		}).Fatal("heights do not match")
	}

	val1 := spec1.GetDataset().HeadValue()
	val2 := spec2.GetDataset().HeadValue()
	compare(val1, val2, ".", logger)
}

func compare(a, b nt.Value, path string, logger log.FieldLogger) {
	logger = logger.WithField("path", path)
	logger.Debug("comparing")

	errs := func(erra, errb error, context string) bool {
		anerr := erra != nil || errb != nil
		if anerr {
			logger.WithFields(log.Fields{
				"err a": erra,
				"err b": errb,
			}).Error(context)
		}
		return anerr
	}

	at := reflect.TypeOf(a)
	bt := reflect.TypeOf(b)
	if at != bt {
		logger = logger.WithFields(log.Fields{
			"a type": at,
			"b type": bt,
		})
		logger.Info("type mismatch")
		return
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
				"a value": aby,
				"b value": bby,
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
				"len(a)": alen,
				"len(b)": blen,
			}).Info("mismatch")
			return
		}

		for idx := uint64(0); idx < alen; idx++ {
			compare(av.Get(idx), bv.Get(idx), fmt.Sprintf("%s[%d]", path, idx), logger)
		}

	// TODO: all the rest of the types

	default:
		log.WithField("type", at).Error("unknown type")
	}
}
