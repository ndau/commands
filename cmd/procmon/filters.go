package main

// Filters for logging task output to honeycomb.

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/oneiro-ndev/o11y/pkg/honeycomb"
	"github.com/oneiro-ndev/writers/pkg/filter"
	"github.com/pkg/errors"
)

// Thread-safe singletons that all filters use for sending logs to honeycomb.
var useHoneycomb bool
var honeycombWriter io.Writer

// initFilters() checks whether honeycomb should be enabled and sets up a writer for it if so.
// We don't use init() since we want the caller to handle errors.
func initFilters() error {
	if os.Getenv("HONEYCOMB_KEY") != "" && os.Getenv("HONEYCOMB_DATASET") != "" {
		writer, err := honeycomb.NewWriter()
		if err != nil {
			return errors.Wrap(err, "Honeycomb env vars set but failed to create writer")
		}
		useHoneycomb = true
		honeycombWriter = writer
	}

	return nil
}

// newFilter constructs a new filter.Filter for the task with the given name.
func newFilter(taskName string) io.Writer {
	var splitter bufio.SplitFunc
	var interpreter filter.Interpreter

	switch taskName {
	case rootTaskName:
		// The root task uses a json splitter with generic json interpreter.
		splitter = filter.JSONSplit
		interpreter = filter.JSONInterpreter{}
	case redisTaskName:
		// The redis uses a line splitter with redis line interpreter.
		splitter = bufio.ScanLines
		interpreter = filter.RedisInterpreter{}
	case nomsTaskName:
		// The noms uses a line splitter with no special interpreter.
		splitter = bufio.ScanLines
	case ndaunodeTaskName:
		// The ndaunode uses a json splitter with generic json interpreter.
		splitter = filter.JSONSplit
		interpreter = filter.JSONInterpreter{}
	case tendermintTaskName:
		// The tendermint uses a json splitter with tendermint json interpreter.
		splitter = filter.JSONSplit
		interpreter = filter.TendermintInterpreter{}
	case ndauapiTaskName:
		// The ndauapi uses a json splitter with generic json interpreter.
		splitter = filter.JSONSplit
		interpreter = filter.JSONInterpreter{}
	default:
		// Generic tasks use line splitters with no special interpreters.
		splitter = bufio.ScanLines
	}

	// We'll have at most one interpreter from above,
	// and every case gets required-fields and last-chance interpreters.
	interpreters := make([]filter.Interpreter, 0, 3)
	if interpreter != nil {
		interpreters = append(interpreters, interpreter)
	}
	// Putting this one after the first interpreter will prevent tasks from overriding the values
	// of the required fields, but we don't expect to ever want to for "node_id" and "bin".
	interpreters = append(interpreters, filter.RequiredFieldsInterpreter{
		// We don't have a logrus entry at this point.  Each task's logging winds up as json here.
		// So we mimic some of the fields that o11y/pkg/honeycomb/honeycomb.com::Fire() adds.
		Defaults: map[string]interface{}{
			"bin":      taskName,
			"node_id":  os.Getenv("NODE_ID"),
			"log_time": time.Now(),
		},
	})
	interpreters = append(interpreters, filter.LastChanceInterpreter{})

	// All filters use this simple outputter that writes json blobs to honeycomb.
	// It ignores errors; we don't want to bring down procmon with a panic if we can't log
	// something, and we can't log errors if logging itself is failing.
	outputter := func(m map[string]interface{}) {
		p, err := json.Marshal(m)
		if err == nil {
			honeycombWriter.Write(p)
		}
	}

	return filter.NewFilter(splitter, outputter, nil, interpreters...)
}
