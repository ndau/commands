package main

import (
	"sort"
	"strings"

	"github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
)

var txnames map[string]metatx.Transactable

func init() {
	// initialize txnames map
	txnames = make(map[string]metatx.Transactable)
	// add all tx full names
	for _, example := range ndau.TxIDs {
		txnames[strings.ToLower(metatx.NameOf(example))] = example
	}
	// add common abbreviations
	txnames["rfe"] = ndau.TxIDs[3]          // releasefromendowment
	txnames["claim"] = ndau.TxIDs[10]       // claimaccount
	txnames["claim-child"] = ndau.TxIDs[21] // claimchildaccount
	txnames["nnr"] = ndau.TxIDs[13]         // nominatenodereward
	txnames["cvc"] = ndau.TxIDs[16]         // commandvalidatorchange
	txnames["sidechain"] = ndau.TxIDs[17]   // sidechaintx
}

func knownNames() []string {
	out := make([]string, 0, len(txnames))
	for n := range txnames {
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
