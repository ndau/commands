#!/usr/bin/env python3

from datetime import timedelta
from string import Template

import dateutil.parser
import requests

ACCOUNTS = [
    "ndacc2gihhrj6rhe3v2jx5k6gqpedy878eaxn35j4tvcdirq",
    "ndadyrd7u7kyjkq9nwcz3rgyi3m6fyeexwwi6hy6giwby7wy",
]

HISTORY = "http://localhost:3032/account/history/{address}"
TRANSACTION = "http://localhost:3032/transaction/{txhash}"

EPOCH = dateutil.parser.parse("2000-01-01T00:00:00Z")


def timestamp_ms(time):
    "Number of milliseconds since the epoch; ndau-style"
    return (dateutil.parser.parse(time) - EPOCH) // timedelta(milliseconds=1)


def getjs(endpoint):
    resp = requests.get(endpoint)
    resp.raise_for_status()
    return resp.json()


class Transaction:
    def __init__(self, txhash):
        self.hash = txhash
        self.data = getjs(TRANSACTION.format(txhash=txhash))


class Acct:
    def __init__(self, address):
        self.address = address
        self.history = getjs(HISTORY.format(address=self.address))["Items"]

        # it's better to just iterate directly on the list, but we're editing
        # the items in-place, which requires index-based access
        for idx in range(len(self.history)):
            txhash = self.history[idx]["TxHash"]
            self.history[idx]["tx"] = Transaction(txhash)


HEADER_TEMPLATE = """
package ndau

import (
    "encoding/base64"
    "testing"

    "github.com/oneiro-ndev/metanode/pkg/meta/app/code"
    metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
    "github.com/oneiro-ndev/ndau/pkg/ndau/backing"
    "github.com/oneiro-ndev/ndaumath/pkg/address"
    "github.com/oneiro-ndev/ndaumath/pkg/constants"
    "github.com/oneiro-ndev/ndaumath/pkg/eai"
    math "github.com/oneiro-ndev/ndaumath/pkg/types"
    "github.com/stretchr/testify/require"
    "github.com/tinylib/msgp/msgp"
)


func makeTx(t *testing.T, id int, data string) metatx.Transactable {
    datab, err := base64.StdEncoding.DecodeString(data)
    require.NoError(t, err)
    mtx := metatx.Transaction{
        Nonce: []byte{},
        TransactableID: metatx.TxID(id),
        Transactable: msgp.Raw(datab),
    }
    tx, err := mtx.AsTransactable(TxIDs)
    require.NoError(t, err)
    return tx
}
"""

TEST_TX_TEMPLATE = """
    {
        tx := makeTx(t, $id, "$data")
        resp := deliverTxAt(t, app, tx, $timestamp)
        require.Equal(t, code.OK, code.ReturnCode(resp.Code))
        acct, _ := app.getAccount(addr)
        require.Equal(t, math.Ndau($balance), acct.Balance)
    }
"""

TEST_TEMPLATE = """
func Test_${address}_History(t *testing.T) {
    app, _ := initApp(t)

    ts := math.Timestamp($creation)
    // create the account
    // from https://github.com/oneiro-ndev/genesis/blob/master/pkg/etl/transform.go
    modify(t, "$address", app, func(ad *backing.AccountData) {
        ad.Balance = 1000 * constants.NapuPerNdau
        ad.LastEAIUpdate = ts
        ad.LastWAAUpdate = ts
        ad.CurrencySeatDate = &ts
        ad.Lock = backing.NewLock($creation + math.Year, eai.DefaultLockBonusEAI)
        ad.Lock.Notify($creation, 0)
        ad.DelegationNode = &nodeAddress
        ad.RecourseSettings.Period = math.Hour
    })

    addr, err := address.Validate("$address")
    require.NoError(t, err)

    $txs
}
"""


def generate_tests():
    print(HEADER_TEMPLATE)
    tx_template = Template(TEST_TX_TEMPLATE)
    test_template = Template(TEST_TEMPLATE)
    for account in ACCOUNTS:
        acct = Acct(account)
        tx_tests = []
        for event in acct.history:
            tx_tests.append(
                tx_template.substitute(
                    id=event["tx"].data["Tx"]["TransactableID"],
                    data=event["tx"].data["TxBytes"],
                    timestamp=timestamp_ms(event["Timestamp"]),
                    balance=event["Balance"],
                )
            )
        print(
            test_template.substitute(
                address=account,
                creation=timestamp_ms("2018-04-05T00:00:00Z"),
                txs="\n".join(tx_tests),
            )
        )


if __name__ == "__main__":
    generate_tests()
