package main

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
