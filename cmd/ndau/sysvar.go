package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	config "github.com/oneiro-ndev/ndau/pkg/tool.config"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
)

func getSysvar(verbose *bool, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Command(
			"get",
			"get system variables",
			getSysvarGet(verbose),
		)

		cmd.Command(
			"set",
			"set system variables",
			getSysvarSet(verbose, keys, emitJSON, compact),
		)
	}
}

func getSysvarGet(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		vars := cmd.StringsArg("VAR", nil, "Retrieve only these system variables. If unset, retrieve all system variables.")
		cmd.Spec = "[VAR...]"

		cmd.Action = func() {
			if verbose != nil && *verbose {
				if vars == nil || len(*vars) == 0 {
					fmt.Println("fetching all")
				} else {
					fmt.Println("fetching ", vars)
				}
			}

			conf := getConfig()
			svs, resp, err := tool.Sysvars(tmnode(conf.Node, nil, nil), *vars...)
			orQuit(errors.Wrap(err, "retrieving sysvars from blockchain"))

			// convert the returned sysvars into json, and re-encode into a new map
			jsvs := make(map[string]interface{})
			for name, sv := range svs {
				var buf bytes.Buffer
				_, err = msgp.UnmarshalAsJSON(&buf, sv)
				if err != nil {
					orQuit(errors.Wrap(err, "unmarshaling "+name))
				}
				var val interface{}
				bbytes := buf.Bytes()
				if len(bbytes) == 0 {
					jsvs[name] = ""
					continue
				}
				err = json.Unmarshal(bbytes, &val)
				if err != nil {
					orQuit(errors.Wrap(err, fmt.Sprintf("converting %s to json", name)))
				}
				jsvs[name] = val
			}

			// pretty-print json map
			jsout, err := json.MarshalIndent(jsvs, "", "  ")
			orQuit(errors.Wrap(err, "marshaling json"))
			fmt.Println(string(jsout))

			finish(*verbose, resp, err, "sysvar get")
		}
	}
}

func getSysvarSet(verbose *bool, keys *int, emitJSON, compact *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"NAME %s",
			getInputSpec("VALUE", true),
		)

		var (
			name     = cmd.StringArg("NAME", "", "name of sysvar to set")
			getValue = getInputClosure(cmd, "VALUE", true, verbose)
		)

		cmd.Action = func() {
			conf := getConfig()

			if conf.SetSysvar == nil {
				orQuit(errors.New("SetSysvar not configured"))
			}

			ssv := ndau.NewSetSysvar(
				*name,
				getValue(),
				sequence(conf, conf.SetSysvar.Address),
				config.FilterK(conf.SetSysvar.Keys, *keys)...,
			)

			result, err := tool.SendCommit(tmnode(conf.Node, emitJSON, compact), ssv)
			finish(*verbose, result, err, "sysvar set")
		}
	}
}
