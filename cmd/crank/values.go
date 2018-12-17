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
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
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

var predefined = predefinedConstants()

func parseValues(s string) ([]vm.Value, error) {
	// timestamp
	tsp := regexp.MustCompile("^[0-9-]+T[0-9:]+Z")
	// number is a base-10 signed integer OR a hex value starting with 0x
	nump := regexp.MustCompile("^0x([0-9A-Fa-f]+)|^-?[0-9]+")
	// napu is a base-10 positive integer preceded with np; it is delivered as an integer number of napu
	napup := regexp.MustCompile("^np([0-9]+)")
	// ndau values are a base-10 positive decimal, which is multiplied by 10^8 and converted to integer
	ndaup := regexp.MustCompile(`^nd([0-9]+)(\.[0-9]*)?`)
	// quoted strings can be either single, double, or triple quotes of either kind; no escapes.
	quotep := regexp.MustCompile(`^'([^']*)'|^"([^"]*)"|^'''(.*)'''|^"""(.*)"""`)
	// arrays of bytes are B(hex) with individual bytes as hex strings with no 0x; embedded spaces are ignored
	bytep := regexp.MustCompile(`^B\((([0-9A-Fa-f][0-9A-Fa-f] *)+)\)`)
	// fields for structs are fieldid:Value; they are returned as a struct with one field that
	// is consolidated when they are enclosed in {} wrappers
	strfieldp := regexp.MustCompile("^([0-9]+|[A-Z_]+) *:")

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
			// see if it's a predefined constant
			// if so, use its value instead
			if v, ok := predefined[f]; ok {
				f = v
			}
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

		case napup.FindString(s) != "":
			found := napup.FindString(s)
			s = s[len(found):]
			n, _ := strconv.ParseInt(found[2:], 0, 64)
			retval = append(retval, vm.NewNumber(n))

		case ndaup.FindString(s) != "":
			found := ndaup.FindString(s)
			s = s[len(found):]
			n, _ := strconv.ParseFloat(found[2:], 64)
			retval = append(retval, vm.NewNumber(int64(n*constants.QuantaPerUnit)))

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
