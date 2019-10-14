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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNetSet(t *testing.T) {
	client, err := getClient("http://example.org:4321", 0)
	require.NoError(t, err)
	sh := NewShell(true, client, Net{})
	err = sh.Exec("net --set http://fake.org:1234")
	require.NoError(t, err)
	require.Equal(t, "http://fake.org:1234", ClientURL.String())
}
