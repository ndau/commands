package main

import (
	"reflect"
	"testing"

	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
)

func Test_parseDollars(t *testing.T) {
	tests := []struct {
		in      string
		want    pricecurve.Nanocent
		wantErr bool
	}{
		{"// TODO: Add test cases.", 0, true},
		{"1", 100000000000, false},
		{"-1", -100000000000, false},
		{"$1", 100000000000, false},
		{"-$1", -100000000000, false},
		{"1.5", 0, true},
		{"-1.5", 0, true},
		{"$1.5", 0, true},
		{"-$1.5", 0, true},
		{"1.50", 150000000000, false},
		{"-1.50", -150000000000, false},
		{"$1.50", 150000000000, false},
		{"-$1.50", -150000000000, false},
		{"0.00000000001", 1, false},
		{"-0.00000000001", -1, false},
		{"$0.00000000001", 1, false},
		{"-$0.00000000001", -1, false},
		{"0.000000000001", 0, true},
		{"-0.000000000001", 0, true},
		{"$0.000000000001", 0, true},
		{"-$0.000000000001", 0, true},
		{"0.00_000_000_001", 1, false},
		{"-0.00_000_000_001", -1, false},
		{"$0.00_000_000_001", 1, false},
		{"-$0.00_000_000_001", -1, false},
		{"0.00,000,000,001", 1, false},
		{"-0.00,000,000,001", -1, false},
		{"$0.00,000,000,001", 1, false},
		{"-$0.00,000,000,001", -1, false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := parseDollars(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDollars() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDollars() = %v, want %v", got, tt.want)
			}
		})
	}
}
