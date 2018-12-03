package main

import (
	"fmt"
	"strings"

	cli "github.com/jawher/mow.cli"
	"github.com/pkg/errors"
)

func getAccountNew(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "NAME [--hd]"

		var (
			name = cmd.StringArg("NAME", "", "Name to associate with the new identity")
			hd   = cmd.BoolOpt("hd", false, "when set, generate a HD key for this account")
		)

		cmd.Action = func() {
			config := getConfig()
			err := config.CreateAccount(*name, *hd)
			orQuit(errors.Wrap(err, "Failed to create identity"))
			err = config.Save()
			orQuit(errors.Wrap(err, "saving config"))
		}
	}
}

func getAccountRecover(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "NAME PHRASE... [--lang=<lang code>]"

		var (
			name   = cmd.StringArg("NAME", "", "Name to associate with the identity")
			phrase = cmd.StringsArg("PHRASE", []string{}, "recovery phrase for this identity")
			lang   = cmd.StringOpt("l lang", "en", "recovery phrase language")
		)

		cmd.Action = func() {
			if name == nil || phrase == nil || lang == nil {
				orQuit(errors.New("nil input--should be impossible"))
			}

			for idx := range *phrase {
				(*phrase)[idx] = strings.ToLower((*phrase)[idx])
			}

			config := getConfig()
			err := config.RecoverAccount(*name, *phrase, *lang)
			orQuit(errors.Wrap(err, "failed to recover identity"))
			err = config.Save()
			orQuit(errors.Wrap(err, "saving config"))
			if verbose != nil && *verbose {
				fmt.Println("OK")
			}
		}
	}
}
