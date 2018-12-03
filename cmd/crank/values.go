package main

import (
	"errors"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/oneiro-ndev/chaincode/pkg/chain"
	"github.com/oneiro-ndev/chaincode/pkg/vm"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
)

// getRandomAccount randomly generates an account object
// it probably needs to be smarter than this
func getRandomAccount() backing.AccountData {
	const ticksPerDay = 24 * 60 * 60 * 1000000
	t, _ := types.TimestampFrom(time.Now())
	ad := backing.NewAccountData(t)
	// give it a balance between .1 and 100 ndau
	ad.Balance = types.Ndau((rand.Intn(1000) + 1) * 1000000)
	// set WAA to some time within 45 days
	ad.WeightedAverageAge = types.Duration(rand.Intn(ticksPerDay * 45))

	ad.LastEAIUpdate = t.Add(types.Duration(-rand.Intn(ticksPerDay * 3)))
	ad.LastWAAUpdate = t.Add(types.Duration(-rand.Intn(ticksPerDay * 10)))
	return ad
}

func parseValues(s string) ([]vm.Value, error) {
	// timestamp
	tsp := regexp.MustCompile("^[0-9-]+T[0-9:]+Z")
	// address is 48 chars starting with nd and not containing io10
	// addrp := regexp.MustCompile("^nd[2-9a-km-np-zA-KM-NP-Z]{46}")
	// number is a base-10 signed integer OR a hex value starting with 0x
	nump := regexp.MustCompile("^0x([0-9A-Fa-f]+)|^-?[0-9]+")
	// hexp := regexp.MustCompile("^0x([0-9A-Fa-f]+)")
	// quoted strings can be either single, double, or triple quotes of either kind; no escapes.
	quotep := regexp.MustCompile(`^'([^']*)'|^"([^"]*)"|^'''(.*)'''|^"""(.*)"""`)
	// arrays of bytes are B(hex) with individual bytes as hex strings with no 0x; embedded spaces are ignored
	bytep := regexp.MustCompile(`^B\((([0-9A-Fa-f][0-9A-Fa-f] *)+)\)`)
	// fields for structs are fieldid:Value; they are returned as a struct with one field that
	// is consolidated when they are enclosed in {} wrappers
	strfieldp := regexp.MustCompile("^([0-9]+) *:")

	s = strings.TrimSpace(s)
	retval := make([]vm.Value, 0)
	for s != "" {
		switch {
		case tsp.FindString(strings.ToUpper(s)) != "":
			t, err := vm.ParseTimestamp(tsp.FindString(strings.ToUpper(s)))
			if err != nil {
				panic(err)
			}
			s = s[len(tsp.FindString(strings.ToUpper(s))):]
			retval = append(retval, t)

		case strings.HasPrefix(s, "account"):
			str, err := chain.ToValue(getRandomAccount())
			s = s[len("account"):]
			if err != nil {
				return retval, err
			}
			retval = append(retval, str)

		case strings.HasPrefix(s, "["):
			if !strings.HasSuffix(s, "]") {
				return retval, errors.New("list start with no list end")
			}
			contents, err := parseValues(s[1 : len(s)-1])
			if err != nil {
				return retval, err
			}
			retval = append(retval, vm.NewList(contents...))
			// there can be only one list per line and it must end the line
			return retval, nil

		case strings.HasPrefix(s, "{"):
			if !strings.HasSuffix(s, "}") {
				return nil, errors.New("struct start with no struct end")
			}
			contents, err := parseValues(s[1 : len(s)-1])
			if err != nil {
				return nil, err
			}
			str := vm.NewStruct()
			for _, v := range contents {
				vs, ok := v.(*vm.Struct)
				if !ok {
					return retval, errors.New("untagged field in struct definition")
				}
				for _, ix := range vs.Indices() {
					v2, _ := vs.Get(ix)
					str = str.Set(ix, v2)
				}
			}
			return []vm.Value{str}, nil

		case strfieldp.FindString(s) != "":
			subm := strfieldp.FindStringSubmatch(s)
			f := subm[1]
			fieldid, _ := strconv.ParseInt(f, 10, 8)
			s = s[len(subm[0]):]
			contents, err := parseValues(s)
			if err != nil {
				return retval, err
			}
			if len(contents) == 0 {
				return retval, errors.New("field index without field value")
			}
			str := vm.NewStruct().Set(byte(fieldid), contents[0])
			retval = append(append(retval, str), contents[1:]...)
			return retval, nil

		case nump.FindString(s) != "":
			found := nump.FindString(s)
			s = s[len(found):]
			n, _ := strconv.ParseInt(found, 0, 64)
			retval = append(retval, vm.NewNumber(n))

		case bytep.FindString(s) != "":
			ba := []byte{}
			// the stream of bytes is the first submatch here
			submatches := bytep.FindStringSubmatch(s)
			contents := submatches[1]
			s = s[len(submatches[0]):]
			pair := regexp.MustCompile("([0-9A-Fa-f][0-9A-Fa-f])")
			for _, it := range pair.FindAllString(contents, -1) {
				b, _ := strconv.ParseInt(strings.TrimSpace(it), 16, 8)
				ba = append(ba, byte(b))
			}
			retval = append(retval, vm.NewBytes(ba))

		case quotep.FindString(s) != "":
			subm := quotep.FindSubmatch([]byte(s))
			contents := subm[1]
			s = s[len(subm[0]):]
			retval = append(retval, vm.NewBytes(contents))

		default:
			return nil, errors.New("unparseable " + s)
		}
		s = strings.TrimSpace(s)
	}
	return retval, nil
}
