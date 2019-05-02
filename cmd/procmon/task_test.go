package main

import (
	"os"
	"sync"
	"syscall"
	"testing"

	"github.com/oneiro-ndev/writers/pkg/testwriter"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestTask_exitMonitor(t *testing.T) {
	type fields struct {
		Path        string
		ExitSignals map[int]os.Signal
	}
	tests := []struct {
		name   string
		fields fields
		expect os.Signal
	}{
		{"true, no handler", fields{"/usr/bin/true", nil}, nil},
		{"true, handle 1", fields{"/usr/bin/true", map[int]os.Signal{1: syscall.SIGHUP}}, nil},
		{"true, handle 0", fields{"/usr/bin/true", map[int]os.Signal{0: syscall.SIGHUP}}, syscall.SIGHUP},
		{"false, handle 1", fields{"/usr/bin/false", map[int]os.Signal{1: syscall.SIGHUP}}, syscall.SIGHUP},
		{"false, handle 0", fields{"/usr/bin/false", map[int]os.Signal{0: syscall.SIGHUP}}, nil},
		{"true, handle 0 and 1", fields{"/usr/bin/true", map[int]os.Signal{0: syscall.SIGHUP, 1: syscall.SIGTERM}}, syscall.SIGHUP},
		{"false, handle 0 and 1", fields{"/usr/bin/false", map[int]os.Signal{0: syscall.SIGHUP, 1: syscall.SIGTERM}}, syscall.SIGTERM},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sigchan := make(chan os.Signal, 1)

			// create and configure the task under test
			task := NewTask(tt.name, tt.fields.Path)
			task.Logger = logrus.New()
			task.Logger.(*logrus.Logger).Out = testwriter.New(t)
			task.ExitSignals = tt.fields.ExitSignals
			task.skipStartMonitors = true
			task.Sigchan = sigchan

			// get it going
			task.Start(nil)

			// dockerized version of this test claims this channel is nil, which
			// causes the test to hang forever. If that's true, something has
			// gone wrong, and we can fail fast.
			require.NotNil(t, task.Status)

			// wait only as long as it takes for the exitMonitor function to complete
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				task.exitMonitor()
				wg.Done()
			}()
			wg.Wait()

			if tt.expect == nil {
				// no data should be available on sigchan
				usedDefault := false
				select {
				case <-sigchan:
					t.Fatal("expected no data on sigchan")
				default:
					usedDefault = true
				}
				require.True(t, usedDefault, "expected no data on sigchan")
			} else {
				select {
				case got, ok := <-sigchan:
					require.True(t, ok)
					require.Equal(t, tt.expect, got)
				default:
					t.Fatal("expected data on sigchan")
				}
			}

			// in all cases, ensure that the Status channel received a Stop
			// which contains information about the task's exit code
			select {
			case status := <-task.Status:
				require.Equal(t, Stop, status.Code())
				require.IsType(t, TerminateEvent{}, status)
			default:
				t.Fatal("expected data on task.Status")
			}
		})
	}
}
