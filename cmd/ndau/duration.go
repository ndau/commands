package main

import (
	"time"

	cli "github.com/jawher/mow.cli"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// Duration is the string of the Duration argument
const Duration = "DURATION"

func getDurationSpec() string {
	return Duration
}

func getDurationClosure(cmd *cli.Cmd) func() math.Duration {
	duration := cmd.StringArg(Duration, "", "duration (go time.ParseDuration format)")

	return func() math.Duration {
		if duration == nil || *duration == "" {
			orQuit(errors.New("duration not set"))
		}
		tdur, err := time.ParseDuration(*duration)
		orQuit(errors.Wrap(err, "parsing duration"))
		return math.DurationFrom(tdur)
	}
}
