package main

//go:generate gopherjs build --minify

// This is an experiment to see if gopherjs can reasonably generate js code from go source
// so that we can have a single-source solution for keys and addresses.
// Use "go generate" to build this.

import (
	"errors"

	"github.com/miratronix/jopher"

	"github.com/gopherjs/gopherjs/js"
)

func main() {
	js.Module.Get("exports").Set("newMagicNumber", newMagicNumber)
}

type MagicNumber struct {
	*js.Object
	Value     float64                         `js:"value"`
	Increment func(...interface{}) *js.Object `js:"increment"`
	Add       func(...interface{}) *js.Object `js:"add"`
}

func newMagicNumber(n float64) *js.Object {
	m := MagicNumber{Object: js.Global.Get("Object").New()}
	m.Value = n
	m.Increment = jopher.Promisify(m.increment)
	m.Add = jopher.Promisify(m.add)
	return jopher.Resolve(m.Object)
}

func (m *MagicNumber) increment() error {
	if m.Value+1 > 10 {
		return errors.New("there are no numbers bigger than 10")
	}
	m.Value++
	return nil
}

func (m *MagicNumber) add(x float64) (float64, error) {
	if m.Value+x > 10 {
		return m.Value, errors.New("there are no numbers bigger than 10")
	}
	m.Value += x
	return m.Value, nil
}
