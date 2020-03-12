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
	"encoding/json"
	"sync"

	"github.com/alexflint/go-arg"
	"github.com/ndau/ndau/pkg/query"
	"github.com/ndau/ndau/pkg/tool"
	"github.com/pkg/errors"
	"github.com/savaki/jq"
)

// Summary displays version information
type Summary struct{}

var _ Command = (*Summary)(nil)

// Name implements Command
func (Summary) Name() string { return "summary sib status info" }

// Run implements Command
func (Summary) Run(argvs []string, sh *Shell) (err error) {
	args := struct {
		JQ string `help:"filter output json by this jq expression"`
	}{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	merge := make(map[string]interface{})
	m := sync.Mutex{}
	wg := sync.WaitGroup{}

	mergeitem := func(get func() (interface{}, error)) {
		defer wg.Done()
		var item interface{}
		item, err = get()
		if err != nil {
			return
		}

		var js []byte
		js, err = json.Marshal(item)
		if err != nil {
			return
		}
		var imap map[string]interface{}
		err = json.Unmarshal(js, &imap)
		if err != nil {
			return
		}

		m.Lock()
		defer m.Unlock()
		for k, v := range imap {
			merge[k] = v
		}
	}

	wg.Add(2)
	go mergeitem(func() (interface{}, error) {
		var summary *query.Summary
		summary, _, err = tool.GetSummary(sh.Node)
		return summary, err
	})
	go mergeitem(func() (interface{}, error) {
		var sib query.SIBResponse
		sib, _, err := tool.GetSIB(sh.Node)
		return sib, err
	})

	wg.Wait()
	if err != nil {
		return
	}

	var js []byte
	js, err = json.MarshalIndent(merge, "", "  ")
	if err != nil {
		return
	}

	if args.JQ != "" {
		op, err := jq.Parse(args.JQ)
		if err != nil {
			return errors.Wrap(err, "parsing JQ selector")
		}
		js, err = op.Apply(js)
		if err != nil {
			return errors.Wrap(err, "applying JQ selector")
		}
	}

	sh.Write(string(js))
	return nil
}
