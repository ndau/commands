package main

import (
	"sort"
	"testing"

	"github.com/oneiro-ndev/ndaumath/pkg/signature"

	"github.com/stretchr/testify/require"
)

func makeacct(t *testing.T) Account {
	a, err := NewAccount(signature.Ed25519, nil, 0)
	require.NoError(t, err)
	return a
}

func makeaccts(t *testing.T, qty int) []*Account {
	as := make([]*Account, qty)
	for idx := range as {
		a := makeacct(t)
		as[idx] = &a
	}
	return as
}

func TestAccounts_Add(t *testing.T) {
	nnd := makeaccts(t, 2)
	type fields struct {
		rnames []string
		accts  []*Account
	}
	type args struct {
		a         Account
		nicknames []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			"add without nicknames to empty accounts",
			fields{make([]string, 0), make([]*Account, 0)},
			args{makeacct(t), nil},
		},
		{
			"add with nicknames to empty accounts",
			fields{make([]string, 0), make([]*Account, 0)},
			args{makeacct(t), []string{"foo", "bar"}},
		},
		{
			"add without nicknames to non-empty accounts",
			fields{[]string{"bar", "foo"}, []*Account{&Account{}, &Account{}}},
			args{makeacct(t), nil},
		},
		{
			"add with nicknames to non-empty accounts",
			fields{[]string{"bar", "foo"}, []*Account{&Account{}, &Account{}}},
			args{makeacct(t), []string{"zip", "bat", "baz"}},
		},
		{
			"add without nicknames to non-empty accounts with nicknames",
			fields{
				[]string{"alpha", "beta", "charlie", "delta"},
				[]*Account{nnd[1], nnd[0], nnd[0], nnd[1]},
			},
			args{makeacct(t), nil},
		},
		{
			"add with nicknames to non-empty accounts with nicknames",
			fields{
				[]string{"alpha", "beta", "charlie", "delta"},
				[]*Account{nnd[0], nnd[0], nnd[0], nnd[1]},
			},
			args{makeacct(t), []string{"zip", "bat", "baz"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := &Accounts{
				rnames: tt.fields.rnames,
				accts:  tt.fields.accts,
			}
			require.Equal(t, len(as.rnames), len(as.accts), "data lengths must start equal")

			// keep track of some data for later tests
			oldlen := len(as.rnames)
			oldamap := make(map[string]*Account)
			for idx := range as.rnames {
				oldamap[as.rnames[idx]] = as.accts[idx]
			}

			t.Log("old rnames:", as.rnames)

			as.Add(tt.args.a, tt.args.nicknames...)

			t.Log("new rnames:", as.rnames)

			require.Equal(t, len(as.rnames), len(as.accts), "data lengths must remain equal")
			require.Equal(t, oldlen+len(tt.args.nicknames)+1, len(as.rnames), "must insert all and only expected data")
			require.True(t, sort.StringsAreSorted(as.rnames), "rnames must always remain sorted")

			// ensure that we haven't overwritten any data
			newamap := make(map[string]*Account)
			for idx := range as.rnames {
				newamap[as.rnames[idx]] = as.accts[idx]
			}

			// start by filtering the old data from the new data
			olddata := make(map[string]*Account)
			newdata := make(map[string]*Account)
			for name, data := range newamap {
				if _, ok := oldamap[name]; ok {
					olddata[name] = data
				} else {
					newdata[name] = data
				}
			}

			require.Equal(t, oldamap, olddata, "we must not have overwritten any data")
			for _, ptr := range newdata {
				require.Equal(t, &tt.args.a, ptr, "all added data pointers must point to the added item")
			}
		})
	}
}
