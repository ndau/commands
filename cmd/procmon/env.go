package main

import (
	"fmt"
	"reflect"
	"regexp"
)

// envmap takes an environment map, copies it, and updates the copy by applying
// the input environment to it
func envmap(env []string, start map[string]string) map[string]string {
	em := make(map[string]string)
	for k, v := range start {
		em[k] = v
	}
	for _, e := range env {
		for i := 0; i < len(e); i++ {
			if e[i] == '=' {
				em[e[:i]] = e[i+1:]
			}
		}
	}
	return em
}

func interpolate(s string, em map[string]string) string {
	p := regexp.MustCompile(`\${?[a-zA-Z0-9_]+}?`)
	replacer := func(m string) string {
		// The first char is $ or we wouldn't be here
		lt := 1
		rt := len(m) - 1
		if m[lt] == '{' {
			lt++
		}
		if m[rt] == '}' {
			rt--
		}
		key := m[lt : rt+1]
		if v, ok := em[key]; ok {
			return v
		}
		return m
	}
	return p.ReplaceAllStringFunc(s, replacer)
}

func bb(i interface{}) {
	k := reflect.TypeOf(i).Kind()
	if k == reflect.Ptr {
		i = reflect.ValueOf(i).Elem().Interface()
	}
}

// func interpolateAll(data interface{}, em map[string]string) interface{} {
// 	k := reflect.TypeOf(data).Kind()
// 	switch k {
// 	case reflect.Map:
// 		m := reflect.ValueOf(data).Elem().Interface().(map[string]string)
// 		return interpolate(d, em)
// 	case
// 	}
// }

func interpolateAll(data interface{}, em map[string]string) interface{} {
	switch d := data.(type) {
	case string:
		return interpolate(d, em)
	case map[string]string:
		r := make(map[string]string)
		for k, v := range d {
			r[k] = interpolate(v, em)
		}
		return r
	case []string:
		r := make([]string, len(d))
		for i, v := range d {
			r[i] = interpolate(v, em)
		}
		return r
	default:
		panic(fmt.Sprintf("bad type: '%v' (%T)", d, d))
	}
}
