package baddress

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
)

// A BadAddress is one users should not use
type BadAddress struct {
	Address address.Address `arg:"positional,required" help:"an address users should never use"`
	Path    string          `arg:"-p" help:"derivation path for the bad address"`
	Reason  string          `arg:"-r" help:"why the user should not use this address"`
}

// Add a bad address to the DB
func Add(ddb *dynamodb.DynamoDB, addr BadAddress) error {
	fmt.Println("add: unimplemented")
	return nil
}

// Remove a bad address from teh DB
func Remove(ddb *dynamodb.DynamoDB, addr BadAddress) error {
	fmt.Println("remove: unimplemented")
	return nil
}
