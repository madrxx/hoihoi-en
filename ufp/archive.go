// Package ufp parses UFP archive files (UFV1 container format) used by
// Hoihoi-san to bundle PSI textures and BSV data tables.
//
// A UFP file consists of tagged chunks describing file records and string
// tables. The format is similar to RIFF but with 4-byte alignment instead of
// 16-byte. The game's CJLUfp::CreateHeader recognises ten chunk tags; this
// package handles the six content chunks and validates against three metadata
// chunks to catch corruption early:
//
//	AAHF  -- file record offset table
//	1AHF  -- file record data
//	AADP  -- string table offset table
//	1ADP  -- string table data
//	KCOE  -- end-of-container marker
//
//	TNCF  -- file count (validated against AAHF table size)
//	TNCP  -- string table entry count (validated against AADP table size)
//	ZISF  -- total UFP size in bytes (validated against len(data))
//
// Three additional tags (ZISH, TCES, and the UFV1 magic) are present in the
// binary but not needed for patching.
package ufp

import (
	"encoding/binary"
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/madrxx/hoihoi-en/encoding"
)

// Entry describes a single file stored within a UFP archive.
type Entry struct {
	// Zero-based index within the archive.
	Index int
	// Full path from archive root, e.g. "2d/menu/icon.psi".
	Path string
	// Byte offset of the file data within the UFP.
	Offset uint64
	// Uncompressed file size in bytes.
	Size uint64
	// Packed size (may be smaller if compressed; currently unused).
	PackedSize uint64
	// Bitfield flags from the file record.
	Flags uint32
}

// chunk is a tagged data section within a UFP file.
type chunk struct {
	Tag     string
	Offset  uint64
	Payload []byte
}

// ----

// U32 reads a little-endian uint32 from data at the given offset,
// fataling if the read would go out of bounds.
func U32(data []byte, off int) uint32 {
	if off < 0 || off+4 > len(data) {
		log.Fatalf("u32 read out of range: off=0x%X len=0x%X", off, len(data))
	}
	return binary.LittleEndian.Uint32(data[off : off+4])
}

// parseChunks splits a UFP file into its tagged chunks. Chunks are
// 4-byte aligned and parsing stops at the KCOE end marker.
func parseChunks(data []byte) []chunk {
	if len(data) < 8 || string(data[:4]) != "UFV1" {
		log.Fatalf("not a UFV1/UFP file")
	}

	var chunks []chunk
	pos := 0

	for pos+8 <= len(data) {
		tag := string(data[pos : pos+4])
		size := int(U32(data, pos+4))
		payloadStart := pos + 8
		payloadEnd := payloadStart + size

		if size < 0 || payloadEnd > len(data) {
			break
		}

		chunks = append(chunks, chunk{
			Tag:     tag,
			Offset:  uint64(payloadStart),
			Payload: data[payloadStart:payloadEnd],
		})

		pos = payloadEnd
		if pos%4 != 0 {
			pos += 4 - (pos % 4)
		}

		if tag == "KCOE" {
			break
		}
	}

	return chunks
}

// findChunk returns the first chunk with the given tag.
func findChunk(chunks []chunk, tag string) (chunk, bool) {
	for _, c := range chunks {
		if c.Tag == tag {
			return c, true
		}
	}
	return chunk{}, false
}

// cleanPath normalises a UFP entry path: backslash to forward slash,
// trims, and cleans.
func cleanPath(s string) string {
	s = strings.ReplaceAll(s, "\\", "/")
	s = strings.Trim(s, "/")
	s = path.Clean(s)
	if s == "." {
		return ""
	}
	return s
}

// ----
// Exported entry operations

// Parse extracts all file entries from a UFP archive.
func Parse(data []byte) ([]Entry, error) {
	chunks := parseChunks(data)

	pathOffsets, ok := findChunk(chunks, "AADP")
	if !ok {
		log.Fatalf("UFP AADP chunk not found")
	}

	pathStrings, ok := findChunk(chunks, "1ADP")
	if !ok {
		log.Fatalf("UFP 1ADP chunk not found")
	}

	fileOffsets, ok := findChunk(chunks, "AAHF")
	if !ok {
		log.Fatalf("UFP AAHF chunk not found")
	}

	fileRecords, ok := findChunk(chunks, "1AHF")
	if !ok {
		log.Fatalf("UFP 1AHF chunk not found")
	}

	stringsTable, err := parseStringTable(pathOffsets.Payload, pathStrings.Payload)
	if err != nil {
		return nil, err
	}
	recordOffsets := parseOffsetTable(fileOffsets.Payload, "AAHF")

	if countChunk, ok := findChunk(chunks, "TNCF"); ok && len(countChunk.Payload) >= 4 {
		expected := int(U32(countChunk.Payload, 0))
		if expected != len(recordOffsets) {
			log.Fatalf("UFP file count mismatch: TNCF=%d AAHF entries=%d", expected, len(recordOffsets))
		}
	}

	if countChunk, ok := findChunk(chunks, "TNCP"); ok && len(countChunk.Payload) >= 4 {
		expected := int(U32(countChunk.Payload, 0))
		actual := len(pathOffsets.Payload) / 4
		if expected != actual {
			log.Fatalf("UFP string count mismatch: TNCP=%d AADP entries=%d", expected, actual)
		}
	}

	if sizeChunk, ok := findChunk(chunks, "ZISF"); ok && len(sizeChunk.Payload) >= 4 {
		expected := int(U32(sizeChunk.Payload, 0))
		if expected != len(data) {
			log.Fatalf("UFP size mismatch: ZISF=%d actual=%d", expected, len(data))
		}
	}

	entries := make([]Entry, 0, len(recordOffsets))

	for i, rawRecordOffset := range recordOffsets {
		recordOffset := int(rawRecordOffset)
		if recordOffset < 0 || recordOffset >= len(fileRecords.Payload) {
			log.Fatalf("UFP file record offset out of range: index=%d offset=0x%X tableSize=0x%X", i, recordOffset, len(fileRecords.Payload))
		}

		recordEnd := len(fileRecords.Payload)
		if i+1 < len(recordOffsets) {
			recordEnd = int(recordOffsets[i+1])
		}

		if recordEnd < recordOffset {
			log.Fatalf("UFP file record offsets not increasing: index=%d offset=0x%X next=0x%X", i, recordOffset, recordEnd)
		}

		record := fileRecords.Payload[recordOffset:recordEnd]
		if len(record) < 18 {
			log.Fatalf("UFP file record too short: index=%d len=0x%X", i, len(record))
		}

		if (len(record)-16)%2 != 0 {
			log.Fatalf("UFP file record has odd path id bytes: index=%d len=0x%X", i, len(record))
		}

		fileOffset := uint64(binary.LittleEndian.Uint32(record[0x00:0x04]))
		size := uint64(binary.LittleEndian.Uint32(record[0x04:0x08]))
		packedSize := uint64(binary.LittleEndian.Uint32(record[0x08:0x0C]))
		meta := binary.LittleEndian.Uint32(record[0x0C:0x10])

		idCount := (len(record) - 16) / 2
		ids := make([]uint16, idCount)
		for j := 0; j < idCount; j++ {
			ids[j] = binary.LittleEndian.Uint16(record[16+j*2 : 18+j*2])
		}

		entryPath := buildPath(stringsTable, meta, ids)
		if entryPath == "" || size == 0 {
			continue
		}

		if fileOffset+size > uint64(len(data)) {
			log.Fatalf(
				"UFP entry out of range: index=%d path=%s offset=0x%X size=0x%X ufpSize=0x%X",
				i,
				entryPath,
				fileOffset,
				size,
				len(data),
			)
		}

		entries = append(entries, Entry{
			Index:      i,
			Path:       entryPath,
			Offset:     fileOffset,
			Size:       size,
			PackedSize: packedSize,
			Flags:      meta,
		})
	}

	return entries, nil
}

// FindEntry looks up a UFP entry by path (case-insensitive). If the path
// has no extension, ".psi" is tried as a fallback suffix.
func FindEntry(entries []Entry, wanted string) (Entry, bool) {
	wanted = cleanPath(wanted)

	for _, entry := range entries {
		entryPath := cleanPath(entry.Path)
		if strings.EqualFold(entryPath, wanted) {
			return entry, true
		}
	}

	if !strings.Contains(path.Base(wanted), ".") {
		withPSI := wanted + ".psi"
		for _, entry := range entries {
			entryPath := cleanPath(entry.Path)
			if strings.EqualFold(entryPath, withPSI) {
				return entry, true
			}
		}
	}

	return Entry{}, false
}

// ReadEntry copies the entry's data out of the UFP byte slice.
func ReadEntry(ufp []byte, entry Entry) []byte {
	end := entry.Offset + entry.Size
	if end > uint64(len(ufp)) {
		log.Fatalf("UFP read out of range: %s offset=0x%X size=0x%X", entry.Path, entry.Offset, entry.Size)
	}
	out := make([]byte, entry.Size)
	copy(out, ufp[entry.Offset:end])
	return out
}

// WriteEntry overwrites an entry's data in-place within the UFP byte slice.
// The replacement must be exactly the same size as the original entry.
func WriteEntry(ufp []byte, entry Entry, replacement []byte) {
	if uint64(len(replacement)) != entry.Size {
		log.Fatalf("UFP replacement for %s must be same size: got=0x%X expected=0x%X", entry.Path, len(replacement), entry.Size)
	}
	end := entry.Offset + entry.Size
	if end > uint64(len(ufp)) {
		log.Fatalf("UFP write out of range: %s offset=0x%X size=0x%X", entry.Path, entry.Offset, entry.Size)
	}
	copy(ufp[entry.Offset:end], replacement)
}

// ----
// Internal string table helpers

// parseStringTable builds a string table from offset and payload chunks.
// Each string is length-prefixed (one byte) followed by Shift-JIS data.
func parseStringTable(offsetPayload []byte, stringPayload []byte) ([]string, error) {
	if len(offsetPayload)%4 != 0 {
		log.Fatalf("UFP AADP size 0x%X is not divisible by 4", len(offsetPayload))
	}

	count := len(offsetPayload) / 4
	out := make([]string, count)

	for i := 0; i < count; i++ {
		off := int(U32(offsetPayload, i*4))
		if off < 0 || off >= len(stringPayload) {
			log.Fatalf("UFP string offset out of range: index=%d offset=0x%X tableSize=0x%X", i, off, len(stringPayload))
		}

		size := int(stringPayload[off])
		start := off + 1
		end := start + size

		if end > len(stringPayload) {
			log.Fatalf("UFP string out of range: index=%d offset=0x%X size=0x%X tableSize=0x%X", i, off, size, len(stringPayload))
		}

		raw := stringPayload[start:end]
		decoded, err := encoding.FromSJIS(raw)
		if err != nil {
			return nil, fmt.Errorf("SJIS decode in UFP string table: %w", err)
		}
		out[i] = decoded
	}

	return out, nil
}

// parseOffsetTable reads a table of little-endian uint32 offsets.
func parseOffsetTable(payload []byte, name string) []uint32 {
	if len(payload)%4 != 0 {
		log.Fatalf("UFP %s size 0x%X is not divisible by 4", name, len(payload))
	}

	out := make([]uint32, len(payload)/4)
	for i := range out {
		out[i] = U32(payload, i*4)
	}
	return out
}

// buildPath assembles a file path from the string table, metadata flags,
// and string-table indices.
//
//	String table index 0 is always the archive root name.
//	The high 16 bits of meta optionally specify a parent directory index.
//	String IDs whose table entry starts with "." are treated as extensions
//	and appended to the previous path component.
func buildPath(stringsTable []string, meta uint32, ids []uint16) string {
	if len(stringsTable) == 0 {
		log.Fatalf("UFP string table is empty")
	}

	parts := []string{}

	// This archive's root path component is string 0, "2d".
	root := stringsTable[0]
	if root != "" {
		parts = append(parts, root)
	}

	// Some records have an extra parent/group directory in the high 16 bits.
	// In 2D.UFP this is usually 0, but sample/test files use index 301.
	parentIndex := int(meta >> 16)
	if parentIndex != 0 {
		if parentIndex >= len(stringsTable) {
			log.Fatalf("UFP parent string index out of range: index=%d strings=%d", parentIndex, len(stringsTable))
		}
		parts = append(parts, stringsTable[parentIndex])
	}

	for _, rawID := range ids {
		id := int(rawID)

		if id >= len(stringsTable) {
			log.Fatalf("UFP path string index out of range: index=%d strings=%d", id, len(stringsTable))
		}

		// Many records include the root string index as a trailing component.
		// We already added it at the front.
		if id == 0 {
			continue
		}

		part := stringsTable[id]
		if part == "" {
			continue
		}

		if strings.HasPrefix(part, ".") {
			if len(parts) == 0 {
				log.Fatalf("UFP extension %q has no filename before it", part)
			}
			parts[len(parts)-1] += part
		} else {
			parts = append(parts, part)
		}
	}

	return cleanPath(strings.Join(parts, "/"))
}
