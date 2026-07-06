package patcher

// ADV cutscene subtitle injection. Generates MIPS machine code at build time
// that is written into a code cave in the game executable, constructing a 2D
// sprite overlay for English subtitles during voiced cutscenes. Also handles
// the binary subtitle lookup table and string pool in the cave.

import (
	"encoding/binary"
	"strings"

	"github.com/madrxx/hoihoi-en/cave"
	"github.com/madrxx/hoihoi-en/encoding"
	"github.com/madrxx/hoihoi-en/mips"
	"github.com/madrxx/hoihoi-en/text"
)

const maxSubtitleLines = 4


func splitSubtitleLines(text string) []string {
	if strings.TrimSpace(text) == "" {
		return []string{}
	}

	parts := strings.Split(text, "\n")
	lines := make([]string, 0, maxSubtitleLines)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			lines = append(lines, part)
		}
	}
	if len(lines) > maxSubtitleLines {
		last := strings.Join(lines[maxSubtitleLines-1:], " ")
		lines = append(lines[:maxSubtitleLines-1], last)
	}
	return lines
}

// WriteAdvSubtitleTable serialises the ADV subtitle lookup table and its
// string pool into the code cave. Each entry maps (eventIndex, voiceID) to
// up to 4 subtitle lines. The string pool follows the fixed-size table.
func (p *Patcher) WriteAdvSubtitleTable(file string, tableOffset uint64, caveEndOffset uint64, patches []text.AdvSubtitlePatch, emptyRuntime uint32) {
	const runtimeBase uint32 = 0x000FFF80
	const entrySize uint64 = 28 // event, voice, lineCount, line0, line1, line2, line3

	tableSize := entrySize * uint64(len(patches)+1)
	stringPoolOffset := tableOffset + tableSize
	if rem := stringPoolOffset % 4; rem != 0 {
		stringPoolOffset += 4 - rem
	}

	cursor := stringPoolOffset
	table := make([]byte, tableSize)
	for i, patch := range patches {
		lines := splitSubtitleLines(patch.Text)
		entryOffset := i * int(entrySize)
		binary.LittleEndian.PutUint32(table[entryOffset+0:entryOffset+4], patch.EventIndex)
		binary.LittleEndian.PutUint32(table[entryOffset+4:entryOffset+8], patch.VoiceID)
		binary.LittleEndian.PutUint32(table[entryOffset+8:entryOffset+12], uint32(len(lines)))

		for lineIndex := 0; lineIndex < maxSubtitleLines; lineIndex++ {
			ptr := emptyRuntime
			if lineIndex < len(lines) && lines[lineIndex] != "" {
				raw := encoding.GameText(lines[lineIndex])
				if len(raw) == 0 || raw[len(raw)-1] != 0 {
					raw = append(raw, 0)
				}
				if cursor+uint64(len(raw)) > caveEndOffset {
					panic("ADV subtitle table/string pool overflow")
				}
				ptr = uint32(cursor) + runtimeBase
				p.WriteFileBytes(file, cursor, raw)
				cursor += uint64(len(raw))
			}
			binary.LittleEndian.PutUint32(table[entryOffset+12+lineIndex*4:entryOffset+16+lineIndex*4], ptr)
		}
	}

	sentinelOffset := len(patches) * int(entrySize)
	binary.LittleEndian.PutUint32(table[sentinelOffset+0:sentinelOffset+4], 0xFFFFFFFF)
	binary.LittleEndian.PutUint32(table[sentinelOffset+4:sentinelOffset+8], 0xFFFFFFFF)
	binary.LittleEndian.PutUint32(table[sentinelOffset+8:sentinelOffset+12], 0)
	for i := 0; i < maxSubtitleLines; i++ {
		binary.LittleEndian.PutUint32(table[sentinelOffset+12+i*4:sentinelOffset+16+i*4], emptyRuntime)
	}
	p.WriteFileBytes(file, tableOffset, table)
	if cursor < caveEndOffset {
		p.WriteFileBytes(file, cursor, make([]byte, caveEndOffset-cursor))
	}
}

// buildTelopBgInitHelper generates MIPS code that constructs the subtitle
// background overlay. This is called once to initialise the 2D sprite
// hierarchy, load the telop PSI texture, and set geometry/colours.
//
// Sprite hierarchy (all allocated within the pre-cleared bgRoot region):
//
//	bgRoot (CJL2DPrimitiveTypeS)   -- container at bgRoot+0x000 (0xAC bytes)
//	  ├── bgFill (CJLPolygonTypeR)  -- black translucent fill at bgRoot+0x0AC
//	  ├── bgTop  (CJLSpriteTypeS)    -- top cap sprite at bgRoot+0x180
//	  └── bgBottom (CJLSpriteTypeS)   -- bottom cap sprite at bgRoot+0x254
//
// Key field offsets within CJL2DPrimitiveTypeS (container):
//
//	+0x0D  visible flag (byte)
//	+0x10  z-index (u32)
//	+0x90  x position (u32)
//	+0x94  y position (u32)
//	+0x98  width (u32)
//	+0x9C  height (u32)
//
// Key field offsets within CJLPolygonTypeR (fill, at root+0xAC):
//
//	+0x00   -- +0xAC from root = bgFill base
//	+0x10  priority (u32)         -- root+0xBC
//	+0x49  colour R (byte)        -- root+0xF5
//	+0x4A  colour G (byte)        -- root+0xF6
//	+0x4B  colour B (byte)        -- root+0xF7
//	+0x4C  colour A (byte)        -- root+0xF8
//	+0x0D  visible flag (byte)    -- root+0xB9
//	+0x90  left (u32)             -- root+0x13C
//	+0x94  top (u32)              -- root+0x140
//	+0x98  right (u32)            -- root+0x144
//	+0x9C  bottom (u32)           -- root+0x148
//
// Key field offsets within CJLSpriteTypeS (top cap at root+0x180, bottom at +0x254):
//
//	+0x00   -- cap base
//	+0x0D  visible flag (byte)    -- root+0x18D / root+0x261
//	+0x10  z-index (u32)          -- root+0x190 / root+0x264
//	+0x4C  alpha (byte)           -- root+0x1CC / root+0x2A0 (via CJLSpriteInterface at +0xAC)
//	+0x90  x position (u32)       -- root+0x210 / root+0x2E4
//	+0x94  y position (u32)       -- root+0x214 / root+0x2E8
//	+0x98  width (u32)            -- root+0x218 / root+0x2EC
//	+0x9C  height (u32)           -- root+0x21C / root+0x2F0
//	+0xA0  scale X (u32, float)   -- root+0x220 / root+0x2F4
//	+0xA4  scale Y (u32, float)   -- root+0x224 / root+0x2F8
//
// The telop PSI texture "sys_000: 8x32 telop/ticker background strip"
// (at runtime string 0x0026DAD7) is loaded via the psiManager and
// assigned to both cap sprites via setName.
func buildTelopBgInitHelper(bgRoot uint32) []uint32 {
	const (
		ctor2DPrimitiveTypeS = 0x0020C3B0
		ctorPolygonTypeR     = 0x00210220
		ctorSpriteTypeS      = 0x001597F0
		findByFileName       = 0x00216630
		loadPsi              = 0x002179C0
		createPrim           = 0x002118B0
		setParent            = 0x00211740
		setName              = 0x00218260

		psiManager  = 0x0038EF70
		telopPsi    = 0x0026DAC0
		telopSprite = 0x0026DAD7 // sys_000: 8x32 telop/ticker background strip
	)

	bgFill := bgRoot + 0xAC
	bgTop := bgRoot + 0x180
	bgTopIface := bgTop + 0xAC
	bgBottom := bgRoot + 0x254
	bgBottomIface := bgBottom + 0xAC

	buf := mips.NewCodeBuf()
	buf.Add(mips.Addiu(29, 29, -0x20), 0xFFBF0010)

	buf.LoadAddr(4, bgRoot)
	buf.Add(mips.Jal(ctor2DPrimitiveTypeS), mips.Addiu(5, 0, 0))

	buf.LoadAddr(4, bgFill)
	buf.Add(
		mips.Addiu(5, 0, 1),
		mips.Jal(ctorPolygonTypeR),
		mips.Daddu(6, 0, 0),
	)

	for _, child := range []uint32{bgTop, bgBottom} {
		buf.LoadAddr(4, child)
		buf.Add(mips.Jal(ctorSpriteTypeS), mips.Addiu(5, 0, 0))
	}

	buf.LoadAddr(5, telopPsi)
	buf.LoadAddr(4, psiManager)
	buf.Add(mips.Jal(findByFileName), 0x00000000)
	skipLoad := buf.AddPlaceholder()
	buf.Add(0x00000000) // delay slot for the placeholder Bne
	buf.Add(0x8F868014) // lw a2,-32748(gp) ; glp2DUfp
	buf.LoadAddr(4, psiManager)
	buf.LoadAddr(5, telopPsi)
	buf.Add(mips.Daddu(7, 0, 0), mips.Jal(loadPsi), mips.Daddu(8, 0, 0))
	buf.ResolveBne(skipLoad, 2, 0, buf.Here())

	buf.LoadAddr(4, bgRoot)
	buf.Add(mips.Jal(createPrim), 0x00000000)
	buf.LoadAddr(3, bgRoot)
	buf.Add(
		mips.Sw(0, mips.RootPosX, 3),
		mips.Addiu(2, 0, 0x98),
		mips.Sw(2, mips.RootPosY, 3),
		mips.Addiu(2, 0, 0x200),
		mips.Sw(2, mips.RootWidth, 3),
		mips.Addiu(2, 0, 0x44),
		mips.Sw(2, mips.RootHeight, 3),
		mips.Addiu(2, 0, 10096),
		mips.Sw(2, mips.RootZIndex, 3),
		mips.Sb(0, mips.RootVisible, 3),
	)

	buf.LoadAddr(4, bgFill)
	buf.LoadAddr(5, bgRoot)
	buf.Add(mips.Jal(setParent), 0x00000000)
	buf.LoadAddr(4, bgFill)
	buf.Add(mips.Jal(createPrim), 0x00000000)

	for _, pair := range []struct{ child, iface uint32 }{{bgTop, bgTopIface}, {bgBottom, bgBottomIface}} {
		buf.LoadAddr(4, pair.child)
		buf.LoadAddr(5, bgRoot)
		buf.Add(mips.Jal(setParent), 0x00000000)
		buf.LoadAddr(4, pair.child)
		buf.Add(mips.Jal(createPrim), 0x00000000)
		buf.LoadAddr(4, pair.iface)
		buf.LoadAddr(5, telopSprite)
		buf.Add(mips.Jal(setName), mips.Addiu(6, 0, 1))
	}

	buf.LoadAddr(3, bgRoot)
	appendTelopBgGeometry(buf)

	buf.Add(0xDFBF0010, mips.Addiu(29, 29, 0x20), 0x03E00008, 0x00000000)
	return buf.Words()
}

// appendTelopBgGeometry emits MIPS instructions that initialise the subtitle
// background sprite hierarchy: a translucent black fill rectangle and two cap
// sprites (top/bottom) positioned relative to the root container ($3).
func appendTelopBgGeometry(buf *mips.CodeBuf) {
	buf.Add(
		// Root/container.
		mips.Sw(0, mips.RootPosX, 3),
		mips.Addiu(2, 0, 0x98),
		mips.Sw(2, mips.RootPosY, 3),
		mips.Addiu(2, 0, 0x200),
		mips.Sw(2, mips.RootWidth, 3),
		mips.Addiu(2, 0, 0x44),
		mips.Sw(2, mips.RootHeight, 3),

		// Polygon fill at bgRoot + 0xAC.
		mips.Addiu(2, 0, 10097),
		mips.Sw(2, mips.FillPriority, 3), // fill priority
		mips.Sb(0, mips.FillColorR, 3),   // fill R
		mips.Sb(0, mips.FillColorG, 3),   // fill G
		mips.Sb(0, mips.FillColorB, 3),   // fill B
		mips.Addiu(2, 0, 0x50),
		mips.Sb(2, mips.FillAlpha, 3), // fill alpha
		mips.Sw(0, mips.FillLeft, 3),  // fill left
		mips.Addiu(2, 0, 0x07),
		mips.Sw(2, mips.FillTop, 3), // fill top
		mips.Addiu(2, 0, 0x200),
		mips.Sw(2, mips.FillRight, 3), // fill right / width-ish
		mips.Addiu(2, 0, 0x3D),
		mips.Sw(2, mips.FillBottom, 3),  // fill bottom / height-ish
		mips.Sb(0, mips.FillVisible, 3), // fill hidden initially

		// Top cap sprite at bgRoot + 0x180.
		mips.Sw(0, mips.TopCapPosX, 3),
		mips.Sw(0, mips.TopCapPosY, 3),
		mips.Addiu(2, 0, 0x08),
		mips.Sw(2, mips.TopCapWidth, 3),
		mips.Addiu(2, 0, 0x07),
		mips.Sw(2, mips.TopCapHeight, 3),
		mips.Lui(2, 0x4280), // scale X = 64.0f
		mips.Sw(2, mips.TopCapScaleX, 3),
		mips.Lui(2, 0x3F80), // scale Y = 1.0f
		mips.Sw(2, mips.TopCapScaleY, 3),
		mips.Addiu(2, 0, 10098),
		mips.Sw(2, mips.TopCapZIndex, 3),
		mips.Sb(0, mips.TopCapVisible, 3),
		mips.Sb(0, mips.TopCapAlpha, 3),

		// Bottom cap sprite at bgRoot + 0x254.
		mips.Sw(0, mips.BotCapPosX, 3),
		mips.Addiu(2, 0, 0x3D),
		mips.Sw(2, mips.BotCapPosY, 3),
		mips.Addiu(2, 0, 0x08),
		mips.Sw(2, mips.BotCapWidth, 3),
		mips.Addiu(2, 0, 0x07),
		mips.Sw(2, mips.BotCapHeight, 3),
		mips.Lui(2, 0x4280),
		mips.Sw(2, mips.BotCapScaleX, 3),
		mips.Lui(2, 0x3F80),
		mips.Sw(2, mips.BotCapScaleY, 3),
		mips.Addiu(2, 0, 10098),
		mips.Sw(2, mips.BotCapZIndex, 3),
		mips.Sb(0, mips.BotCapVisible, 3),
		mips.Sb(0, mips.BotCapAlpha, 3),
	)
}

// buildShowHelper generates the MIPS stub that handles voice-line dispatch
// for ADV cutscenes. It scans the subtitle table for a matching (eventIndex,
// voiceID) entry, formats hex nibbles into a debug buffer when no match is
// found, and conditionally shows or hides the subtitle window and background.
func buildShowHelper(subtitleTable, debugBuffer, hexTable uint32, subtitleStrings [maxSubtitleLines]uint32, emptyString, bgRoot uint32, textY int32, subtitleEntryCount int) []uint32 {
	const (
		setString        = 0x00218E50
		continueSetVoice = 0x001F4898
	)
	if subtitleEntryCount < 1 || subtitleEntryCount > 0x7FFF {
		panic("ADV subtitle entry count out of range")
	}

	buf := mips.NewCodeBuf()
	buf.Add(mips.Addiu(29, 29, -0x30), 0xFFBF0010, mips.Lw(8, 0x18, 16), mips.Lw(9, 0x38, 16))
	buf.LoadAddr(10, subtitleTable)
	buf.Add(mips.Addiu(3, 0, -1), mips.Addiu(14, 0, int32(subtitleEntryCount)))

	// Loop top: check count exhausted, check sentinel, check EventIndex/VoiceID.
	countExhausted := buf.AddPlaceholder()
	buf.Add(0x00000000) // delay slot for count-exhausted Beq
	loopBody := buf.Here() // first real instruction of loop body, also the loop back-edge target
	buf.Add(mips.Lw(11, 0, 10), mips.Lw(12, 4, 10))
	beqSentinel := buf.AddPlaceholder()
	buf.Add(0x00000000)
	bneEvent := buf.AddPlaceholder()
	buf.Add(0x00000000)
	bneVoice := buf.AddPlaceholder()
	buf.Add(0x00000000)

	// Match: check lineCount.
	buf.Add(mips.Lw(15, 8, 10))
	manualEmptyBranch := buf.AddPlaceholder()
	buf.Add(0x00000000)

	// Non-empty match: load up to 4 line pointers.
	buf.Add(
		mips.Lw(11, 12, 10),
		mips.Lw(12, 16, 10),
		mips.Lw(13, 20, 10),
		mips.Lw(14, 24, 10),
		mips.Addiu(15, 0, 1), // show background flag
	)
	jumpToSet := buf.AddPlaceholder()
	buf.Add(0x00000000)

	// Manually empty match: use empty strings, hide background.
	manualEmptyIndex := buf.Here()
	buf.LoadAddr(11, emptyString)
	buf.LoadAddr(12, emptyString)
	buf.LoadAddr(13, emptyString)
	buf.LoadAddr(14, emptyString)
	buf.Add(mips.Addiu(15, 0, 0)) // show background flag = false
	manualEmptyJump := buf.AddPlaceholder()
	buf.Add(0x00000000)

	// Advance to next entry and loop back.
	nextIndex := buf.Here()
	buf.Add(mips.Addiu(10, 10, 28), mips.Addiu(14, 14, -1))
	loopBranch := buf.AddPlaceholder()
	buf.Add(0x00000000)
	buf.ResolveBeq(loopBranch, 0, 0, loopBody)

	// No-match / debug fallback.
	noMatchIndex := buf.Here()
	buf.LoadAddr(14, debugBuffer)
	buf.LoadAddr(15, hexTable)

	emitHalfword := func(hw uint32, off int) { buf.Add(mips.Ori(2, 0, hw), mips.Sh(2, off, 14)) }
	emitHexNibble := func(srcReg int, shift int, off int) {
		if shift != 0 {
			buf.Add(mips.Srl(2, srcReg, shift))
		} else {
			buf.Add(mips.Addu(2, srcReg, 0))
		}
		buf.Add(mips.Andi(2, 2, 0xF), mips.Sll(2, 2, 1), mips.Addu(2, 15, 2), mips.Lhu(2, 0, 2), mips.Sh(2, off, 14))
	}

	emitHalfword(0x6482, 0)
	emitHalfword(0x7582, 2)
	emitHalfword(0x4681, 4)
	emitHexNibble(8, 12, 6)
	emitHexNibble(8, 8, 8)
	emitHexNibble(8, 4, 10)
	emitHexNibble(8, 0, 12)
	emitHalfword(0x4081, 14)
	emitHalfword(0x7582, 16)
	emitHalfword(0x6E82, 18)
	emitHalfword(0x4681, 20)
	emitHexNibble(9, 12, 22)
	emitHexNibble(9, 8, 24)
	emitHexNibble(9, 4, 26)
	emitHexNibble(9, 0, 28)
	buf.Add(mips.Sb(0, 30, 14))
	buf.LoadAddr(11, debugBuffer)
	buf.LoadAddr(12, emptyString)
	buf.LoadAddr(13, emptyString)
	buf.LoadAddr(14, emptyString)
	buf.Add(mips.Addiu(15, 0, 1)) // debug text should still show background

	doSetIndex := buf.Here()
	buf.Add(
		mips.Sw(11, 0, 29),
		mips.Sw(12, 4, 29),
		mips.Sw(13, 8, 29),
		mips.Sw(14, 12, 29),
		mips.Sw(15, 16, 29), // show background flag
	)

	// Resolve all forward branches from the loop.
	buf.ResolveBeq(countExhausted, 14, 0, noMatchIndex)
	buf.ResolveBeq(beqSentinel, 11, 3, noMatchIndex)
	buf.ResolveBne(bneEvent, 11, 8, nextIndex)
	buf.ResolveBne(bneVoice, 12, 9, nextIndex)
	buf.ResolveBeq(manualEmptyBranch, 15, 0, manualEmptyIndex)
	buf.ResolveBeq(jumpToSet, 0, 0, doSetIndex)
	buf.ResolveBeq(manualEmptyJump, 0, 0, doSetIndex)

	// Conditionally show/hide background.
	buf.Add(mips.Lw(2, 16, 29))
	hideBgBranch := buf.AddPlaceholder()
	buf.Add(0x00000000)

	// Show background.
	buf.LoadAddr(3, bgRoot)
	appendTelopBgGeometry(buf)
	buf.Add(
		mips.Addiu(2, 0, 0x50),
		mips.Sb(2, mips.FillAlpha, 3),
		mips.Sb(2, mips.TopCapAlpha, 3),
		mips.Sb(2, mips.BotCapAlpha, 3),
		mips.Addiu(2, 0, 1),
		mips.Sb(2, mips.RootVisible, 3),
		mips.Sb(2, mips.FillVisible, 3),
		mips.Sb(2, mips.TopCapVisible, 3),
		mips.Sb(2, mips.BotCapVisible, 3),
	)
	afterBgJump := buf.AddPlaceholder()
	buf.Add(0x00000000)

	// Hide background.
	hideBgIndex := buf.Here()
	buf.LoadAddr(3, bgRoot)
	buf.Add(
		mips.Sb(0, mips.RootVisible, 3),
		mips.Sb(0, mips.FillVisible, 3), mips.Sb(0, mips.FillAlpha, 3),
		mips.Sb(0, mips.TopCapVisible, 3), mips.Sb(0, mips.TopCapAlpha, 3),
		mips.Sb(0, mips.BotCapVisible, 3), mips.Sb(0, mips.BotCapAlpha, 3),
	)
	afterBgIndex := buf.Here()
	buf.ResolveBeq(hideBgBranch, 2, 0, hideBgIndex)
	buf.ResolveBeq(afterBgJump, 0, 0, afterBgIndex)

	lineYs := []int32{textY, textY + 0x0C, textY + 0x18, textY + 0x24}
	for i := 0; i < maxSubtitleLines; i++ {
		buf.LoadAddr(4, subtitleStrings[i])
		buf.Add(mips.Lw(5, i*4, 29), mips.Jal(setString), 0x00000000)
		buf.LoadAddr(3, subtitleStrings[i])
		buf.Add(mips.Addiu(2, 0, 0x10), mips.Sw(2, mips.RootPosX, 3), mips.Addiu(2, 0, lineYs[i]), mips.Sw(2, mips.RootPosY, 3), mips.Addiu(2, 0, 10099), mips.Sw(2, mips.RootZIndex, 3), mips.Addiu(2, 0, 1), mips.Sb(2, mips.RootVisible, 3))
	}

	buf.Add(0xDFBF0010, mips.Addiu(29, 29, 0x30), mips.J(continueSetVoice), 0x00000000)
	return buf.Words()
}

// buildHideHelper generates MIPS code that clears all subtitle text and
// hides the background overlay. It is called when a voice line ends or
// is interrupted.
//
// For each of the 4 subtitle string sprites: call setString with the
// empty string, then set the sprite's visible flag (offset +0x0D) to 0.
//
// For the background root: clear visible flags on root (+0x0D), fill
// (+0xB9, +0xF8 for alpha), top cap (+0x18D, +0x1CC for alpha), and
// bottom cap (+0x261, +0x2A0 for alpha). Finally restore two global
// fields and jump to the epilogue hook (setVoiceEpilogue at 0x001F48C8).
func buildHideHelper(subtitleStrings [maxSubtitleLines]uint32, emptyString, bgRoot uint32) []uint32 {
	const (
		setString        = 0x00218E50
		setVoiceEpilogue = 0x001F48C8
	)
	buf := mips.NewCodeBuf(); buf.Add(mips.Addiu(29, 29, -0x20), 0xFFBF0010)
	for i := 0; i < maxSubtitleLines; i++ {
		buf.LoadAddr( 4, subtitleStrings[i])
		buf.LoadAddr( 5, emptyString)
		buf.Add(mips.Jal(setString), 0x00000000)
		buf.LoadAddr( 3, subtitleStrings[i])
		buf.Add(mips.Sb(0, mips.RootVisible, 3))
	}
	buf.LoadAddr(3, bgRoot)
	buf.Add(
		mips.Sb(0, mips.RootVisible, 3),
		mips.Sb(0, mips.FillVisible, 3), mips.Sb(0, mips.FillAlpha, 3),
		mips.Sb(0, mips.TopCapVisible, 3), mips.Sb(0, mips.TopCapAlpha, 3),
		mips.Sb(0, mips.BotCapVisible, 3), mips.Sb(0, mips.BotCapAlpha, 3),
		0xAE000038,
		0xAE000034,
		mips.Addiu(2, 0, 1),
		0xDFBF0010,
		mips.Addiu(29, 29, 0x20),
		mips.J(setVoiceEpilogue),
		0x00000000,
	)
	return buf.Words()
}

// buildAdvDoneCleanupHelper generates MIPS code that runs when an ADV
// cutscene ends. It clears all 4 subtitle string sprites, hides the
// background overlay (root, fill, top cap, bottom cap), and jumps to the
// original done__22CHoiSubScene_Event_AdvFv continuation at 0x001F4F78.
//
// The original $a0 (CHoiSubScene *this) is preserved on the stack at
// sp+0x14 so it can be restored before the tail jump.
func buildAdvDoneCleanupHelper(subtitleStrings [maxSubtitleLines]uint32, emptyString, bgRoot uint32) []uint32 {
	const (
		setString    = 0x00218E50
		doneContinue = 0x001F4F78
	)

	buf := mips.NewCodeBuf()
	buf.Add(
		// Original instructions overwritten at done__22CHoiSubScene_Event_AdvFv.
		mips.Addiu(29, 29, -0x20),
		0xFFBF0010, // sd ra, 0x10(sp)

		// Preserve original a0, i.e. CHoiSubScene_Event_Adv *this.
		mips.Sw(4, 0x14, 29),
	)

	// Clear all subtitle strings and hide the string sprites.
	for i := 0; i < maxSubtitleLines; i++ {
		buf.LoadAddr(4, subtitleStrings[i])
		buf.LoadAddr(5, emptyString)
		buf.Add(mips.Jal(setString), 0x00000000)
		buf.LoadAddr(3, subtitleStrings[i])
		buf.Add(mips.Sb(0, mips.RootVisible, 3))
	}

	// Hide the injected subtitle background:
	// root, polygon fill, top cap, bottom cap.
	buf.LoadAddr(3, bgRoot)
	buf.Add(
		mips.Sb(0, mips.RootVisible, 3), // root hidden
		mips.Sb(0, mips.FillVisible, 3), mips.Sb(0, mips.FillAlpha, 3),
		mips.Sb(0, mips.TopCapVisible, 3), mips.Sb(0, mips.TopCapAlpha, 3),
		mips.Sb(0, mips.BotCapVisible, 3), mips.Sb(0, mips.BotCapAlpha, 3),

		// Restore original a0 before returning to done__22CHoiSubScene_Event_AdvFv.
		mips.Lw(4, 0x14, 29),
		mips.J(doneContinue),
		0x00000000,
	)

	return buf.Words()
}

// appendStringSpriteInit emits MIPS code to initialise a single subtitle
// line sprite: allocate a CJLStringSprite via 0x00219570, create
// the primitive via 0x002118B0, set colour to 0x808080 (grey) at offsets
// +0x49/+0x4A/+0x4B, configure alignment at 0x00219440, set the initial
// (empty) string via setString (0x00218E50), and position the sprite
// at the given textY coordinate.
func appendStringSpriteInit(buf *mips.CodeBuf, subtitleStringRuntime uint32, emptyRuntime uint32, textY int32) {
	buf.LoadAddr(4, subtitleStringRuntime)
	buf.Add(mips.Jal(0x00219570), mips.Addiu(5, 0, 0))
	buf.LoadAddr(4, subtitleStringRuntime)
	buf.Add(mips.Jal(0x002118B0), 0x00000000)
	buf.LoadAddr(3, subtitleStringRuntime)
	buf.Add(mips.Addiu(2, 0, 0x80), mips.Sb(2, mips.StringColorR, 3), mips.Sb(2, mips.StringColorG, 3), mips.Sb(2, mips.StringColorB, 3))
	buf.LoadAddr(4, subtitleStringRuntime)
	buf.Add(mips.Addiu(5, 0, 0x24), mips.Addiu(6, 0, 1), mips.Jal(0x00219440), mips.Addiu(7, 0, 0))
	buf.LoadAddr(4, subtitleStringRuntime)
	buf.LoadAddr(5, emptyRuntime)
	buf.Add(mips.Jal(0x00218E50), 0x00000000)
	buf.LoadAddr(3, subtitleStringRuntime)
	buf.Add(mips.Addiu(2, 0, 10099), mips.Sw(2, mips.RootZIndex, 3), mips.Addiu(2, 0, 0x10), mips.Sw(2, mips.RootPosX, 3), mips.Addiu(2, 0, textY), mips.Sw(2, mips.RootPosY, 3), mips.Sb(0, mips.RootVisible, 3))
}

// text.AdvSubtitlePatch injects a complete English subtitle rendering
// system into the game's ADV cutscene engine.
//
// Memory layout of the code cave (0x00159844 - 0x00163400 in SLPM_623.91):
//
//	0x00159844  init helper (MIPS code)       -- one-time initialisation stub
//	0x0015A044  window object (scratch RAM)     -- reserve 0x700 bytes
//	0x0015A744  string0 object (scratch RAM)    -- subtitle line 0 sprite
//	0x0015A944  string1 object (scratch RAM)    -- subtitle line 1 sprite
//	0x0015AB44  string2 object (scratch RAM)    -- subtitle line 2 sprite
//	0x0015AD44  string3 object (scratch RAM)    -- subtitle line 3 sprite
//	0x0015AF44  init flag (u32)                -- 0 = uninitialised, 1 = ready
//	0x0015AF48  empty string (1 null byte)      -- shared by all blank lines
//	0x0015AF80  debug buffer (128 bytes)        -- "debugevent/NNvoice/NN\0"
//	0x0015B000  show helper (MIPS code)         -- voice-line dispatch handler
//	0x0015C000  hex lookup table                -- SJIS "0123456789ABCDEF"
//	0x0015C080  bg root (scratch RAM)           -- sprite hierarchy container
//	0x0015C500  bg init helper (MIPS code)      -- background construction
//	0x0015CA00  hide helper (MIPS code)         -- clear text, hide background
//	0x0015CC00  done cleanup helper (MIPS code)  -- cutscene-end teardown
//	0x0015CE00  subtitle lookup table            -- binary table + string pool
//	0x00163400  cave end
//
// Hook points patched in the game executable:
//
//	0x000F4FD0  init hook                   -- calls init helper once
//	0x000F48E4  voice dispatch (path A)     -- calls show helper
//	0x000F4914  voice dispatch (path B)     -- calls show helper
//	0x000F492C  voice end                   -- calls hide helper
//	0x000F4FF0  ADV scene done              -- calls done cleanup helper
func (p *Patcher) AdvSubtitlePatch() {
	const file = "SLPM_623.91"
	const subtitleTextY int32 = 0xA2

	layout := cave.AdvSubtitleLayout
	layout.Verify()

	initHelperOff := layout.Region("initHelper").Offset
	windowOff := layout.Region("window").Offset
	string0Off := layout.Region("string0").Offset
	string1Off := layout.Region("string1").Offset
	string2Off := layout.Region("string2").Offset
	string3Off := layout.Region("string3").Offset
	flagOff := layout.Region("flag").Offset
	emptyOff := layout.Region("empty").Offset
	debugBufferOff := layout.Region("debugBuffer").Offset
	showHelperOff := layout.Region("showHelper").Offset
	hexTableOff := layout.Region("hexTable").Offset
	bgRootOff := layout.Region("bgRoot").Offset
	bgInitHelperOff := layout.Region("bgInitHelper").Offset
	hideHelperOff := layout.Region("hideHelper").Offset
	doneCleanupHelperOff := layout.Region("doneCleanupHelper").Offset
	subtitleTableOff := layout.Region("subtitleTable").Offset
	caveEndOff := layout.End

	subtitleStrings := [maxSubtitleLines]uint32{mips.RuntimeAddr(string0Off), mips.RuntimeAddr(string1Off), mips.RuntimeAddr(string2Off), mips.RuntimeAddr(string3Off)}
	emptyRuntime := mips.RuntimeAddr(emptyOff)
	debugBufferRuntime := mips.RuntimeAddr(debugBufferOff)
	hexTableRuntime := mips.RuntimeAddr(hexTableOff)
	bgRootRuntime := mips.RuntimeAddr(bgRootOff)
	bgInitRuntime := mips.RuntimeAddr(bgInitHelperOff)
	subtitleTableRuntime := mips.RuntimeAddr(subtitleTableOff)
	showRuntime := mips.RuntimeAddr(showHelperOff)
	hideRuntime := mips.RuntimeAddr(hideHelperOff)
	doneCleanupRuntime := mips.RuntimeAddr(doneCleanupHelperOff)

	p.WriteFileBytes(file, windowOff, make([]byte, string0Off-windowOff))
	p.WriteFileBytes(file, string0Off, make([]byte, layout.Region("string0").Size))
	p.WriteFileBytes(file, string1Off, make([]byte, layout.Region("string1").Size))
	p.WriteFileBytes(file, string2Off, make([]byte, layout.Region("string2").Size))
	p.WriteFileBytes(file, string3Off, make([]byte, layout.Region("string3").Size))
	p.WriteFileU32LE(file, flagOff, 0)
	p.WriteFileBytes(file, emptyOff, []byte{0})
	p.WriteFileBytes(file, debugBufferOff, make([]byte, layout.Region("debugBuffer").Size))
	p.WriteFileBytes(file, bgRootOff, make([]byte, layout.Region("bgRoot").Size))
	p.WriteFileBytes(file, hexTableOff, []byte{0x82, 0x4F, 0x82, 0x50, 0x82, 0x51, 0x82, 0x52, 0x82, 0x53, 0x82, 0x54, 0x82, 0x55, 0x82, 0x56, 0x82, 0x57, 0x82, 0x58, 0x82, 0x60, 0x82, 0x61, 0x82, 0x62, 0x82, 0x63, 0x82, 0x64, 0x82, 0x65})
	p.WriteAdvSubtitleTable(file, subtitleTableOff, caveEndOff, text.AdvSubtitles, emptyRuntime)

	initBuf := mips.NewCodeBuf()
	initBuf.Add(mips.Addiu(29, 29, -0x20), 0xFFBF0010)
	initBuf.LoadAddr(3, mips.RuntimeAddr(flagOff))
	initBuf.Add(mips.Sw(0, 0, 3), 0x00000000, 0x00000000)
	initBuf.LoadAddr(4, mips.RuntimeAddr(windowOff))
	initBuf.Add(mips.Jal(0x00159750), 0x00000000)
	initBuf.LoadAddr(4, mips.RuntimeAddr(windowOff))
	initBuf.Add(mips.Addiu(5, 0, 0x12), mips.Jal(0x00159490), mips.Addiu(6, 0, 3))
	appendStringSpriteInit(initBuf, subtitleStrings[0], emptyRuntime, subtitleTextY)
	appendStringSpriteInit(initBuf, subtitleStrings[1], emptyRuntime, subtitleTextY+0x0C)
	appendStringSpriteInit(initBuf, subtitleStrings[2], emptyRuntime, subtitleTextY+0x18)
	appendStringSpriteInit(initBuf, subtitleStrings[3], emptyRuntime, subtitleTextY+0x24)
	initBuf.Add(mips.Jal(bgInitRuntime), 0x00000000)
	initBuf.LoadAddr(3, mips.RuntimeAddr(flagOff))
	initBuf.Add(mips.Addiu(2, 0, 1), mips.Sw(2, 0, 3), 0xDFBF0010, mips.Addiu(29, 29, 0x20), 0xAE200CCC, mips.J(0x001F4F58), 0x00000000)
	initHelper := initBuf.Words()

	bgInitHelper := buildTelopBgInitHelper(bgRootRuntime)
		showHelper := buildShowHelper(subtitleTableRuntime, debugBufferRuntime, hexTableRuntime, subtitleStrings, emptyRuntime, bgRootRuntime, subtitleTextY, len(text.AdvSubtitles)+1)
	hideHelper := buildHideHelper(subtitleStrings, emptyRuntime, bgRootRuntime)
	doneCleanupHelper := buildAdvDoneCleanupHelper(subtitleStrings, emptyRuntime, bgRootRuntime)

	// Verify generated code fits within its allocated regions.
	if len(initHelper)*4 > int(layout.Region("initHelper").Size) {
		panic("ADV subtitle init helper overflows its cave region")
	}
	if len(showHelper)*4 > int(layout.Region("showHelper").Size) {
		panic("ADV subtitle show helper overflows its cave region")
	}
	if len(bgInitHelper)*4 > int(layout.Region("bgInitHelper").Size) {
		panic("ADV subtitle bg init helper overflows its cave region")
	}
	if len(hideHelper)*4 > int(layout.Region("hideHelper").Size) {
		panic("ADV subtitle hide helper overflows its cave region")
	}
	if len(doneCleanupHelper)*4 > int(layout.Region("doneCleanupHelper").Size) {
		panic("ADV subtitle done cleanup helper overflows its cave region")
	}

	p.WriteFileBytes(file, initHelperOff, mips.WriteU32sLE(initHelper))
	p.WriteFileBytes(file, bgInitHelperOff, mips.WriteU32sLE(bgInitHelper))
	p.WriteFileBytes(file, showHelperOff, mips.WriteU32sLE(showHelper))
	p.WriteFileBytes(file, hideHelperOff, mips.WriteU32sLE(hideHelper))
	p.WriteFileBytes(file, doneCleanupHelperOff, mips.WriteU32sLE(doneCleanupHelper))

	hooks := cave.AdvHookPoints
	p.WriteFileU32LE(file, hooks.Init, mips.J(0x002597C4))
	p.WriteFileU32LE(file, hooks.VoicePathA, mips.J(showRuntime))
	p.WriteFileU32LE(file, hooks.VoicePathB, mips.J(showRuntime))
	p.WriteFileU32LE(file, hooks.VoiceEnd, mips.J(hideRuntime))
	p.WriteFileU32LE(file, hooks.AdvDone, mips.J(doneCleanupRuntime))
	p.WriteFileU32LE(file, hooks.AdvDone+4, 0x00000000)
}
