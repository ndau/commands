package main

import (
	"log"

	"github.com/oneiro-ndev/commands/cmd/meic/ots"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	ndaumath "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/recovery/pkg/signer"
)

type Account struct {
	Balance        ndaumath.Ndau         `json:"balance"`
	ValidationKeys []signature.PublicKey `json:"validationKeys"`
	Sequence       uint64                `json:"sequence"`
}

// handle a sale of ndau by creating and sending an Issue tx
func (ius *IssuanceUpdateSystem) handleSale(sale ots.TargetPriceSale, sigserv *signer.ServerManager) {
	// - get the current issuance sequence from the blockchain
	// var issuer address.Address
	// err := tool.Sysvar(ius.tmNode, sv.ReleaseFromEndowmentAddressName, &issuer)
	rfeAddress := struct {
		ReleaseFromEndowmentAddress []string `json:"ReleaseFromEndowmentAddress"`
	}{}

	err := ius.reqMan.Get("/system/get/ReleaseFromEndowmentAddress", &rfeAddress)
	if err != nil {
		check(err, "failed to get sysvar with address of issuer")
	}
	log.Println("RFE address = ", rfeAddress)

	// acctData, _, err := tool.GetAccount(ius.tmNode, issuer)
	amap := make(map[string]Account)
	err = ius.reqMan.Get("/account/account/"+rfeAddress.ReleaseFromEndowmentAddress[0], &amap)
	if err != nil {
		check(err, "failed to get issuer account data")
	}
	log.Println("account data = ", amap)
	sequence := amap[rfeAddress.ReleaseFromEndowmentAddress[0]].Sequence
	// - generate the issuance tx
	tx := ndau.NewIssue(sale.Qty, sequence+1)
	// - send the tx to the signature server
	// - update the tx with the returned signatures
	tx.Signatures = sigserv.SignTx(tx, amap[rfeAddress.ReleaseFromEndowmentAddress[0]].ValidationKeys)
	// - send it to the blockchain using ndau API

	retVal := struct {
		Hash string `json:"hash"`
		Msg  string `json:"msg"`
		Code int    `json:"code"`
	}{}

	err = ius.reqMan.Post("/tx/submit/Issue", tx, &retVal)
	if err != nil {
		check(err, "sending issue tx to blockchain")
	}
	log.Println("retval = ", retVal)
}
