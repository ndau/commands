package main

// Filters for logging task output to honeycomb.

import (
	"encoding/json"
	"io"
	"os"

	"github.com/oneiro-ndev/o11y/pkg/honeycomb"
	"github.com/oneiro-ndev/writers/pkg/bufio"
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

	// We'll have at most two interpreters per task,
	// and every case gets required-fields and last-chance interpreters.
	interpreters := make([]filter.Interpreter, 0, 4)

	switch taskName {
	case rootTaskName:
		// The root task uses a json splitter with generic json interpreter.
		splitter = filter.JSONSplit
		interpreters = append(interpreters, filter.JSONInterpreter{})
	case redisTaskName:
		// The redis uses a line splitter with redis line interpreter.
		splitter = bufio.ScanLines
		interpreters = append(interpreters, filter.RedisInterpreter{})
	case nomsTaskName:
		// The noms uses a line splitter with no special interpreter.
		splitter = bufio.ScanLines
	case ndaunodeTaskName:
		// The ndaunode uses a json splitter with generic json interpreter.
		splitter = filter.JSONSplit
		interpreters = append(interpreters, filter.JSONInterpreter{})
	case tendermintTaskName:
		// The tendermint uses a json splitter with tendermint json interpreter.
		splitter = filter.JSONSplit
		interpreters = append(interpreters, filter.JSONInterpreter{})
		interpreters = append(interpreters, filter.TendermintInterpreter{})
	case ndauapiTaskName:
		// The ndauapi uses a json splitter with generic json interpreter.
		splitter = filter.JSONSplit
		interpreters = append(interpreters, filter.JSONInterpreter{})
	default:
		// Generic tasks use line splitters with no special interpreters.
		splitter = bufio.ScanLines
	}

	// Putting this one after the first interpreter will prevent tasks from overriding the values
	// of the required fields, but we don't expect to ever want to for "bin" and "node_id".
	interpreters = append(interpreters, filter.RequiredFieldsInterpreter{
		Defaults: map[string]interface{}{
			"bin":     taskName,
			"node_id": os.Getenv("NODE_ID"),
		},
	})
	interpreters = append(interpreters, filter.LastChanceInterpreter{})

	// All filters use this simple outputter that writes json blobs to honeycomb.
	// It ignores errors; we don't want to bring down procmon with a panic if we can't log
	// something, and we can't log errors if logging itself is failing.
	outputter := func(m map[string]interface{}) {
		p, err := json.Marshal(m)
		if err == nil {
			// FIXME: testing
			fixme.Write(p)
			//honeycombWriter.Write(p)
		}
	}

	return filter.NewFilter(splitter, outputter, nil, interpreters...)
}

var fixme, _ = os.OpenFile("/image/logs/all.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
