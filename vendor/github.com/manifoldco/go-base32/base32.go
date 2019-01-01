// Package base32 implements unpadded base32 encoding, using our chosen
// alphabet.
package base32

import (
	"encoding/base32"
	"strings"
)

const base32Alphabet = "0123456789abcdefghjkmnpqrtuvwxyz"

var lowerBase32 = base32.NewEncoding(base32Alphabet)

// EncodeToString encodes the given byte slice in base32
func EncodeToString(in []byte) string {
	return strings.TrimRight(lowerBase32.EncodeToString(in), "=")
}

// DecodeString decodes the given base32 encodeed bytes
func DecodeString(raw string) ([]byte, error) {
	pad := 8 - (len(raw) % 8)
	nb := []byte(raw)
	if pad != 8 {
		nb = make([]byte, len(raw)+pad)
		copy(nb, raw)
		for i := 0; i < pad; i++ {
			nb[len(raw)+i] = '='
		}
	}

	return lowerBase32.DecodeString(string(nb))
}
