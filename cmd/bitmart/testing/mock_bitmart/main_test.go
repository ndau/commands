package main

import (
	"fmt"
	"testing"

	bitmart "github.com/oneiro-ndev/commands/cmd/bitmart/api"
	"github.com/stretchr/testify/require"
)

func TestRandomOrder(t *testing.T) {
	const qty = 50
	sales := 0
	for eid := 0; eid < qty; eid++ {
		t.Run(fmt.Sprint(eid), func(t *testing.T) {
			order := randomOrder(int64(eid))
			if order.IsSale() {
				sales++
			}

			require.True(t, order.Price >= 17)
			require.True(t, order.Price < 20)
			require.True(t, order.OriginalAmount >= 0)
			require.True(t, order.OriginalAmount < 1500)
		})
	}
	t.Run("more than half must be sales", func(t *testing.T) {
		require.True(t, sales > qty/2)
	})
	t.Run("not all may be sales", func(t *testing.T) {
		require.True(t, sales < qty)
	})
}

func TestRandomTrade(t *testing.T) {
	// make a bunch of orders
	orders := make(map[int64]bitmart.Order)
	for i := int64(1); i <= 100; i++ {
		orders[i] = randomOrder(i)
	}
	// make a bunch of trades
	trades := make(map[int64]bitmart.Trade)
	for i := int64(1); i <= 100; i++ {
		trade := randomTrade(orders, i)
		if trade == nil {
			break
		}
		trades[i] = *trade
	}
	t.Log("qty trades generated: ", len(trades))
	for eid, order := range orders {
		t.Run(fmt.Sprintf("order %d", eid), func(t *testing.T) {
			t.Run("no negative remaining balance", func(t *testing.T) {
				require.True(t, order.RemainingAmount >= 0)
			})
			t.Run("no remaining balance over initial balance", func(t *testing.T) {
				require.True(t, order.RemainingAmount <= order.OriginalAmount)
			})
			t.Run("status set correctly", func(t *testing.T) {
				if order.RemainingAmount <= 0 {
					require.Equal(t, bitmart.Success, bitmart.OrderStatus(order.Status))
				} else if order.RemainingAmount >= order.OriginalAmount {
					require.Equal(t, bitmart.Pending, bitmart.OrderStatus(order.Status))
				} else {
					require.Equal(t, bitmart.PartialSuccess, bitmart.OrderStatus(order.Status))
				}
			})
		})
	}
	t.Run("must generate at least 1 trade", func(t *testing.T) {
		require.NotEmpty(t, trades)
	})
	t.Run("must generate 100 or fewer trades", func(t *testing.T) {
		require.True(t, len(trades) <= 100)
	})
	t.Run("orders get modified", func(t *testing.T) {
		modified := false
		for _, order := range orders {
			if order.Status == int64(bitmart.PartialSuccess) || order.Status == int64(bitmart.Success) {
				modified = true
				break
			}
		}
		require.True(t, modified)
	})
	for oid, trade := range trades {
		t.Run(fmt.Sprint(oid), func(t *testing.T) {
			_, ok := orders[trade.EntrustID]
			require.True(t, ok, "order must be set successfully")

			t.Run("trade must have positive amount", func(t *testing.T) {
				require.True(t, trade.Amount > 0)
			})
		})
	}
}
