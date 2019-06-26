package main

import (
	"reflect"
	"testing"

	"github.com/oneiro-ndev/chaincode/pkg/vm"
)

func Test_parseValues(t *testing.T) {
	ts := "2018-07-18T20:00:00Z"
	pts, _ := vm.ParseTimestamp(ts)
	tests := []struct {
		name    string
		input   string
		want    []vm.Value
		wantErr bool
	}{
		{"number 1", "1", []vm.Value{vm.NewNumber(1)}, false},
		{"number 1000", "1000", []vm.Value{vm.NewNumber(1000)}, false},
		{"binary number 0b1000", "0b1000", []vm.Value{vm.NewNumber(8)}, false},
		{"octal number 0100", "0100", []vm.Value{vm.NewNumber(0100)}, false},
		{"octal number 0777", "0777", []vm.Value{vm.NewNumber(0777)}, false},
		{"hex number 0x100", "0x100", []vm.Value{vm.NewNumber(0x100)}, false},
		{"hex number 0x10q", "0x10q", nil, true},
		{"string 1", "'hiya'", []vm.Value{vm.NewBytes([]byte("hiya"))}, false},
		{"string 2", `"hiya"`, []vm.Value{vm.NewBytes([]byte("hiya"))}, false},
		// {"string 3", `"""hiya"""`, []vm.Value{vm.NewBytes([]byte("hiya"))}, false},
		// {"string 4", `'''hiya'''`, []vm.Value{vm.NewBytes([]byte("hiya"))}, false},
		{"simple list", "[ 1 ]", []vm.Value{vm.NewList(vm.NewNumber(1))}, false},
		{"multi-item list", "[1, 2]", []vm.Value{vm.NewList(vm.NewNumber(1), vm.NewNumber(2))}, false},
		{"multi-item list2", "[ 1 2 3 4]", []vm.Value{vm.NewList(vm.NewNumber(1), vm.NewNumber(2), vm.NewNumber(3), vm.NewNumber(4))}, false},
		{"multi-type list", "[1, 2, 'buckle my shoe']", []vm.Value{vm.NewList(vm.NewNumber(1), vm.NewNumber(2), vm.NewBytes([]byte("buckle my shoe")))}, false},
		{"nested list", "[1, 2, [3]]", []vm.Value{vm.NewList(vm.NewNumber(1), vm.NewNumber(2), vm.NewList(vm.NewNumber(3)))}, false},
		{"timestamp", ts, []vm.Value{pts}, false},
		{"simple struct", "{ 1:2 }", []vm.Value{vm.NewStruct().Set(1, vm.NewNumber(2))}, false},
		{"multi struct", "{ 1:2, 3:'hello' }", []vm.Value{vm.NewStruct().Set(1, vm.NewNumber(2)).Set(3, vm.NewBytes([]byte("hello")))}, false},
		{"list of struct", "[{ 1:2 }]", []vm.Value{vm.NewList(vm.NewStruct().Set(1, vm.NewNumber(2)))}, false},
		{"struct of list", "{ 1:[2] }", []vm.Value{vm.NewStruct().Set(1, vm.NewList(vm.NewNumber(2)))}, false},
		{"struct of list2", "{ ACCT_VALIDATIONKEYS: [ 1 2] }", []vm.Value{vm.NewStruct().Set(62, vm.NewList(vm.NewNumber(1), vm.NewNumber(2)))}, false},
		{"struct bad key 1", "{ 1.2: 1 }", nil, true},
		{"struct bad key 2", "{ 'foo': 2 }", nil, true},
		{"struct bad key 3", "{ BAR: 3 }", nil, true},
		{"ndau", "nd2", []vm.Value{vm.NewNumber(200000000)}, false},
		{"napu", "np33", []vm.Value{vm.NewNumber(33)}, false},
		{"hex bytes", "B(4869)", []vm.Value{vm.NewBytes([]byte("Hi"))}, false},
		{"boolean truth", "tRUE", []vm.Value{vm.NewNumber(1)}, false},
		{"boolean falsity", "FAlsE", []vm.Value{vm.NewNumber(0)}, false},
		{"realworld", "[{121: False, 122: True}, {121: True, 122: False}]",
			[]vm.Value{vm.NewList(
				vm.NewStruct().Set(121, vm.NewNumber(0)).Set(122, vm.NewNumber(1)),
				vm.NewStruct().Set(121, vm.NewNumber(1)).Set(122, vm.NewNumber(0)))}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseValues(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseValues() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_parseValueAccount(t *testing.T) {
	// because the "account" command returns random data, we can't test it in the table-driven test
	got, err := parseValues("account")
	if err != nil {
		t.Errorf("parseValues returned error %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1, got %d", len(got))
	}
	str, ok := got[0].(*vm.Struct)
	if !ok {
		t.Error("account was not a vm.Struct")
	}
	if str.Len() < 18 {
		t.Error("account struct doesn't have enough fields.")
	}
}
