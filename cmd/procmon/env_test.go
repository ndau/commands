package main

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"reflect"
	"testing"
)

func Test_interpolate(t *testing.T) {
	em := map[string]string{
		"CAT": "Kitty",
		"DOG": "Shiner",
	}
	type args struct {
		s  string
		em map[string]string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"none", args{"hello there", em}, "hello there"},
		{"plain", args{"hello $CAT", em}, "hello Kitty"},
		{"braces", args{"hello ${DOG}", em}, "hello Shiner"},
		{"multiple", args{"hello $CAT and ${DOG}", em}, "hello Kitty and Shiner"},
		{"repeat", args{"hello $DOG/${DOG}", em}, "hello Shiner/Shiner"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := interpolate(tt.args.s, tt.args.em); got != tt.want {
				t.Errorf("interpolate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_interpolateAll(t *testing.T) {
	em := map[string]string{
		"CAT": "Kitty",
		"DOG": "Shiner",
	}

	type args struct {
		data interface{}
		em   map[string]string
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{"string", args{"hello $CAT", em}, "hello Kitty"},
		{"[]string", args{[]string{
			"hello $CAT",
			"hello ${DOG}",
		}, em}, []string{
			"hello Kitty",
			"hello Shiner",
		}},
		{"map[string]string", args{map[string]string{
			"cat": "hello $CAT",
			"dog": "hello ${DOG}",
		}, em}, map[string]string{
			"cat": "hello Kitty",
			"dog": "hello Shiner",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := interpolateAll(tt.args.data, tt.args.em); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("interpolateAll() = %v, want %v", got, tt.want)
			}
		})
	}
}
