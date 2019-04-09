package main

import "testing"

func Test_parseBool(t *testing.T) {
	type args struct {
		v   interface{}
		def bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"'true'", args{"true", false}, true},
		{"true", args{true, false}, true},
		{"false", args{false, true}, false},
		{"'yes'", args{"yes", false}, true},
		{"'maybe not'", args{"maybe", false}, false},
		{"'maybe so'", args{"maybe", true}, true},
		{"nil", args{nil, true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseBool(tt.args.v, tt.args.def); got != tt.want {
				t.Errorf("parseBool() = %v, want %v", got, tt.want)
			}
		})
	}
}
