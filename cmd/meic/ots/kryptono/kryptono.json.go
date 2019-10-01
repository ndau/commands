package kryptono

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
)

// bitmart has unusual ideas about what constitutes a number, sometimes

func getInt(obj map[string]interface{}, name string) (i int64, err error) {
	field, ok := obj[name]
	if !ok {
		err = fmt.Errorf("wallet field %s not found", name)
		return
	}
	switch fd := field.(type) {
	case int64:
		i = fd
		return
	case float64:
		i = int64(fd)
		return
	case string:
		i, err = strconv.ParseInt(fd, 0, 64)
		return
	}
	err = fmt.Errorf("unexpected type for Wallet.%s: %T", name, field)
	return
}

func getHex(obj map[string]interface{}, name string) (i int64, err error) {
	field, ok := obj[name]
	if !ok {
		err = fmt.Errorf("wallet field %s not found", name)
		return
	}
	switch fd := field.(type) {
	case int64:
		i = fd
		return
	case float64:
		i = int64(fd)
		return
	case string:
		i, err = strconv.ParseInt(fd, 16, 64)
		return
	}
	err = fmt.Errorf("unexpected type for Wallet.%s: %T", name, field)
	return
}

func getFloat(obj map[string]interface{}, name string) (f float64, err error) {
	field, ok := obj[name]
	if !ok {
		err = fmt.Errorf("wallet field %s not found", name)
		return
	}
	switch fd := field.(type) {
	case float64:
		f = fd
		return
	case string:
		f, err = strconv.ParseFloat(fd, 64)
		return
	}
	err = fmt.Errorf("unexpected type for Wallet.%s: %T", name, field)
	return
}

func getNdau(obj map[string]interface{}, name string) (n math.Ndau, err error) {
	field, ok := obj[name]
	if !ok {
		err = fmt.Errorf("wallet field %s not found", name)
		return
	}
	// we just have to approximate to the best of our ability
	switch fd := field.(type) {
	case float64:
		n = math.Ndau(fd * constants.NapuPerNdau)
		return
	case string:
		var f float64
		f, err = strconv.ParseFloat(fd, 64)
		n = math.Ndau(f * constants.NapuPerNdau)
		return
	}
	err = fmt.Errorf("unexpected type for Wallet.%s: %T", name, field)
	return
}

func getStr(obj map[string]interface{}, name string) (s string, err error) {
	field, ok := obj[name]
	if !ok {
		err = fmt.Errorf("wallet field %s not found", name)
		return
	}
	switch fd := field.(type) {
	case string:
		s = fd
		return
	}
	err = fmt.Errorf("unexpected type for Wallet.%s: %T", name, field)
	return
}

func getBool(obj map[string]interface{}, name string) (b bool, err error) {
	field, ok := obj[name]
	if !ok {
		err = fmt.Errorf("wallet field %s not found", name)
		return
	}
	switch fd := field.(type) {
	case bool:
		b = fd
	case string:
		l := strings.ToLower(fd)
		b = len(l) > 0 && (l[0] == 't' || l[0] == 'y')
		return
	case float64:
		b = fd != 0
		return
	}
	err = fmt.Errorf("unexpected type for Wallet.%s: %T", name, field)
	return
}
