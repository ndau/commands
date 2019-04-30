package main

import (
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_parseBool(t *testing.T) {
	type args struct {
		v   interface{}
		def bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"'true'", args{"true", false}, true},
		{"true", args{true, false}, true},
		{"false", args{false, true}, false},
		{"'yes'", args{"yes", false}, true},
		{"'maybe not'", args{"maybe", false}, false},
		{"'maybe so'", args{"maybe", true}, true},
		{"nil", args{nil, true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseBool(tt.args.v, tt.args.def); got != tt.want {
				t.Errorf("parseBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseSample(t *testing.T) {
	// go test suite always sets CWD to the dir containing the test file
	conf, err := Load("sample.toml", false)
	require.NoError(t, err)

	logger := conf.BuildLogger()
	tasks, err := conf.BuildTasks(logger)
	require.NoError(t, err)
	require.NotNil(t, tasks)

	// ensure that we have parsed the exit signal map properly
	require.NotNil(t, tasks.All)
	mi := tasks.All["MAYBE_INT"]
	require.NotNil(t, mi)
	require.NotNil(t, mi.ExitSignals)
	require.NotContains(t, mi.ExitSignals, 0)
	require.Contains(t, mi.ExitSignals, 1)
	require.Equal(t, mi.ExitSignals[1], syscall.SIGHUP)

	mi = tasks.All["MAYBE_INT_2"]
	require.NotNil(t, mi)
	require.NotNil(t, mi.ExitSignals)
	require.Equal(t, mi.ExitSignals[0xfe], syscall.SIGHUP)
	require.Equal(t, mi.ExitSignals[0xff], syscall.SIGINT)
	require.Equal(t, mi.ExitSignals[0777], syscall.SIGTERM)

	i := tasks.All["INT"]
	require.NotNil(t, i)
	require.Nil(t, i.ExitSignals)
}
