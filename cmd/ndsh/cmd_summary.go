package main

import (
	"encoding/json"
	"sync"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

// Summary displays version information
type Summary struct{}

var _ Command = (*Summary)(nil)

// Name implements Command
func (Summary) Name() string { return "summary sib status info" }

// Run implements Command
func (Summary) Run(argvs []string, sh *Shell) (err error) {
	args := struct{}{}

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
	sh.Write(string(js))
	return nil
}
