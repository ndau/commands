package main

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

func Test_nthBit(t *testing.T) {
	type args struct {
		n int
		b []byte
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"a", args{0, []byte{0xA0}}, 1},
		{"b", args{1, []byte{0xA0}}, 0},
		{"c", args{2, []byte{0xA0}}, 1},
		{"d", args{9, []byte{0xAB, 0xCD}}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nthBit(tt.args.n, tt.args.b); got != tt.want {
				t.Errorf("nthBit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nthRun(t *testing.T) {
	type args struct {
		n      int
		runlen int
		b      []byte
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"a", args{0, 2, []byte{0xC5}}, 3},
		{"b", args{1, 2, []byte{0xC5}}, 0},
		{"c", args{2, 2, []byte{0xC5}}, 1},
		{"d", args{2, 4, []byte{0xAB, 0xCD}}, 12},
		{"e", args{2, 11, []byte{0x01, 0x23, 0x45, 0x67, 0x89}}, 0x2CF},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nthRun(tt.args.n, tt.args.runlen, tt.args.b); got != tt.want {
				t.Errorf("nthRun() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPadData(t *testing.T) {
	type args struct {
		input []byte
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 []byte
	}{
		// because of the way our CRC works, input of 0 will have a CRC of 0
		{"a", args{[]byte{0x00, 0x00}}, 2, []byte{0x00, 0x00, 0x00}},
		// but change one bit and the crc changes
		{"b", args{[]byte{0x80, 0x00}}, 2, []byte{0x80, 0x00, 36}},
		// here's a more normal short one
		{"c", args{[]byte{0x82, 0x41}}, 2, []byte{0x82, 0x41, 124}},
		// and this is the basic 16-byte test; it should generate a 12-word result
		{"d", args{[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}}, 12,
			[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 32}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := PadData(tt.args.input)
			if got != tt.want {
				t.Errorf("PadData() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("PadData() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestWordsFromBytes(t *testing.T) {
	type args struct {
		lang string
		b    []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		// a simple input yields a simple output
		{"a", args{"en", []byte{0, 1, 2}}, []string{"abandon", "amount", "mom"}, false},
		// this is the generic basic test
		{"b", args{"en", []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}},
			strings.Split("abandon amount liar amount expire adjust cage candy arch gather drum bundle", " "), false},
		// if you change the first byte, the last and the first word should change
		{"c", args{"en", []byte{100, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}},
			strings.Split("goat amount liar amount expire adjust cage candy arch gather drum buyer", " "), false},
		// these next items were failures identified early in fuzz testing
		{"d", args{"en", []byte{15, 218, 104}}, strings.Split("average spy bicycle", " "), false},
		{"e", args{"en", []byte{91, 165, 36, 247, 53, 142, 251, 181, 184, 50, 32, 207, 88, 99, 108, 188, 64, 207, 172, 154, 235, 60, 200, 192}},
			strings.Split("forum circle differ help use suspect this dune soon seek swamp joy artefact stone hill guide silver addict", " "), false},
		{"f", args{"en", []byte{41, 247, 253, 146, 141, 146, 202, 67, 241, 147, 222, 228, 127, 89, 21, 73, 245, 151, 168, 17, 200}},
			strings.Split("clarify say gorilla brass coach capable shock knock tongue width earn negative floor staff elbow aim", " "), false},
		// check that a bad language is detected
		{"g", args{"sp", []byte{0, 1, 2}}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := WordsFromBytes(tt.args.lang, tt.args.b)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WordsFromBytes() = %v, want %v", got, tt.want)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("Ndau.Sub() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestBytesFromWords(t *testing.T) {
	type args struct {
		lang string
		s    []string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// these are the reverse of all of the tests above
		{"a", args{"en", []string{"abandon", "amount", "mom"}}, []byte{0, 1, 2}, false},
		{"b", args{"en", strings.Split("abandon amount liar amount expire adjust cage candy arch gather drum bundle", " ")},
			[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, false},
		{"c", args{"en", strings.Split("goat amount liar amount expire adjust cage candy arch gather drum buyer", " ")},
			[]byte{100, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, false},
		{"d", args{"en", strings.Split("average spy bicycle", " ")}, []byte{15, 218, 104}, false},
		{"e", args{"en", strings.Split("forum circle differ help use suspect this dune soon seek swamp joy artefact stone hill guide silver addict", " ")},
			[]byte{91, 165, 36, 247, 53, 142, 251, 181, 184, 50, 32, 207, 88, 99, 108, 188, 64, 207, 172, 154, 235, 60, 200, 192},
			false},
		{"f", args{"en", strings.Split("clarify say gorilla brass coach capable shock knock tongue width earn negative floor staff elbow aim", " ")},
			[]byte{41, 247, 253, 146, 141, 146, 202, 67, 241, 147, 222, 228, 127, 89, 21, 73, 245, 151, 168, 17, 200},
			false},
		// check for bad language
		{"g", args{"sp", []string{"abandon", "amount", "mom"}}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BytesFromWords(tt.args.lang, tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("BytesFromWords() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BytesFromWords() = %v, want %v", got, tt.want)
			}
		})
	}
}

// This does a randomized test of the roundtrip -- it generates random lengths (1-32) of random data bytes
// then converts them to words and back to data; do this 10K times and we're reasonably confident that
// the algorithm works properly.
func Test_RoundTrip(t *testing.T) {
	for i := 0; i < 10000; i++ {
		nbytes := rand.Intn(32) + 1
		b := make([]byte, nbytes)
		for j := 0; j < nbytes; j++ {
			b[j] = byte(rand.Intn(256))
		}
		words, err := WordsFromBytes("en", b)
		if err != nil {
			t.Error(err)
		}
		b2, err := BytesFromWords("en", words)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(b, b2) {
			t.Errorf("Round Trip failure: input %v, got %v", b, b2)
		}
	}
}

func Test_lookupWord(t *testing.T) {
	type args struct {
		lang string
		s    string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		// test the first, second, and last words, plus one in the niddle, plus
		// one that's not there
		{"a", args{"en", "abandon"}, 0, false},
		{"b", args{"en", "ability"}, 1, false},
		{"c", args{"en", "roof"}, 1501, false},
		{"d", args{"en", "zoo"}, 2047, false},
		{"e", args{"en", "foo"}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lookupWord(tt.args.lang, tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("lookupWord() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("lookupWord() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setBit(t *testing.T) {
	type args struct {
		n int
		b []byte
		v int
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// make sure we can set arbitrary bits in an array of bytes
		{"a", args{0, []byte{0}, 1}, []byte{0x80}},
		{"b", args{7, []byte{0}, 1}, []byte{0x01}},
		{"c", args{3, []byte{0}, 1}, []byte{0x10}},
		{"d", args{13, []byte{0, 0}, 1}, []byte{0x00, 0x04}},
		// and also clear them
		{"e", args{13, []byte{0xFF, 0xFF}, 0}, []byte{0xFF, 0xFB}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := setBit(tt.args.n, tt.args.b, tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("setBit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setRun(t *testing.T) {
	type args struct {
		n      int
		runlen int
		b      []byte
		run    int
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// set an arbitrary chunk of 11 bits to 1s and 0s.
		{"a", args{1, 11, []byte{0, 0, 0, 0}, 0x55}, []byte{0, 0x01, 0x54, 0}},
		// set 11 1s in a row
		{"b", args{1, 11, []byte{0, 0, 0, 0}, 0x7FF}, []byte{0, 0x1F, 0xFC, 0x00}},
		// set 11 zeros in a row
		{"c", args{1, 11, []byte{0xFF, 0xFF, 0xFF, 0xFF}, 0}, []byte{0xFF, 0xE0, 0x03, 0xFF}},
		// set at the left
		{"d", args{0, 11, []byte{0, 0, 0, 0}, 0x7FF}, []byte{0xFF, 0xE0, 0, 0}},
		// set at the right
		{"e", args{7, 11, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 0x7FF},
			[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0x07, 0xFF}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := setRun(tt.args.n, tt.args.runlen, tt.args.b, tt.args.run); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("setRun() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_crc8(t *testing.T) {
	tests := []struct {
		name string
		b    []byte
		want byte
	}{
		// crc of 0 is 0
		{"a", []byte{0}, 0},
		// crc of 1 is 0x1D (the polynomial for the 8 bit crc)
		{"b", []byte{1}, 0x1D},
		// simple one
		{"c", []byte{1, 2, 3, 4}, 62},
		// show that if we make a tiny change the crc changes a lot
		{"d", []byte("This is a test"), 214},
		{"e", []byte("this is a test"), 59},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := crc8(tt.b); got != tt.want {
				t.Errorf("crc8() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nextMultipleOf(t *testing.T) {
	type args struct {
		n int
		m int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// this function just gives us the next multiple of a particular value;
		// test some common occurrences in our system
		{"a", args{8, 11}, 11},
		{"b", args{24, 11}, 33},
		{"c", args{88, 11}, 88},
		{"d", args{128, 11}, 132},
		{"e", args{33, 8}, 40},
		{"f", args{132, 8}, 136},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nextMultipleOf(tt.args.n, tt.args.m); got != tt.want {
				t.Errorf("nextMultipleOf() = %v, want %v", got, tt.want)
			}
		})
	}
}
