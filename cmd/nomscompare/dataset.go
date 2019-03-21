package main

import (
	"os"
	"strings"

	"github.com/attic-labs/noms/go/datas"
	"github.com/attic-labs/noms/go/spec"
	nt "github.com/attic-labs/noms/go/types"
	log "github.com/sirupsen/logrus"
)

func validateInput(dsa, dsb string) {
	var errs []string
	if dsa == "" {
		errs = append(errs, "DATASET_A must be set")
	}
	if dsb == "" {
		errs = append(errs, "DATASET_B must be set")
	}
	if dsa == dsb {
		errs = append(errs, "There are never any diffs when DATASET_A == DATASET_B")
	}
	if len(errs) > 0 {
		bail(strings.Join(errs, "\n"))
	}
}

func compareDS(dsa, dsb string, height int, nodeHeight int) {
	validateInput(dsa, dsb)

	log.SetFormatter(new(log.JSONFormatter))
	log.SetOutput(os.Stdout)
	logger := log.WithFields(log.Fields{
		"a dataset": dsa,
		"b dataset": dsb,
	})

	speca, err := spec.ForDataset(dsa)
	checkc(err, dsa)
	defer speca.Close()
	specb, err := spec.ForDataset(dsb)
	checkc(err, dsb)
	defer specb.Close()

	refa, ok1 := speca.GetDataset().MaybeHeadRef()
	refb, ok2 := specb.GetDataset().MaybeHeadRef()

	if !(ok1 && ok2) {
		logger.WithFields(log.Fields{
			"a has head": ok1,
			"b has head": ok2,
		}).Fatal("not both datasets have heads")
	}

	dba := speca.GetDatabase()
	defer dba.Close()
	dbb := specb.GetDatabase()
	defer dbb.Close()

	if height >= 0 {
		refa = seekHeight(
			uint64(height),
			dba, refa,
			func(db datas.Database, ref nt.Ref) uint64 { return ref.Height() },
			logger.WithFields(log.Fields{
				"dataset": "a",
				"seek":    "noms",
			}),
		)
		refb = seekHeight(
			uint64(height),
			dbb, refb,
			func(db datas.Database, ref nt.Ref) uint64 { return ref.Height() },
			logger.WithFields(log.Fields{
				"dataset": "b",
				"seek":    "noms",
			}),
		)
	}

	if nodeHeight >= 0 {
		refa = seekHeight(
			uint64(nodeHeight),
			dba, refa,
			metanodeheight,
			logger.WithFields(log.Fields{
				"dataset": "a",
				"seek":    "node",
			}),
		)
		refb = seekHeight(
			uint64(nodeHeight),
			dbb, refb,
			metanodeheight,
			logger.WithFields(log.Fields{
				"dataset": "b",
				"seek":    "node",
			}),
		)
	}

	if height < 0 && nodeHeight < 0 && refa.Height() != refb.Height() {
		logger.WithFields(log.Fields{
			"a height": refa.Height(),
			"b height": refb.Height(),
		}).Fatal("heights do not match")
	}

	vala := valueAt(dba, refa)
	valb := valueAt(dbb, refb)
	compare(vala, valb, "", logger)
}
