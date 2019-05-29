package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/oneiro-ndev/json2msgp"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
)

// Sysvar changes an account's validation
type Sysvar struct{}

var _ Command = (*Sysvar)(nil)

// Name implements Command
func (Sysvar) Name() string { return "sysvar" }

type sysvarargs struct {
	Action    string   `arg:"positional,required" help:"get or set"`
	Names     []string `arg:"positional" help:"name of a sysvar to interact with"`
	Value     string   `arg:"-v" help:"value to set"`
	TypeHints string   `arg:"--type-hints" help:"must be map[string][]string"`
	Stage     bool     `arg:"-S" help:"stage this tx; do not send it"`
}

func (sysvarargs) Description() string {
	return strings.TrimSpace(`
Get or set system variables
	`)
}

// Run implements Command
func (Sysvar) Run(argvs []string, sh *Shell) (err error) {
	args := sysvarargs{}

	err = ParseInto(argvs, &args)
	if err != nil {
		if err == arg.ErrHelp || err == arg.ErrVersion {
			err = nil
		}
		return
	}

	switch args.Action {
	case "get":
		return sysvarGet(sh, args)
	case "set":
		return sysvarSet(sh, args)
	default:
		return errors.New("sysvar action must be 'get' or 'set'")
	}
}

func sysvarGet(sh *Shell, args sysvarargs) error {
	svs, _, err := tool.Sysvars(sh.Node, args.Names...)

	// convert the returned sysvars into json, and re-encode into a new map
	// (they arrive in msgp format, and this makes them human-readable)
	jsvs := make(map[string]interface{})
	for name, sv := range svs {
		var buf bytes.Buffer
		_, err = msgp.UnmarshalAsJSON(&buf, sv)
		if err != nil {
			return errors.Wrap(err, "unmarshaling "+name)
		}
		var val interface{}
		bbytes := buf.Bytes()
		if len(bbytes) == 0 {
			jsvs[name] = ""
			continue
		}
		err = json.Unmarshal(bbytes, &val)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("converting %s to json", name))
		}
		jsvs[name] = val
	}

	// pretty-print json map
	jsout, err := json.MarshalIndent(jsvs, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshaling json")
	}
	sh.Write(string(jsout))
	return nil
}

func sysvarSet(sh *Shell, args sysvarargs) error {
	if len(args.Names) != 1 {
		return errors.New("must specify exactly 1 sysvar to set")
	}

	value, err := parseJSON(args.Value)
	if err != nil {
		return errors.Wrap(err, "parsing value")
	}

	var hints map[string][]string
	if args.TypeHints != "" {
		hintsi, err := parseJSON(args.TypeHints)
		if err != nil {
			return errors.Wrap(err, "parsing type hints")
		}
		var ok bool
		hints, ok = hintsi.(map[string][]string)
		if !ok {
			return errors.New("type hints must be map of string to list of string")
		}
	}

	// now heuristically convert this json value to msgp
	msgpdata, err := json2msgp.Convert(value, hints)
	if err != nil {
		return errors.Wrap(err, "converting value to msgp")
	}
	sh.VWrite("msgpdata: %x", msgpdata)

	magic, err := sh.SAAcct(sv.SetSysvarAddressName, sv.SetSysvarValidationPrivateName)
	if err != nil {
		return errors.Wrap(err, "SetSysvar")
	}

	key, err := sh.SAPrivateKey(sv.SetSysvarValidationPrivateName)
	if err != nil {
		return errors.Wrap(err, "SetSysvar")
	}

	magic.Data.Sequence++
	ssv := ndau.NewSetSysvar(
		args.Names[0],
		msgpdata,
		magic.Data.Sequence,
		*key,
	)

	if args.Stage {
		sh.Staged = &Stage{
			Tx:      ssv,
			Account: magic,
		}
		return nil
	}

	_, err = tool.SendCommit(sh.Node, ssv)
	if err != nil {
		sh.Staged = &Stage{
			Tx:      ssv,
			Account: magic,
		}
	}
	return err
}
