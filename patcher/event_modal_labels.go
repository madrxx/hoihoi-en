// Event modal label replacement in event BSV files. Scans the SRTS string
// pool of event BSV tables for known Japanese modal labels (Yes/No/Cancel
// variants), replaces them with English text, and rewrites all u32 string
// references. If the replacement is longer than the original, the SRTS
// section is grown by shifting EMAN/DNIK/ATAD downward.
package patcher

import (
	"log"

	"github.com/madrxx/hoihoi-en/bsv"
	"github.com/madrxx/hoihoi-en/encoding"
	"github.com/madrxx/hoihoi-en/game"
	"github.com/madrxx/hoihoi-en/text"
)



func alignEventModalU64(value, alignment uint64) uint64 {
	if alignment == 0 {
		return value
	}
	if rem := value % alignment; rem != 0 {
		return value + alignment - rem
	}
	return value
}

func eventModalOpenBSV(p *Patcher, base uint64) BSVTable {
	t := BSVTable{
		FilePath: game.GameUFP,
		Base:     base,
	}

	t.SRTSOffset, t.SRTSSize = p.findBSVSection(t.FilePath, t.Base, "SRTS")
	t.EMANOffset, t.EMANSize = p.findBSVSection(t.FilePath, t.Base, "EMAN")
	t.DNIKOffset, t.DNIKSize = p.findBSVSection(t.FilePath, t.Base, "DNIK")
	t.ATADOffset, t.ATADSize = p.findBSVSection(t.FilePath, t.Base, "ATAD")

	return t
}

// Event BSVs have useful SRTS tail bytes up to EMAN, even where the declared
// SRTS size field appears slightly short. This effective end is used only for
// reading/clearing the string area, not for calculating the declared SRTS size.
func eventModalSRTSEnd(t BSVTable) uint64 {
	if t.EMANOffset > t.SRTSPoolStart() {
		return t.EMANOffset
	}
	return t.SRTSEnd()
}

// collectEventModalSRTSStrings scans the SRTS string pool byte-by-byte,
// collecting every null-terminated string and its base-relative offset.
// This is more conservative than reading via u32 references  -- it finds all
// strings regardless of whether they are currently referenced.
func (p *Patcher) collectEventModalSRTSStrings(t BSVTable) []text.EventModalSRTSString {
	out := []text.EventModalSRTSString{}

	pos := t.SRTSPoolStart()
	end := eventModalSRTSEnd(t)

	for pos < end {
		first := p.ReadFileBytes(t.FilePath, pos, 1)[0]
		if first == 0x00 {
			pos++
			continue
		}

		data := p.ReadFileBytes(t.FilePath, pos, end-pos)

		nul := -1
		for i, b := range data {
			if b == 0x00 {
				nul = i
				break
			}
		}

		if nul < 0 {
			log.Fatalf(
				"unterminated event SRTS string: table=0x%X pos=0x%X effectiveEnd=0x%X",
				t.Base,
				pos,
				end,
			)
		}

		raw := make([]byte, nul)
		copy(raw, data[:nul])

		out = append(out, text.EventModalSRTSString{
			Rel: uint32(pos - t.Base),
			Raw: raw,
		})

		pos += uint64(nul + 1)
	}

	return out
}

func eventModalStringStartSet(strings []text.EventModalSRTSString) map[uint32]bool {
	out := map[uint32]bool{}
	for _, s := range strings {
		out[s.Rel] = true
	}
	return out
}

// Scan aligned u32 fields only.
// Byte-by-byte scanning can corrupt event VM operands by finding accidental
// offset-looking values.
func (p *Patcher) collectEventModalRelRefs(t BSVTable, starts map[uint32]bool) map[uint32][]uint64 {
	refs := map[uint32][]uint64{}

	scanRange := func(start, end uint64) {
		for off := start; off+4 <= end; off += 4 {
			rel := p.ReadFileU32LE(t.FilePath, off)
			if starts[rel] {
				refs[rel] = append(refs[rel], off)
			}
		}
	}

	scanRange(t.EMANOffset+0x08, t.EMANOffset+0x08+t.EMANSize)
	scanRange(t.DNIKOffset+0x08, t.DNIKOffset+0x08+t.DNIKSize)
	scanRange(t.ATADOffset+0x08, t.ATADOffset+0x08+t.ATADSize)

	return refs
}

// Moves EMAN/DNIK/ATAD and following data downward by delta bytes.
//
// Important: the declared SRTS size is increased by delta, not recomputed from
// the new EMAN boundary. The original event BSVs commonly have padding between
// declared SRTS end and EMAN; preserving that pattern seems important.
func (p *Patcher) growEventModalSRTSBy(t BSVTable, delta uint64) {
	if delta == 0 {
		return
	}

	delta = alignEventModalU64(delta, 0x10)

	bsvSize := uint64(p.ReadFileU32LE(t.FilePath, t.Base+0x08))
	if bsvSize == 0 {
		log.Fatalf("event BSV size is zero: table=0x%X", t.Base)
	}

	bsvEnd := t.Base + bsvSize
	moveStart := t.EMANOffset
	moveEnd := t.ATADOffset + 0x08 + t.ATADSize

	if moveEnd < moveStart {
		log.Fatalf(
			"invalid event BSV section order: table=0x%X moveStart=0x%X moveEnd=0x%X",
			t.Base,
			moveStart,
			moveEnd,
		)
	}

	if moveEnd+delta > bsvEnd {
		log.Fatalf(
			"not enough event BSV tail space to grow SRTS: table=0x%X needDelta=0x%X moveEnd=0x%X bsvEnd=0x%X overBy=0x%X",
			t.Base,
			delta,
			moveEnd,
			bsvEnd,
			moveEnd+delta-bsvEnd,
		)
	}

	payload := p.ReadFileBytes(t.FilePath, moveStart, moveEnd-moveStart)

	p.WriteFileBytes(t.FilePath, moveStart+delta, payload)
	p.WriteFileBytes(t.FilePath, moveStart, make([]byte, delta))

	newSRTSSize := t.SRTSSize + delta
	if newSRTSSize > 0xFFFFFFFF {
		log.Fatalf("event SRTS size overflow: table=0x%X", t.Base)
	}

	p.WriteFileU32LE(t.FilePath, t.SRTSOffset+0x04, uint32(newSRTSSize))
}

// Rebuilds the existing SRTS area from referenced strings only, replacing one
// modal label. If required, it grows SRTS by 0x10 into the fixed BSV tail buffer.
// It does not use ASCII; NewText goes through EncodePatchText, i.e. fullwidth.
func (p *Patcher) patchOneEventModalLabel(patch text.EventModalPatch) {
	t := eventModalOpenBSV(p, patch.Base)

	strings := p.collectEventModalSRTSStrings(t)
	starts := eventModalStringStartSet(strings)
	refs := p.collectEventModalRelRefs(t, starts)

	oldRel := uint32(0)
	oldFound := false

	for _, s := range strings {
		decoded, err := encoding.FromSJIS(s.Raw)
		if err != nil {
			log.Fatal(err)
		}
		if decoded == patch.OldText {
			oldRel = s.Rel
			oldFound = true
			break
		}
	}

	if !oldFound {
		log.Fatalf(
			"event modal label not found: table=0x%X oldText=%q",
			patch.Base,
			patch.OldText,
		)
	}

	if len(refs[oldRel]) == 0 {
		log.Fatalf(
			"event modal label string found but never referenced: table=0x%X rel=0x%X old=%q new=%q",
			patch.Base,
			oldRel,
			patch.OldText,
			patch.NewText,
		)
	}

	newLabelRaw := bsv.EncodePatchText(patch.NewText)

	type keptString struct {
		OldRel uint32
		NewRel uint32
		Raw    []byte
	}

	buildKept := func(t BSVTable, strings []text.EventModalSRTSString, refs map[uint32][]uint64) ([]keptString, uint64) {
		kept := []keptString{}
		cursorRel := uint32(t.SRTSPoolStart() - t.Base)

		for _, s := range strings {
			// Drop strings that have no aligned refs.
			if len(refs[s.Rel]) == 0 {
				continue
			}

			raw := s.Raw
			if s.Rel == oldRel {
				raw = newLabelRaw
			}

			kept = append(kept, keptString{
				OldRel: s.Rel,
				NewRel: cursorRel,
				Raw:    raw,
			})

			cursorRel += uint32(len(raw) + 1)
		}

		newPoolLen := uint64(cursorRel) - uint64(t.SRTSPoolStart()-t.Base)
		return kept, newPoolLen
	}

	kept, newPoolLen := buildKept(t, strings, refs)
	oldPoolCap := eventModalSRTSEnd(t) - t.SRTSPoolStart()

	if newPoolLen > oldPoolCap {
		p.growEventModalSRTSBy(t, newPoolLen-oldPoolCap)

		// Re-open and re-collect references because EMAN/DNIK/ATAD moved.
		t = eventModalOpenBSV(p, patch.Base)
		strings = p.collectEventModalSRTSStrings(t)
		starts = eventModalStringStartSet(strings)
		refs = p.collectEventModalRelRefs(t, starts)

		kept, newPoolLen = buildKept(t, strings, refs)
		oldPoolCap = eventModalSRTSEnd(t) - t.SRTSPoolStart()
	}

	if newPoolLen > oldPoolCap {
		log.Fatalf(
			"event modal label cannot fit even after safe SRTS growth: table=0x%X newPool=0x%X cap=0x%X old=%q new=%q",
			patch.Base,
			newPoolLen,
			oldPoolCap,
			patch.OldText,
			patch.NewText,
		)
	}

	pool := make([]byte, oldPoolCap)

	for _, s := range kept {
		off := uint64(s.NewRel) - uint64(t.SRTSPoolStart()-t.Base)
		copy(pool[off:], s.Raw)
		// pool is zero-filled, so the terminator is already present
	}

	p.WriteFileBytes(t.FilePath, t.SRTSPoolStart(), pool)

	for _, s := range kept {
		if s.OldRel == s.NewRel {
			continue
		}

		for _, off := range refs[s.OldRel] {
			p.WriteFileU32LE(t.FilePath, off, s.NewRel)
		}
	}
}

// EventModalLabels applies all event modal label replacements to the disc.
func (p *Patcher) EventModalLabels() {
	for _, patch := range text.EventModals {
		p.patchOneEventModalLabel(patch)
	}
}
