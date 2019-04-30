package main

import (
	"os"
	"syscall"
	"testing"
	"time"

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

	logger := logrus.New()
	logger.Out = testwriter.New(t)

	const maxShutdown = 50 * time.Millisecond
	// 50 milliseconds to run true or false should be way more than enough time:
	//
	// $ hyperfine -iw 10 true false
	// Tue Apr 30 17:05:44 CEST 2019
	// Benchmark #1: true
	//   Time (mean ± σ):       0.1 ms ±   0.1 ms    [User: 0.0 ms, System: 0.0 ms]
	//   Range (min … max):     0.0 ms …   0.8 ms    1261 runs
	//
	//   Warning: Command took less than 5 ms to complete. Results might be inaccurate.
	//   Warning: The first benchmarking run for this command was significantly slower than the rest (0.1 ms). This could be caused by (filesystem) caches that were not filled until after the first run. You should consider using the '--warmup' option to fill those caches before the actual benchmark. Alternatively, use the '--prepare' option to clear the caches before each timing run.
	//
	// Benchmark #2: false
	//   Time (mean ± σ):       0.1 ms ±   0.1 ms    [User: 0.0 ms, System: 0.0 ms]
	//   Range (min … max):     0.0 ms …   0.7 ms    1317 runs
	//
	//   Warning: Command took less than 5 ms to complete. Results might be inaccurate.
	//   Warning: Ignoring non-zero exit code.
	//   Warning: Statistical outliers were detected. Consider re-running this benchmark on a quiet PC without any interferences from other programs. It might help to use the '--warmup' or '--prepare' options.
	//
	// Summary
	//   'false' ran
	//     1.26 ± 3.21 times faster than 'true'

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sigchan := make(chan os.Signal, 1)
			task := NewTask(tt.name, tt.fields.Path)
			task.Logger = logger
			task.MaxShutdown = maxShutdown

			task.Start(nil)
			go task.exitMonitor()

			// wait briefly to help resolve race conditions
			time.Sleep(maxShutdown)

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
		})
	}
}
