package main

// BSV table I/O for the root package  -- type aliases, table opening, string
// reading, and EMAN/DNIK reference collection. All BSV operations go
// through the Patcher struct for sector-boundary-crossing I/O.

import (
	"log"

	"github.com/madrxx/hoihoi-en/bsv"
)

// Type aliases re-export bsv package types into the main package so call
// sites in other root-package files can use short names without importing bsv.
type (
	BSVTable      = bsv.Table
	BSVStringPool = bsv.StringPool
	BSVFieldRef   = bsv.FieldRef
)

// OpenBSVTable opens a BSV table within a UFP entry by scanning for SRTS,
// EMAN, DNIK, and ATAD section tags from the given base offset.
func (p *Patcher) OpenBSVTable(filePath string, base uint64) BSVTable {
	table := BSVTable{
		FilePath: filePath,
		Base:     base,
	}

	table.SRTSOffset, table.SRTSSize = p.findBSVSection(filePath, base, "SRTS")
	table.EMANOffset, table.EMANSize = p.findBSVSection(filePath, base, "EMAN")
	table.DNIKOffset, table.DNIKSize = p.findBSVSection(filePath, base, "DNIK")
	table.ATADOffset, table.ATADSize = p.findBSVSection(filePath, base, "ATAD")

	table.RowCount = p.ReadFileU32LE(filePath, table.ATADOffset+0x08)
	if table.RowCount == 0 {
		log.Fatalf("BSV table at 0x%X has zero rows", base)
	}

	table.RowSize = table.ATADSize / uint64(table.RowCount)
	table.ATADTrailerSize = table.ATADSize - table.RowSize*uint64(table.RowCount)

	if table.RowSize == 0 {
		log.Fatalf("BSV table at 0x%X has zero row size", base)
	}

	if table.RowSize%4 != 0 {
		log.Fatalf(
			"BSV table at 0x%X has non-u32-aligned row size 0x%X from ATAD size 0x%X and row count %d",
			base,
			table.RowSize,
			table.ATADSize,
			table.RowCount,
		)
	}

	return table
}

func (p *Patcher) findBSVSection(filePath string, base uint64, tag string) (uint64, uint64) {
	const maxSearch uint64 = 0x10000

	want := []byte(tag)
	data := p.ReadFileBytes(filePath, base, maxSearch)

	for i := 0; i+8 <= len(data); i++ {
		if string(data[i:i+4]) == string(want) {
			offset := base + uint64(i)
			size := p.ReadFileU32LE(filePath, offset+0x04)
			return offset, uint64(size)
		}
	}

	log.Fatalf("section %s not found in BSV at 0x%X", tag, base)
	return 0, 0
}

// ReadBSVString reads a null-terminated string from a BSV table's SRTS
// section, given a relative offset (from the table base).
func (p *Patcher) ReadBSVString(t BSVTable, relOffset uint32) []byte {
	if relOffset == 0 || relOffset == 0xFFFFFFFF {
		return nil
	}

	fileOffset := t.Base + uint64(relOffset)

	if fileOffset < t.SRTSPoolStart() || fileOffset >= t.SRTSEnd() {
		log.Fatalf(
			"BSV string offset out of SRTS range: table=0x%X rel=0x%X fileOffset=0x%X",
			t.Base,
			relOffset,
			fileOffset,
		)
	}

	maxLen := t.SRTSEnd() - fileOffset
	data := p.ReadFileBytes(t.FilePath, fileOffset, maxLen)

	for i, b := range data {
		if b == 0x00 {
			out := make([]byte, i)
			copy(out, data[:i])
			return out
		}
	}

	log.Fatalf("unterminated BSV string: table=0x%X rel=0x%X", t.Base, relOffset)
	return nil
}

// collectEMANRefs reads all u32 string references from the EMAN section
// of a BSV table and returns them with their resolved raw strings.
func (p *Patcher) collectEMANRefs(t BSVTable) []BSVFieldRef {
	count := p.ReadFileU32LE(t.FilePath, t.EMANOffset+0x08)
	refs := make([]BSVFieldRef, 0, count)

	for i := uint32(0); i < count; i++ {
		fieldOffset := t.EMANOffset + 0x10 + uint64(i)*8
		rel := p.ReadFileU32LE(t.FilePath, fieldOffset)
		raw := p.ReadBSVString(t, rel)

		refs = append(refs, BSVFieldRef{
			FileOffset: fieldOffset,
			Raw:        raw,
		})
	}

	return refs
}

// collectDNIKRefs scans the DNIK section for u32 values that look like
// string pool references (>= 0x30, pointing within SRTS) and returns them
// with their resolved raw strings.
func (p *Patcher) collectDNIKRefs(t BSVTable) []BSVFieldRef {
	refs := []BSVFieldRef{}

	start := t.DNIKOffset + 0x08
	end := start + t.DNIKSize

	for off := start; off+4 <= end; off += 4 {
		rel := p.ReadFileU32LE(t.FilePath, off)

		if rel < 0x30 {
			continue
		}

		fileOffset := t.Base + uint64(rel)
		if fileOffset < t.SRTSPoolStart() || fileOffset >= t.SRTSEnd() {
			continue
		}

		raw := p.ReadBSVString(t, rel)
		if len(raw) == 0 {
			continue
		}

		refs = append(refs, BSVFieldRef{
			FileOffset: off,
			Raw:        raw,
		})
	}

	return refs
}
