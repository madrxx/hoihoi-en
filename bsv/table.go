// Package bsv provides types and helpers for the BSV binary table format
// used by Hoihoi-san for item databases, mission records, event scripts,
// and probably other stuff.
//
// A BSV table is a tagged-section format with four standard chunks:
//
//	SRTS  -- string pool (null-terminated Shift-JIS strings)
//	EMAN  -- string references (name/value pairs)
//	DNIK  -- string references (bulk u32 fields)
//	ATAD  -- row data (fixed-size records referencing the string pool)
//
// Each table has a base address (file offset within a UFP entry) and the
// sections are located by scanning for 4-byte tags.
//
// This package handles the structural parsing and offset arithmetic. The
// actual binary read/write operations are performed through the ReaderWriter
// interface, allowing callers to provide their own disc I/O layer.
package bsv

import (
	"log"

	"github.com/madrxx/hoihoi-en/encoding"
)

// ReaderWriter abstracts the disc I/O operations needed by the BSV package.
// The Patcher in the parent package implements this interface.
type ReaderWriter interface {
	ReadFileBytes(filePath string, fileOffset uint64, size uint64) []byte
	WriteFileBytes(filePath string, fileOffset uint64, data []byte)
	ReadFileU32LE(filePath string, fileOffset uint64) uint32
	WriteFileU32LE(filePath string, fileOffset uint64, value uint32)
}

// Table describes the layout of a BSV table within a UFP entry.
// All offsets are relative to the file containing the table.
type Table struct {
	FilePath string
	Base     uint64

	SRTSOffset uint64
	SRTSSize   uint64

	EMANOffset uint64
	EMANSize   uint64

	DNIKOffset uint64
	DNIKSize   uint64

	ATADOffset uint64
	ATADSize   uint64

	RowCount        uint32
	RowSize         uint64
	ATADTrailerSize uint64
}

// SRTSPoolStart returns the byte offset where the string pool begins.
// This is always Base + 0x30 (after the BSV header).
func (t Table) SRTSPoolStart() uint64 {
	return t.Base + 0x30
}

// SRTSEnd returns one past the last byte of the declared SRTS section.
func (t Table) SRTSEnd() uint64 {
	return t.SRTSOffset + 0x08 + t.SRTSSize
}

// RowOffset returns the byte offset of the start of a row within ATAD.
// Row 0 begins at ATADOffset + 0x10 (after the ATAD header).
func (t Table) RowOffset(row int) uint64 {
	if row < 0 || row >= int(t.RowCount) {
		log.Fatalf("BSV row out of range: table=0x%X row=%d rows=%d", t.Base, row, t.RowCount)
	}
	return t.ATADOffset + 0x10 + uint64(row)*t.RowSize
}

// FieldOffset returns the byte offset of a word within a row.
func (t Table) FieldOffset(row int, word int) uint64 {
	fieldOffset := t.RowOffset(row) + uint64(word)*4
	rowEnd := t.RowOffset(row) + t.RowSize

	if fieldOffset+4 > rowEnd {
		log.Fatalf(
			"BSV field out of row range: table=0x%X row=%d word=%d rowSize=0x%X",
			t.Base,
			row,
			word,
			t.RowSize,
		)
	}

	return fieldOffset
}

// ----
// String pool
// ----

// StringPool manages packing strings into a BSV SRTS region.
// It deduplicates identical strings and tracks the write cursor.
//
// The actual writing is done by the Patcher (via WriteFileBytes); this
// struct only tracks what has been written and where.
type StringPool struct {
	Table  Table
	cursor uint64
	start  uint64
	end    uint64
	seen   map[string]uint32
}

// NewStringPool creates a string pool spanning the entire SRTS area.
func NewStringPool(t Table) StringPool {
	return NewStringPoolRange(t, t.SRTSPoolStart(), t.SRTSEnd())
}

// NewStringPoolRange creates a string pool restricted to [start, end).
// Both must be within the SRTS area.
func NewStringPoolRange(t Table, start uint64, end uint64) StringPool {
	if start < t.SRTSPoolStart() || start > t.SRTSEnd() {
		log.Fatalf(
			"BSV string pool start out of range: table=0x%X start=0x%X valid=0x%X..0x%X",
			t.Base,
			start,
			t.SRTSPoolStart(),
			t.SRTSEnd(),
		)
	}

	if end < start || end > t.SRTSEnd() {
		log.Fatalf(
			"BSV string pool end out of range: table=0x%X end=0x%X valid=0x%X..0x%X",
			t.Base,
			end,
			start,
			t.SRTSEnd(),
		)
	}

	return StringPool{
		Table:  t,
		cursor: start,
		start:  start,
		end:    end,
		seen:   make(map[string]uint32),
	}
}

// Cursor returns the current write position.
func (pool *StringPool) Cursor() uint64 { return pool.cursor }

// Start returns the pool start offset.
func (pool *StringPool) Start() uint64 { return pool.start }

// End returns the pool end offset.
func (pool *StringPool) End() uint64 { return pool.end }

// SetCursor updates the write position.
func (pool *StringPool) SetCursor(c uint64) { pool.cursor = c }

// Advance moves the cursor forward by n bytes.
func (pool *StringPool) Advance(n uint64) { pool.cursor += n }

// Dedup checks whether a string has already been written and returns its
// relative offset. If not yet seen, it records the given rel offset.
func (pool *StringPool) Dedup(key string, rel uint32) {
	pool.seen[key] = rel
}

// Seen returns true if the key has already been written to the pool.
func (pool *StringPool) Seen(key string) bool {
	_, ok := pool.seen[key]
	return ok
}

// RelFor returns the previously-recorded relative offset for a key.
func (pool *StringPool) RelFor(key string) uint32 {
	return pool.seen[key]
}

// ----
// Field reference
// ----

// FieldRef describes a reference to a string in a BSV table.
// FileOffset is the offset of the u32 pointer field, and Raw is the
// current string data (without null terminator).
type FieldRef struct {
	FileOffset uint64
	Raw        []byte
}

// ----
// Text encoding
// ----

// EncodePatchText encodes a Go string into the on-disc format expected by
// the game: fullwidth -> Shift-JIS, without the usual null terminator.
// The caller (BSV string pool) appends its own null byte.
func EncodePatchText(input string) []byte {
	if input == "" {
		return nil
	}

	encoded := encoding.GameText(input)
	if len(encoded) > 0 && encoded[len(encoded)-1] == 0x00 {
		encoded = encoded[:len(encoded)-1]
	}

	out := make([]byte, len(encoded))
	copy(out, encoded)
	return out
}
