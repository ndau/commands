package baddress

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/ndau/ndaumath/pkg/address"
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

func addrKey(addr address.Address) map[string]*dynamodb.AttributeValue {
	av := dynamodb.AttributeValue{}
	av.SetS(addr.String())
	return map[string]*dynamodb.AttributeValue{
		addressField: &av,
	}
}

// Remove a bad address from the DB
func Remove(ddb *dynamodb.DynamoDB, addr address.Address, verbose bool) error {
	_, err := ddb.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(Table),
		Key:       addrKey(addr),
	})
	if err != nil {
		return errors.Wrap(err, "removing key from db")
	}
	return nil
}

// Check whether an address exists in the bad address DB
func Check(ddb *dynamodb.DynamoDB, addr address.Address) (bool, error) {
	gio, err := ddb.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(Table),
		Key:       addrKey(addr),
	})
	if err != nil {
		return false, errors.Wrap(err, "checking whether address is in db")
	}
	return gio != nil && len(gio.Item) > 0, nil
}
