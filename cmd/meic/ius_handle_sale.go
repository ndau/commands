package main

import (
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/recovery/pkg/signer"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
)

// handle a sale of ndau by creating and sending an Issue tx
func (ius *IssuanceUpdateSystem) handleSale(sale TargetPriceSale, sigserv *signer.ServerManager) {
	// - get the current issuance sequence from the blockchain
	var issuer address.Address
	err := tool.Sysvar(ius.tmNode, sv.ReleaseFromEndowmentAddressName, &issuer)
	if err != nil {
		check(err, "failed to get sysvar with address of issuer")
	}
	acctData, _, err := tool.GetAccount(ius.tmNode, issuer)
	if err != nil {
		check(err, "failed to get issuer account data")
	}
	// - generate the issuance tx
	tx := ndau.NewIssue(sale.Qty, acctData.Sequence+1)
	// - send the tx to the signature server
	// - update the tx with the returned signatures
	tx.Signatures = sigserv.SignTx(tx, acctData.ValidationKeys)
	// - send it to the blockchain
	// note that this is SendSync, not SendCommit. TM promises that txs
	// sent via this method have been validated by the recipient node,
	// but returns without forcing the caller to wait
	// for the chain to actually vote and update. This seems like a worthwhile
	// compromise to keep things flowing smoothly.
	_, err = tool.SendSync(ius.tmNode, tx)
	if err != nil {
		check(err, "sending issue tx to blockchain")
	}
}
