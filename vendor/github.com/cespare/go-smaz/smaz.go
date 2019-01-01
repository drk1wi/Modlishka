// Package smaz is an implementation of the smaz library
// (https://github.com/antirez/smaz) for compressing small strings.
package smaz

import "errors"

var (
	codeStrings = []string{" ",
		"the", "e", "t", "a", "of", "o", "and", "i", "n", "s", "e ", "r", " th",
		" t", "in", "he", "th", "h", "he ", "to", "\r\n", "l", "s ", "d", " a", "an",
		"er", "c", " o", "d ", "on", " of", "re", "of ", "t ", ", ", "is", "u", "at",
		"   ", "n ", "or", "which", "f", "m", "as", "it", "that", "\n", "was", "en",
		"  ", " w", "es", " an", " i", "\r", "f ", "g", "p", "nd", " s", "nd ", "ed ",
		"w", "ed", "http://", "for", "te", "ing", "y ", "The", " c", "ti", "r ", "his",
		"st", " in", "ar", "nt", ",", " to", "y", "ng", " h", "with", "le", "al", "to ",
		"b", "ou", "be", "were", " b", "se", "o ", "ent", "ha", "ng ", "their", "\"",
		"hi", "from", " f", "in ", "de", "ion", "me", "v", ".", "ve", "all", "re ",
		"ri", "ro", "is ", "co", "f t", "are", "ea", ". ", "her", " m", "er ", " p",
		"es ", "by", "they", "di", "ra", "ic", "not", "s, ", "d t", "at ", "ce", "la",
		"h ", "ne", "as ", "tio", "on ", "n t", "io", "we", " a ", "om", ", a", "s o",
		"ur", "li", "ll", "ch", "had", "this", "e t", "g ", "e\r\n", " wh", "ere",
		" co", "e o", "a ", "us", " d", "ss", "\n\r\n", "\r\n\r", "=\"", " be", " e",
		"s a", "ma", "one", "t t", "or ", "but", "el", "so", "l ", "e s", "s,", "no",
		"ter", " wa", "iv", "ho", "e a", " r", "hat", "s t", "ns", "ch ", "wh", "tr",
		"ut", "/", "have", "ly ", "ta", " ha", " on", "tha", "-", " l", "ati", "en ",
		"pe", " re", "there", "ass", "si", " fo", "wa", "ec", "our", "who", "its", "z",
		"fo", "rs", ">", "ot", "un", "<", "im", "th ", "nc", "ate", "><", "ver", "ad",
		" we", "ly", "ee", " n", "id", " cl", "ac", "il", "</", "rt", " wi", "div",
		"e, ", " it", "whi", " ma", "ge", "x", "e c", "men", ".com",
	}

	codes    = make([][]byte, len(codeStrings))
	codeTrie trieNode
)

func init() {
	for i, code := range codeStrings {
		codes[i] = []byte(code)
		codeTrie.put([]byte(code), byte(i))
	}
}

// A trieNode represents a logical vertex in the trie structure.
// The trie maps []byte -> byte.
type trieNode struct {
	branches [256]*trieNode
	val      byte
	terminal bool
}

// put inserts the mapping k -> v into the trie, overwriting any previous value.
// It returns true if the element was not previously in t.
func (n *trieNode) put(k []byte, v byte) bool {
	for _, c := range k {
		next := n.branches[int(c)]
		if next == nil {
			next = &trieNode{}
			n.branches[c] = next
		}
		n = next
	}
	n.val = v
	if n.terminal {
		return false
	}
	n.terminal = true
	return true
}

func flushVerb(out, verb []byte) []byte {
	// We can write a max of 255 continuous verbatim characters,
	// because the length of the continuous verbatim section is represented
	// by a single byte.
	var chunk []byte
	for len(verb) > 0 {
		if len(verb) < 255 {
			chunk, verb = verb, nil
		} else {
			chunk, verb = verb[:255], verb[255:]
		}
		if len(chunk) == 1 {
			// 254 is code for a single verbatim byte.
			out = append(out, 254)
		} else {
			// 255 is code for a verbatim string.
			// It is followed by a byte containing the length of the string.
			out = append(out, 255, byte(len(chunk)))
		}
		out = append(out, chunk...)
	}
	return out
}

// Compress compresses a byte slice and returns the compressed data.
func Compress(input []byte) []byte {
	out := make([]byte, 0, len(input)/2) // estimate output size
	var verb []byte

	for len(input) > 0 {
		prefixLen := 0
		var code byte
		n := &codeTrie
		for i, c := range input {
			next := n.branches[int(c)]
			if next == nil {
				break
			}
			n = next
			if n.terminal {
				prefixLen = i + 1
				code = n.val
			}
		}

		if prefixLen > 0 {
			input = input[prefixLen:]
			out = flushVerb(out, verb)
			verb = verb[:0]
			out = append(out, code)
		} else {
			verb = append(verb, input[0])
			input = input[1:]
		}
	}
	return flushVerb(out, verb)
}

// ErrDecompression is returned when decompressing invalid smaz-encoded data.
var ErrDecompression = errors.New("invalid or corrupted compressed data")

// Decompress decompresses a smaz-compressed byte slice and return a new slice
// with the decompressed data.
// err is nil if and only if decompression fails for any reason
// (e.g., corrupted data).
func Decompress(b []byte) ([]byte, error) {
	dec := make([]byte, 0, len(b)) // estimate initial size

	for len(b) > 0 {
		switch b[0] {
		case 254: // verbatim byte
			if len(b) < 2 {
				return nil, ErrDecompression
			}
			dec = append(dec, b[1])
			b = b[2:]
		case 255: // verbatim string
			if len(b) < 2 {
				return nil, ErrDecompression
			}
			n := int(b[1])
			if len(b) < n+2 {
				return nil, ErrDecompression
			}
			dec = append(dec, b[2:n+2]...)
			b = b[n+2:]
		default: // look up encoded value
			dec = append(dec, codes[int(b[0])]...)
			b = b[1:]
		}
	}

	return dec, nil
}
