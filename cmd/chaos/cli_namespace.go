package main

import (
	"encoding/base64"
	"fmt"

	cli "github.com/jawher/mow.cli"
	cns "github.com/oneiro-ndev/chaos/pkg/chaos/ns"
	"github.com/oneiro-ndev/chaos/pkg/tool"
	"github.com/pkg/errors"
)

func getNamespaceSpec() string {
	return "(--ns=<NS> | NAME)"
}

func getNamespaceClosure(cmd *cli.Cmd) func(*tool.Config) (namespace []byte) {
	var (
		name = cmd.StringArg("NAME", "", "Use the namespace associated with this identity")
		ns   = cmd.StringOpt("ns", "", "Use the given namespace literal. Must be base64-encoded with trailing `=` padding stripped.")
	)

	return func(c *tool.Config) (namespace []byte) {
		if len(*name) > 0 {
			ns, err := c.NamespaceFor(*name)
			orQuit(errors.Wrap(err, "getting namespace"))
			return ns
		}

		if len(*ns) > 0 {
			// it can be hard to write a base64 literal to the command line;
			// the mow.cli library chokes on values which contain an `=` char.
			// The solution is to re-insert any necessary padding ourselves.
			missingPadding := (4 - (len(*ns) % 4)) % 4
			for i := 0; i < missingPadding; i++ {
				*ns = *ns + "="
			}

			nsB, err := base64.StdEncoding.DecodeString(*ns)
			orQuit(err)

			if len(nsB) != cns.Size {
				orQuit(fmt.Errorf(
					"namespace must have size %d; found %d",
					cns.Size, len(nsB),
				))
			}

			return nsB
		}

		orQuit(fmt.Errorf("Programming error in getNamespaceClosure: unreachable reached"))
		return []byte{}
	}
}
