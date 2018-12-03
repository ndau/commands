package main

import (
	"errors"
	"strings"
)

var wordlists = map[string][]string{
	"en": _english,
}

// getMask returns a byte offset and a mask for a given bit index
func getMask(n int) (int, byte) {
	byteix := n / 8
	bitix := n % 8
	mask := byte(1 << uint(7-bitix))
	return byteix, mask
}

// nthBit returns the nth bit from b as an int (either 0 or 1),
// treating b as a big-endian continuous array of bits.
// For example, nthBit(9, []byte{0xAB, 0xCD}) is 1, because 0xABCD is
// 10101011 11001111 and we are selecting the bit with the ^ under it).
//           ^
func nthBit(n int, b []byte) int {
	byteix, mask := getMask(n)
	if b[byteix]&mask == 0 {
		return 0
	}
	return 1
}

// setBit sets the nth bit in b to zero if v is zero, otherwise 1
func setBit(n int, b []byte, v int) []byte {
	byteix, mask := getMask(n)
	if v == 0 {
		b[byteix] &= ^mask
	} else {
		b[byteix] |= mask
	}
	return b
}

// getRun returns a continuous run of bits starting at a given point
func getRun(start int, runlen int, b []byte) int {
	r := 0
	for i := 0; i < runlen; i++ {
		r = (r << 1) | nthBit(start+i, b)
	}
	return r
}

// nthRun returns the nth run of the given runlen from b as an integer,
// treating b as a big-endian continuous array of bits.
// For example, nthrun(2, 4, []byte{0xAB, 0xCD}) is 12 (0xC).
func nthRun(n int, runlen int, b []byte) int {
	return getRun(n*runlen, runlen, b)
}

// setRun sets a run of bits to the values in run
func setRun(n int, runlen int, b []byte, run int) []byte {
	for i := 0; i < runlen; i++ {
		mask := 1 << uint(runlen-i-1)
		bit := run & mask
		setBit(n*runlen+i, b, bit)
	}
	return b
}

// crc8 computes an 8-bit CRC over a byte array
// this is a bitwise calculation but the goal here is to minimize code and memory size
// rather than performance; we don't use it a lot.
func crc8(b []byte) byte {
	const generator = byte(0x1D)
	var crc byte

	for _, c := range b {
		crc ^= c

		for i := 0; i < 8; i++ {
			if (crc & 0x80) != 0 {
				crc = (crc << 1) ^ generator
			} else {
				crc <<= 1
			}
		}
	}

	return crc
}

func nextMultipleOf(n, m int) int {
	if n%m == 0 {
		return n
	}
	return m * ((n / m) + 1)
}

// PadData takes a number of bytes of input and pads it to a multiple of 11 bits
// making sure that that at least 4 bits are available for a crc check.
func PadData(input []byte) (int, []byte) {
	nbits := len(input) * 8
	// make sure that we have at least 4 bits for a crc
	nbits11 := nextMultipleOf(nbits+4, 11)
	nbytesneeded := nextMultipleOf(nbits11, 8) / 8

	result := make([]byte, nbytesneeded)
	for i := range input {
		result[i] = input[i]
	}
	crc := int(crc8(input))
	crclen := nbits11 - nbits

	// now set the crc bits
	// This is the last N bits of the CRC where N is the
	// number of leftover bits after the data
	m := 1 << uint(crclen-1)
	for i := 0; i < crclen; i++ {
		if m&int(crc) != 0 {
			setBit(nbits+i, result, 1)
		}
		m >>= 1
	}
	return nbits11 / 11, result
}

// WordsFromBytes generates the list of words corresponding to a given
// sequence of bytes.
func WordsFromBytes(lang string, b []byte) ([]string, error) {
	nwords, data := PadData(b)
	wordlist, ok := wordlists[lang]
	if !ok {
		return nil, errors.New("invalid language code")
	}
	output := make([]string, nwords)
	for w := 0; w < nwords; w++ {
		output[w] = wordlist[nthRun(w, 11, data)]
	}
	return output, nil
}

// lookupWord does a binary search for a word in the wordlist
// This isn't a frequent occurrence, so this is better than using
// extra memory to store a hash of the words
func lookupWord(lang, s string) (int, error) {
	wordlist, ok := wordlists[lang]
	if !ok {
		return 0, errors.New("invalid language code")
	}
	// do a binary search for the word
	// loop invariant - min <= s < max
	min := 0
	max := len(wordlist)
	for {
		if min >= max {
			return 0, errors.New("word not found in wordlist")
		}
		n := (min + max) / 2
		switch strings.Compare(s, wordlist[n]) {
		case 0:
			return n, nil
		case -1:
			max = n
		case 1:
			min = n + 1
		}
	}
}

// BytesFromWords returns an array of the bytes a list of words corresponds to.
// It can error if lookup of any of the words fails.
func BytesFromWords(lang string, s []string) ([]byte, error) {
	nbits := len(s) * 11
	nbytes := nbits / 8
	resultbytes := nbytes
	if nbytes*8 < nbits {
		nbytes++
	}
	result := make([]byte, nbytes)
	for n, w := range s {
		value, err := lookupWord(lang, w)
		if err != nil {
			return nil, err
		}
		setRun(n, 11, result, value)
	}

	// if a shorter version (with a longer crc) is possible, test it first
	// because if this crc is correct it's more likely to be the right answer
	for _, rb := range []int{resultbytes - 1, resultbytes} {
		runlen := nbits - rb*8
		mask := (1 << uint(runlen)) - 1 // generate runlen 1 bits
		ckcalc := int(crc8(result[:rb])) & mask
		ckfound := getRun(rb*8, runlen, result)
		if ckcalc == ckfound {
			return result[:rb], nil
		}
	}

	return nil, errors.New("checksum failed; word list not valid or not created by this app")
}
