package main

import (
	"fmt"
	"sort"
	"strconv"
)

type opcodeInfo struct {
	Value   byte
	Name    string
	Synonym string
	Summary string
	Doc     string
	Errors  string
	Example example
	Parms   []parm
	Enabled bool
	NoAsm   bool
}

type example struct {
	Pre  string
	Inst string
	Post string
}

// getparm is a helper function for templates to index into the array
func getParm(o opcodeInfo, i int) parm {
	return o.Parms[i]
}

func nbytes(o opcodeInfo) string {
	switch len(o.Parms) {
	case 1:
		return o.Parms[0].Nbytes()
	case 2:
		a, _ := strconv.Atoi(o.Parms[0].Nbytes())
		b, _ := strconv.Atoi(o.Parms[0].Nbytes())
		return strconv.Itoa(a + b)
	default:
		return ""
	}
}

type parm interface {
	Nbytes() string
	PeggoParm() string
	PeggoTmpl() string
	Placeholder() string
}

type timeParm struct {
}

func (p timeParm) Nbytes() string {
	return "8"
}

func (p timeParm) PeggoParm() string {
	return fmt.Sprintf("Timestamp")
}

func (p timeParm) PeggoTmpl() string {
	return fmt.Sprintf("PushT")
}

func (p timeParm) Placeholder() string {
	return "t"
}

type pushbParm struct {
}

func (p pushbParm) Nbytes() string {
	return "int(getat(offset+1)) + 1"
}

func (p pushbParm) PeggoParm() string {
	return fmt.Sprintf("Bytes")
}

func (p pushbParm) PeggoTmpl() string {
	return fmt.Sprintf("PushB")
}

func (p pushbParm) Placeholder() string {
	return "ba"
}

type eventListParm struct {
}

func (p eventListParm) Nbytes() string {
	return "int(getat(offset+1)) + 1"
}

func (p eventListParm) PeggoParm() string {
	return fmt.Sprintf("Bytes")
}

func (p eventListParm) PeggoTmpl() string {
	return fmt.Sprintf("Handler")
}

func (p eventListParm) Placeholder() string {
	return "events"
}

type functionIDParm struct{}

func (p functionIDParm) Nbytes() string {
	return "1"
}

func (p functionIDParm) PeggoParm() string {
	return "FunctionName"
}

func (p functionIDParm) PeggoTmpl() string {
	return "CallOp"
}

func (p functionIDParm) Placeholder() string {
	return "id"
}

type indexParm struct {
	Name string
}

func (p indexParm) Nbytes() string {
	return "1"
}

func (p indexParm) PeggoParm() string {
	return "Value"
}

func (p indexParm) PeggoTmpl() string {
	return "BinOp"
}

func (p indexParm) Placeholder() string {
	return p.Name
}

// embeddedParm is intended only for
type embeddedParm struct {
	N string
}

func (p embeddedParm) Nbytes() string {
	return p.N
}

func (p embeddedParm) PeggoParm() string {
	return "Value"
}

func (p embeddedParm) PeggoTmpl() string {
	return ""
}

func (p embeddedParm) Placeholder() string {
	return ""
}

type opcodeInfos []opcodeInfo

type byValue opcodeInfos

func (a byValue) Len() int           { return len(a) }
func (a byValue) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byValue) Less(i, j int) bool { return a[i].Value < a[j].Value }

// byName sorts in reverse order (deliberately)
type byName opcodeInfos

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name > a[j].Name }

// subset returns the subset of the opcodeInfo that matches
// the state of the associated flags: enabled, withSynonym
func (o opcodeInfos) subset(enabled bool, withSynonym bool) opcodeInfos {
	o2 := make(opcodeInfos, 0)
	for i := range o {
		if o[i].Enabled == enabled {
			o2 = append(o2, o[i])
			if withSynonym && o[i].Synonym != "" {
				syn := o[i]
				syn.Name = syn.Synonym
				o2 = append(o2, syn)
			}
		}
	}
	sort.Sort(byValue(o2))
	return o2
}

func (o opcodeInfos) Enabled() opcodeInfos {
	return o.subset(true, false)
}

func (o opcodeInfos) EnabledWithSynonyms() opcodeInfos {
	return o.subset(true, true)
}

func (o opcodeInfos) Disabled() opcodeInfos {
	return o.subset(false, false)
}

func (o opcodeInfos) ChasmOpcodes() opcodeInfos {
	o2 := o.subset(true, true)
	o3 := make(opcodeInfos, 0)
	for i := range o2 {
		if !o2[i].NoAsm {
			o3 = append(o3, o2[i])
		}
	}
	sort.Sort(byName(o3))
	return o3
}
