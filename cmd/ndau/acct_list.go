package main

import (
	"fmt"
	"os"
	"sort"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/pkg/errors"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

func getAccountList(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Command("remote", "list accounts known to the ndau chain", getAccountListRemote(verbose))

		cmd.Action = func() {
			config := getConfig()
			config.EmitAccounts(os.Stdout)
		}
	}
}

func getAccountListRemote(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Action = func() {
			config := getConfig()

			accts := make([]string, 0)

			index := 0
			const pageSize = 100

			var (
				qaccts *query.AccountListQueryResponse
				qresp  *rpctypes.ResultABCIQuery
				err    error
			)

			getPage := func() {
				qaccts, qresp, err = tool.GetAccountList(
					tmnode(config.Node, nil, nil),
					index, pageSize,
				)
				orQuit(errors.Wrap(err, fmt.Sprintf(
					"getPage(%d)", index,
				)))
				accts = append(accts, qaccts.Accounts...)
				index++
			}

			getPage()
			// we had to do it once manually in order to prime the pump
			for len(qaccts.Accounts) == pageSize {
				getPage()
			}

			// accts probably contains most of the accounts known on the blockchain
			// however, race conditions mean we can't always keep up with new
			// accounts, which in turn means that we may encounter duplicates
			// or skipped entries. There isn't a good way around this except to
			// implement cursor-based paging, which would be much more complex
			// to implement and come with its own tradeoffs.
			//
			// we can't do anything about the skipped entries (or even detect them),
			// but we can at least eliminate the duplicates
			sort.Strings(accts)
			accts = dedup(accts)

			for _, acct := range accts {
				fmt.Println(acct)
			}
			finish(*verbose, qresp, err, "account list remote")
		}
	}
}
