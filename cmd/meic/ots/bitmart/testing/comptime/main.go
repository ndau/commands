package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	bitmart "github.com/oneiro-ndev/commands/cmd/meic/ots/bitmart"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// Your timestamp must be within 60 seconds of the API service time or
// your request will be considered expired and rejected. We recommend using
// the time endpoint to query for the API server time if you believe there
// may be time skew between your server and the API servers.

// return the bitmart time in milliseconds since unix epoch, UTC
func serverTime() (int64, error) {
	resp, err := http.Get(bitmart.APITime)
	if err != nil {
		return 0, errors.Wrap(err, "http.Get")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.Wrap(err, "read response body")
	}
	var jsdata map[string]interface{}
	err = json.Unmarshal(body, &jsdata)
	if err != nil {
		return 0, errors.Wrap(err, "unmarshal response json")
	}
	timeobj, ok := jsdata["server_time"]
	if !ok {
		return 0, errors.New("server time not present in response")
	}
	// it would sure be nice if go was a bit smarter about compound type switches
	switch t := timeobj.(type) {
	case int:
		return int64(t), nil
	case int8:
		return int64(t), nil
	case int16:
		return int64(t), nil
	case int32:
		return int64(t), nil
	case int64:
		return t, nil
	case uint:
		return int64(t), nil
	case uint8:
		return int64(t), nil
	case uint16:
		return int64(t), nil
	case uint32:
		return int64(t), nil
	case uint64:
		return int64(t), nil
	case string:
		tint, err := strconv.ParseInt(t, 10, 64)
		err = errors.Wrap(err, "parsing server time")
		return tint, err
	}
	return 0, fmt.Errorf("unknown server time type %T", timeobj)
}

// return the number of milliseconds since unix epoch in utc
func localTime() int64 {
	now := time.Now()
	return (now.UnixNano() / 1000000)
	//         nano to micro  ^^^
	//        micro to milli     ^^^
}

func print(label string, data, debug interface{}) {
	fmt.Printf("%15s: %18v (%v)\n", label, data, debug)
}

func millisAsTime(milliseconds int64) time.Time {
	seconds := milliseconds / 1000
	nanoseconds := (milliseconds % 1000) * 1000000
	return time.Unix(seconds, nanoseconds).In(time.UTC)
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	st, err := serverTime()
	check(err)
	print("server time", st, millisAsTime(st).String())

	lt := localTime()
	print("local time", lt, millisAsTime(lt).String())

	skew := time.Duration((lt - st) * 1000000)
	print("skew", skew.String(), math.DurationFrom(skew).String())

	for i := 0; i < 9; i++ {
		time.Sleep(6 * time.Second)
		st, err = serverTime()
		check(err)
		lt = localTime()
		skew = time.Duration((lt - st) * 1000000)
		print("skew", skew.String(), math.DurationFrom(skew).String())
	}
}
