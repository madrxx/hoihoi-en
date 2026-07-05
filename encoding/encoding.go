// Package encoding provides text encoding helpers for the Hoihoi-san
// English translation patch.
//
// The game stores all user-visible text as Shift-JIS encoded fullwidth
// characters. The functions in this package convert between Go strings and
// the on-disc representation:
//
//	go string  ->  toFullWidth  ->  toSJIS  ->  on-disc bytes
//	on-disc bytes  ->  fromSJIS  ->  go string
//
// GameText combines toFullWidth + toSJIS with a null terminator, which is
// the most common conversion needed when writing patch text into the binary.
// GameText panics on errors (they are unreachable for ASCII input  -- all
// patch strings are ASCII literals).
package encoding

import (
	"bytes"
	"errors"
	"fmt"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// ToFullWidth converts ASCII characters to their fullwidth equivalents as
// used by the game's text rendering system. Spaces become ideographic spaces
// (U+3000), ¥ becomes fullwidth yen (U+FFE5), and printable ASCII (0x21-0x7E)
// are shifted by 0xFEE0. Characters already outside ASCII (kana, kanji, newlines)
// pass through unchanged.
func ToFullWidth(input string) string {
	out := make([]rune, 0, len(input))

	for _, r := range input {
		switch {
		case r == ' ':
			out = append(out, '　')
		case r == '¥':
			out = append(out, '￥')
		case r >= 0x21 && r <= 0x7e:
			out = append(out, r+0xfee0)
		default:
			out = append(out, r)
		}
	}

	return string(out)
}

// ToSJIS encodes a Go string as Shift-JIS and appends nullBytes zero bytes.
// nullBytes must be >= 0 (a programming error; panics otherwise).
func ToSJIS(input string, nullBytes int) ([]byte, error) {
	if nullBytes < 0 {
		return nil, errors.New("nullBytes must be >= 0")
	}
	var buf bytes.Buffer

	writer := transform.NewWriter(&buf, japanese.ShiftJIS.NewEncoder())

	if _, err := writer.Write([]byte(input)); err != nil {
		return nil, fmt.Errorf("SJIS encode: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("SJIS close: %w", err)
	}

	output := append([]byte(nil), buf.Bytes()...)
	output = append(output, make([]byte, nullBytes)...)
	return output, nil
}

// FromSJIS decodes a Shift-JIS byte slice to a Go string.
func FromSJIS(input []byte) (string, error) {
	var buf bytes.Buffer

	reader := transform.NewReader(
		bytes.NewReader(input),
		japanese.ShiftJIS.NewDecoder(),
	)

	if _, err := buf.ReadFrom(reader); err != nil {
		return "", fmt.Errorf("SJIS decode: %w", err)
	}

	return buf.String(), nil
}

// GameText is the standard game text encoding: ASCII -> fullwidth -> Shift-JIS
// with a single null terminator byte appended.
//
// All patch text in this project is hard-coded ASCII. These inputs are
// trivially encodable to Shift-JIS and will never produce an error. If
// you need error handling (e.g. for user-supplied input), use ToSJIS
// directly.
func GameText(input string) []byte {
	out, err := ToSJIS(ToFullWidth(input), 1)
	if err != nil {
		panic("GameText encoding error (unreachable for ASCII input): " + err.Error())
	}
	return out
}
