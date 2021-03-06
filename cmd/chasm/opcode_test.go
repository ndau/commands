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
	"testing"

	"github.com/ndau/chaincode/pkg/vm"
	"github.com/stretchr/testify/assert"
)

func TestToBytes(t *testing.T) {
	bcheck(t, vm.ToBytes(1), "0100000000000000")
	bcheck(t, vm.ToBytes(255), "FF00000000000000")
	bcheck(t, vm.ToBytes(1000000000000), "0010A5D4E8000000")
	bcheck(t, vm.ToBytes(-1), "FFFFFFFFFFFFFFFF")
	bcheck(t, vm.ToBytes(0x1122334455667788), "8877665544332211")
}

func TestToBytesU(t *testing.T) {
	bcheck(t, vm.ToBytesU(1), "0100000000000000")
	bcheck(t, vm.ToBytesU(0xFFFFFFFFFFFFFFFF), "FFFFFFFFFFFFFFFF")
	bcheck(t, vm.ToBytes(0x1122334455667788), "8877665544332211")
}

func TestZero(t *testing.T) {
	op, err := newPushOpcode("0")
	assert.Nil(t, err)
	b := op.bytes()
	bcheck(t, b, "20")
}

func TestOne(t *testing.T) {
	op, err := newPushOpcode("1")
	assert.Nil(t, err)
	b := op.bytes()
	bcheck(t, b, "1A")
}

func TestNegOne(t *testing.T) {
	op, err := newPushOpcode("-1")
	assert.Nil(t, err)
	b := op.bytes()
	bcheck(t, b, "1B")
}

func TestPushLobyte(t *testing.T) {
	op, err := newPushOpcode("17")
	assert.Nil(t, err)
	b := op.bytes()
	bcheck(t, b, "2111")
}

func TestPushHibyteZero(t *testing.T) {
	// a value in the range from 128 to 255 should express as a push2 rather than
	// a push1, because we don't want it to be sign-extended.
	op, err := newPushOpcode("192")
	assert.Nil(t, err)
	b := op.bytes()
	bcheck(t, b, "22C000")
}

func TestPushNeg2(t *testing.T) {
	op, err := newPushOpcode("-2")
	assert.Nil(t, err)
	b := op.bytes()
	bcheck(t, b, "21FE")
}

func TestPushNeg200(t *testing.T) {
	// a value in the range from -128 to -255 should also express as a push2 with an FF
	// for sign extension
	op, err := newPushOpcode("-207")
	assert.Nil(t, err)
	b := op.bytes()
	bcheck(t, b, "2231FF")
}

func TestPush2Bytes(t *testing.T) {
	op, err := newPushOpcode("0x3478")
	assert.Nil(t, err)
	b := op.bytes()
	bcheck(t, b, "227834")
}

func TestPush3Bytes(t *testing.T) {
	op, err := newPushOpcode("0x125678")
	assert.Nil(t, err)
	b := op.bytes()
	bcheck(t, b, "23785612")
}

func TestPush4Bytes(t *testing.T) {
	op, err := newPushOpcode("0x12345678")
	assert.Nil(t, err)
	b := op.bytes()
	bcheck(t, b, "2478563412")
}

func TestPush5Bytes(t *testing.T) {
	op, err := newPushOpcode("0x123456780A")
	assert.Nil(t, err)
	b := op.bytes()
	bcheck(t, b, "250A78563412")
}

func TestPush6BytesHighBit(t *testing.T) {
	// this tests that a value with the high bit set gets encoded with 1 extra byte
	// this positive value is 5 non-zero bytes but needs to be a push6 opcode because
	// the high bit is set
	op, err := newPushOpcode("1000000000000")
	assert.Nil(t, err)
	b := op.bytes()
	bcheck(t, b, "260010A5D4E800")
}

func TestPush6Bytes(t *testing.T) {
	op, err := newPushOpcode("0x12345678AA55")
	assert.Nil(t, err)
	b := op.bytes()
	bcheck(t, b, "2655AA78563412")
}

func TestPush7Bytes(t *testing.T) {
	op, err := newPushOpcode("0x12345678CAFEDA")
	assert.Nil(t, err)
	b := op.bytes()
	bcheck(t, b, "27DAFECA78563412")
}

func TestPush8BytesPositive(t *testing.T) {
	op, err := newPushOpcode("0x1BAD1DEACAFEBABE")
	assert.Nil(t, err)
	b := op.bytes()
	bcheck(t, b, "28BEBAFECAEA1DAD1B")
}

func TestPushAddress(t *testing.T) {
	op, err := newPushAddr("ndadprx764ciigti8d8whtw2kct733r85qvjukhqhke3dka4")
	assert.Nil(t, err)
	b := op.bytes()
	bcheck(t, b, "2a306e64616470727837363463696967746938643877687477326b637437333372383571766a756b6871686b6533646b6134")
}

func TestPushBadAddress(t *testing.T) {
	_, err := newPushAddr("ndadprx764ciigti8d8whxw2kct733r85qvjukhqhke3dka4")
	assert.NotNil(t, err)
}

func TestPushTimestamp(t *testing.T) {
	op, err := newPushTimestamp("2018-07-18T20:00:58Z")
	assert.Nil(t, err)
	b := op.bytes()
	bcheck(t, b, "2b 80 b2 2c 4a 4a 14 02  00")
}
