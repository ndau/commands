package baddress

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/pkg/errors"
)

// Table is the table name for the bad address dynamodb table
const Table = "bad-addresses"

// Region is the AWS region to use
const Region = "us-east-1"

const (
	addressField = "address"
	pathField    = "derivation-path"
)

// A BadAddress is one users should not use
type BadAddress struct {
	Address address.Address `json:"address" arg:"positional,required" help:"an address users should never use"`
	Path    string          `json:"derivation-path" arg:"-p" help:"derivation path for the bad address"`
	Reason  string          `json:"reason" arg:"-r" help:"why the user should not use this address"`
}

func (b BadAddress) marshal() (map[string]*dynamodb.AttributeValue, error) {
	// ddb attrs can't be empty strings
	if b.Path == "" {
		b.Path = "/"
	}

	m, err := dynamodbattribute.MarshalMap(b)
	if err != nil {
		return nil, errors.Wrap(err, "initial aws marshaling")
	}
	// we need to override the default marshaling for the address;
	// ddbattr.MarshalMap isn't quite smart enough for our crazy type patterns
	m[addressField], err = dynamodbattribute.Marshal(b.Address.String())
	if err != nil {
		return nil, errors.Wrap(err, "override address marshaling")
	}
	return m, nil
}

// Add a bad address to the DB
func Add(ddb *dynamodb.DynamoDB, addr BadAddress, verbose bool) error {
	if verbose {
		fmt.Printf("unmarshaled form:\n%#v\n", addr)
	}
	item, err := addr.marshal()
	if err != nil {
		return errors.Wrap(err, "marshaling bad address")
	}
	if verbose {
		fmt.Printf("marshaled form:\n%#v\n", item)
	}
	_, err = ddb.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(Table),
		Item:      item,
	})
	if err != nil {
		return errors.Wrap(err, "put address in db")
	}
	return nil
}

// Remove a bad address from teh DB
func Remove(ddb *dynamodb.DynamoDB, addr BadAddress, verbose bool) error {
	fmt.Println("remove: unimplemented")
	return nil
}

// Check whether an address exists in the bad address DB
func Check(ddb *dynamodb.DynamoDB, addr address.Address) (bool, error) {
	av := dynamodb.AttributeValue{}
	av.SetS(addr.String())
	gio, err := ddb.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(Table),
		Key: map[string]*dynamodb.AttributeValue{
			addressField: &av,
		},
	})
	if err != nil {
		return false, errors.Wrap(err, "checking whether address is in db")
	}
	return gio != nil && len(gio.Item) > 0, nil
}
