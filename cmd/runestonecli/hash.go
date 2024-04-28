package main

import "encoding/hex"

const HashLength = 32

type Hash [HashLength]byte

var ZeroHash = Hash{}

// BytesToHash sets b to hash.
// If b is larger than len(h), b will be cropped from the left.
func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}
func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLength:]
	}

	copy(h[HashLength-len(b):], b)
}

// BtcString returns the Hash as the hexadecimal string of the byte-reversed hash.
func (hash Hash) BtcString() string {
	for i := 0; i < HashLength/2; i++ {
		hash[i], hash[HashLength-1-i] = hash[HashLength-1-i], hash[i]
	}
	return hex.EncodeToString(hash[:])
}
func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}
func HexToHash(s string) Hash {

	return BytesToHash(FromHex(s))
}

// FromHex returns the bytes represented by the hexadecimal string s.
// s may be prefixed with "0x".
func FromHex(s string) []byte {
	if has0xPrefix(s) {
		s = s[2:]
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	return Hex2Bytes(s)
}

// has0xPrefix validates str begins with '0x' or '0X'.
func has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

// Hex2Bytes returns the bytes represented by the hexadecimal string str.
func Hex2Bytes(str string) []byte {
	h, _ := hex.DecodeString(str)
	return h
}
