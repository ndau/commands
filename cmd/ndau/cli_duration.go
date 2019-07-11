package main

import (
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
	duration := cmd.StringArg(Duration, "", "duration (ndaumath types.ParseDuration format)")

	return func() math.Duration {
		if duration == nil || *duration == "" {
			orQuit(errors.New("duration not set"))
		}
		tdur, err := math.ParseDuration(*duration)
		orQuit(errors.Wrap(err, "parsing duration"))
		return tdur
	}
}
