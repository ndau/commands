package main

import (
	"log"
	"net/http"

	cli "github.com/jawher/mow.cli"
	"github.com/kentquirk/boneful"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

func server(cmd *cli.Cmd) {
	cmd.Spec = "PORT"

	port := cmd.StringArg("PORT", "", "port number for server to listen on")

	cmd.Action = func() {
		svc := new(boneful.Service).
			Path("/").
			Doc(`This service provides the API for Tendermint and Chaos/Order/ndau blockchain data`)

		svc.Route(svc.GET("/status").To(getStatus).
			Doc("Returns the status of the current node.").
			Operation("Status").
			Produces("application/json").
			Writes(rpctypes.ResultStatus{}))

		svc.Route(svc.GET("/health").To(getHealth).
			Doc("Returns the health of the current node.").
			Operation("Health").
			Produces("application/json").
			Writes(rpctypes.ResultHealth{}))

		svc.Route(svc.GET("/net_info").To(getNetInfo).
			Doc("Returns the network information of the current node.").
			Operation("Net Info").
			Produces("application/json").
			Writes(rpctypes.ResultNetInfo{}))

		svc.Route(svc.GET("/genesis").To(getGenesis).
			Doc("Returns the genesis block of the current node.").
			Operation("Genesis").
			Produces("application/json").
			Writes(rpctypes.ResultGenesis{}))

		svc.Route(svc.GET("/abci_info").To(getABCIInfo).
			Doc("Returns info on the ABCI interface.").
			Operation("ABCI Info").
			Produces("application/json").
			Writes(rpctypes.ResultABCIInfo{}))

		svc.Route(svc.GET("/num_unconfirmed_txs").To(getNumUnconfirmedTxs).
			Doc("Returns the number of unconfirmed transactions on the chain.").
			Operation("Num Unconfirmed Transactions").
			Produces("application/json").
			Writes(rpctypes.ResultStatus{}))

		svc.Route(svc.GET("/dump_consensus_state").To(getDumpConsensusState).
			Doc("Returns the current Tendermint consensus state in JSON").
			Operation("Dump Consensus State").
			Produces("application/json").
			Writes(rpctypes.ResultDumpConsensusState{}))

		svc.Route(svc.GET("/block").To(getBlock).
			Doc("Returns the block in the chain at the given height.").
			Operation("Get Block").
			Param(boneful.QueryParameter("height", "Height of the block in chain to return.").DataType("string").Required(true)).
			Produces("application/json").
			Writes(rpctypes.ResultBlock{}))

		svc.Route(svc.GET("/blockchain").To(getBlockChain).
			Doc("Returns a sequence of blocks starting at min_height and ending at max_height").
			Operation("Get Block Chain").
			Param(boneful.QueryParameter("min_height", "Height at which to begin retrieval of blockchain sequence.").DataType("string").Required(true)).
			Param(boneful.QueryParameter("max_height", "Height at which to end retrieval of blockchain sequence.").DataType("string").Required(true)).
			Produces("application/json").
			Writes(rpctypes.ResultBlockchainInfo{}))

		log.Printf("Chaos server listening on port %s\n", *port)
		server := &http.Server{Addr: ":" + *port, Handler: svc.Mux()}
		log.Fatal(server.ListenAndServe())
	}
}
