package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTaskset(t *testing.T) {
	ts := TaskSet{}
	t1 := &Task{}
	t2 := &Task{}
	t3 := &Task{}

	ts.Add(t1, t2)
	require.True(t, ts.Has(t1))
	require.True(t, ts.Has(t2))
	require.False(t, ts.Has(t3))
	ts.Add(t1)
	ts.Add(t2)
	require.True(t, ts.Has(t1))
	require.True(t, ts.Has(t2))
	require.False(t, ts.Has(t3))
	ts.Add(t3)
	require.True(t, ts.Has(t1))
	require.True(t, ts.Has(t2))
	require.True(t, ts.Has(t3))
	ts.Delete(t2)
	require.True(t, ts.Has(t1))
	require.False(t, ts.Has(t2))
	require.True(t, ts.Has(t3))
	ts.Delete(t2)
	require.True(t, ts.Has(t1))
	require.False(t, ts.Has(t2))
	require.True(t, ts.Has(t3))
	ts.Delete(t1)
	require.False(t, ts.Has(t1))
	require.False(t, ts.Has(t2))
	require.True(t, ts.Has(t3))
	ts.Delete(t3)
	require.False(t, ts.Has(t1))
	require.False(t, ts.Has(t2))
	require.False(t, ts.Has(t3))
}
