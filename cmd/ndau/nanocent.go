package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	"github.com/pkg/errors"
)

var dollarsRE *regexp.Regexp

func init() {
	dollarsRE = regexp.MustCompile(`^(?P<neg>-?)\$?(?P<dollars>[\d,_]+)(\.(?P<cents>\d{2,11}))?$`)
}

func parseDollars(dollars string) (pricecurve.Nanocent, error) {
	dollars = strings.TrimSpace(dollars)
	// allow for separation by just eliminating spacing chars
	// there isn't a great way to do this within the regex itself
	dollars = strings.Replace(dollars, ",", "", -1)
	dollars = strings.Replace(dollars, "_", "", -1)

	// perform regex matching
	match := dollarsRE.FindStringSubmatch(dollars)
	if len(match) == 0 {
		return 0, fmt.Errorf("'%s' doesn't look like dollars", dollars)
	}

	// get submatches by name
	submatches := make(map[string]string)
	for i, name := range dollarsRE.SubexpNames() {
		if i != 0 && i < len(match) && name != "" {
			submatches[name] = match[i]
		}
	}

	// parse integers
	var err error
	d := int64(0)
	if submatches["dollars"] != "" {
		d, err = strconv.ParseInt(submatches["dollars"], 10, 64)
		if err != nil {
			return 0, errors.Wrap(err, "parsing dollars as integer: "+submatches["dollars"])
		}
	}
	c := int64(0)
	if submatches["cents"] != "" {
		c, err = strconv.ParseInt(submatches["cents"], 10, 64)
		if err != nil {
			return 0, errors.Wrap(err, "parsing cents as integer: "+submatches["cents"])
		}
		for i := 0; i < 11-len(submatches["cents"]); i++ {
			c *= 10
		}
	}

	// add it all up
	nc := pricecurve.Nanocent(100000000000*d + c)

	// handle negatives
	if submatches["neg"] == "-" {
		nc = -nc
	}

	return nc, err
}

func getNanocentSpec() string {
	return "(USD | --nanocents=<NANOCENTS>)"
}

func getNanocentClosure(cmd *cli.Cmd) func() pricecurve.Nanocent {
	var (
		usd       = cmd.StringArg("USD", "", "US Dollars")
		nanocents = cmd.IntOpt("nanocents", 0, "Integer quantity of nanocents. Allows sub-cent precision.")
	)

	return func() pricecurve.Nanocent {
		if nanocents != nil && *nanocents != 0 {
			return pricecurve.Nanocent(*nanocents)
		}
		if usd != nil && *usd != "" {
			nc, err := parseDollars(*usd)
			orQuit(errors.Wrap(err, "parsing usd"))
			return nc
		}
		orQuit(errors.New("usd and nanocent not set"))
		return pricecurve.Nanocent(0)
	}
}
