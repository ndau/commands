package main

import "github.com/oneiro-ndev/chaincode/pkg/vm"

type nower struct{ now vm.Timestamp }

func (n nower) Now() (vm.Timestamp, error) {
	return n.now, nil
}

var _ vm.Nower = (*nower)(nil)
