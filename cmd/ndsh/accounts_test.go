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
	"encoding/hex"
	"sort"
	"testing"

	"github.com/ndau/ndaumath/pkg/address"
	"github.com/stretchr/testify/require"
)

func makeacct(t *testing.T) *Account {
	a, err := NewAccount(nil, "", 0)
	require.NoError(t, err)
	return &a
}

func makeaccts(t *testing.T, qty int) []*Account {
	as := make([]*Account, qty)
	for idx := range as {
		as[idx] = makeacct(t)
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
		a         *Account
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
				require.Equal(t, tt.args.a, ptr, "all added data pointers must point to the added item")
			}
		})
	}
}

func TestAccounts_AddIsIdempotent(t *testing.T) {
	a := makeacct(t)
	as := NewAccounts()

	getfirst := func(m map[*Account][]string) []string {
		for _, ss := range m {
			return ss
		}
		t.Fatal("no elements in m")
		return nil
	}

	logaddrs := func() {
		t.Logf("address of a: %p", &a)
		for idx := range as.rnames {
			t.Logf("              %p (%s)", as.accts[idx], rev(as.rnames[idx]))
		}
	}

	logaddrs()
	as.Add(a)
	logaddrs()
	require.Equal(t, 1, len(as.Reverse()))
	as.Add(a)
	logaddrs()
	require.Equal(t, 1, len(as.Reverse()))
	require.Equal(t, 0, len(getfirst(as.Reverse())))
	as.Add(a, "nick")
	logaddrs()
	require.Equal(t, 1, len(as.Reverse()))
	require.Equal(t, 1, len(getfirst(as.Reverse())))
	as.Add(a, "another", "nickname")
	logaddrs()
	require.Equal(t, 1, len(as.Reverse()))
	require.Equal(t, 3, len(getfirst(as.Reverse())))
}

func TestAccounts_Get(t *testing.T) {
	tests := []struct {
		name       string
		want       string
		existnames []string
		wantErr    func(error) bool
	}{
		{"nonempty list: not found in tail", "foo", []string{"bar"}, IsNoMatch},
		{"nonempty list: not found in head", "apple", []string{"bar"}, IsNoMatch},
		{"nonempty list: not found in middle", "bravo", []string{"alpha", "charlie"}, IsNoMatch},
		{"success: head", "alpha", []string{"alpha", "bravo", "charlie"}, nil},
		{"success: mid", "bravo", []string{"alpha", "bravo", "charlie"}, nil},
		{"success: tail", "charlie", []string{"alpha", "bravo", "charlie"}, nil},
		{"success: unique suffix", "o", []string{"alpha", "bravo", "charlie"}, nil},
		{"success: full word despite prefixes", "bravo", []string{"alpha", "sbravo", "bravo", "charlie"}, nil},
		{"not unique", "ravo", []string{"alpha", "bravo", "sbravo", "charlie"}, IsNotUniqueSuffix},
		{"empty search with 0 items", "", nil, func(error) bool { return true }}, // don't care what kind of error
		{"empty search with >1 item", "", []string{"foo", "bar"}, IsNotUniqueSuffix},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := NewAccounts()
			if len(tt.existnames) > 0 {
				as.Add(makeacct(t), tt.existnames...)
			}
			_, err := as.Get(tt.want)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.True(t, tt.wantErr(err), "wrong error type returned")
			} else {
				require.NoError(t, err)
			}
			// we're not testing what value we got, just because that's
			// difficult to test, and most of the hard work is in Add anyway
		})
	}
	t.Run("search in empty list", func(t *testing.T) {
		as := NewAccounts()
		_, err := as.Get("foo")
		require.Error(t, err)
		require.True(t, IsNoMatch(err), "wrong error type returned")
	})
	t.Run("empty search with 1 item", func(t *testing.T) {
		as := NewAccounts()
		as.Add(makeacct(t))
		_, err := as.Get("")
		require.NoError(t, err)
	})
}

func TestNewAccount(t *testing.T) {
	seed, err := hex.DecodeString("512a2544a89512a2544a89512a2544a8")
	require.NoError(t, err)
	paths := []string{
		"/44'/20036'/100/0",
		"/44'/20036'/100/1",
		"/44'/20036'/100/2",
	}
	kind := address.KindUser

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			a, err := NewAccount(seed, path, kind)
			require.NoError(t, err)
			require.NotZero(t, a)
		})
	}
}
