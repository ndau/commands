package main

import "testing"

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
