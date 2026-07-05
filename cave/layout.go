// Package cave models the code cave layout within the game executable.
// The cave is an unused padding region in SLPM_623.91 that the patcher
// repurposes for injected MIPS code and scratch RAM.
//
// The Layout struct defines all regions and verifies they do not overlap.
package cave

import "fmt"

// Region describes a single allocation within the code cave.
type Region struct {
	Name   string // human-readable label
	Offset uint64 // file offset within SLPM_623.91
	Size   uint64 // allocation size in bytes
}

// Layout defines the full code cave with all its regions.
type Layout struct {
	Start   uint64
	End     uint64
	Regions []Region
}

// Region returns the Region with the given name, or panics if not found.
func (l Layout) Region(name string) Region {
	for _, r := range l.Regions {
		if r.Name == name {
			return r
		}
	}
	panic(fmt.Sprintf("cave region %q not found", name))
}

// Verify checks that no regions overlap and all fit within cave bounds.
// It panics on any violation.
func (l Layout) Verify() {
	// Check each region is within bounds.
	for _, r := range l.Regions {
		if r.Offset < l.Start {
			panic(fmt.Sprintf(
				"cave region %q offset 0x%X is before cave start 0x%X",
				r.Name, r.Offset, l.Start,
			))
		}
		if r.Offset+r.Size > l.End {
			panic(fmt.Sprintf(
				"cave region %q ends at 0x%X, beyond cave end 0x%X",
				r.Name, r.Offset+r.Size, l.End,
			))
		}
	}

	// Check no two regions overlap (O(n²) is fine for ~20 regions).
	for i := 0; i < len(l.Regions); i++ {
		for j := i + 1; j < len(l.Regions); j++ {
			a, b := l.Regions[i], l.Regions[j]
			aEnd := a.Offset + a.Size
			bEnd := b.Offset + b.Size
			if a.Offset < bEnd && b.Offset < aEnd {
				panic(fmt.Sprintf(
					"cave regions overlap: %q [0x%X-0x%X) and %q [0x%X-0x%X)",
					a.Name, a.Offset, aEnd, b.Name, b.Offset, bEnd,
				))
			}
		}
	}
}

// ---- ADV subtitle code cave ----

// AdvSubtitleLayout is the pre-defined layout for the ADV subtitle
// injection cave in SLPM_623.91 (0x00159844 - 0x00163400).
var AdvSubtitleLayout = Layout{
	Start: 0x00159844,
	End:   0x00163400,
	Regions: []Region{
		{Name: "initHelper", Offset: 0x00159844, Size: 0x800},
		{Name: "window", Offset: 0x0015A044, Size: 0x700},
		{Name: "string0", Offset: 0x0015A744, Size: 0x200},
		{Name: "string1", Offset: 0x0015A944, Size: 0x200},
		{Name: "string2", Offset: 0x0015AB44, Size: 0x200},
		{Name: "string3", Offset: 0x0015AD44, Size: 0x200},
		{Name: "flag", Offset: 0x0015AF44, Size: 4},
		{Name: "empty", Offset: 0x0015AF48, Size: 1},
		{Name: "debugBuffer", Offset: 0x0015AF80, Size: 0x80},
		{Name: "showHelper", Offset: 0x0015B000, Size: 0x1000},
		{Name: "hexTable", Offset: 0x0015C000, Size: 0x80},
		{Name: "bgRoot", Offset: 0x0015C080, Size: 0x480},
		{Name: "bgInitHelper", Offset: 0x0015C500, Size: 0x500},
		{Name: "hideHelper", Offset: 0x0015CA00, Size: 0x200},
		{Name: "doneCleanupHelper", Offset: 0x0015CC00, Size: 0x200},
		{Name: "subtitleTable", Offset: 0x0015CE00, Size: 0x6600},
	},
}

// AdvHookPoints are the addresses in SLPM_623.91 where J instructions are
// written to redirect the game's voice-line dispatch to the injected stubs.
var AdvHookPoints = struct {
	Init       uint64 // one-time initialisation hook
	VoicePathA uint64 // voice dispatch hook (path A)
	VoicePathB uint64 // voice dispatch hook (path B)
	VoiceEnd   uint64 // voice end / hide subtitle hook
	AdvDone    uint64 // adventure cutscene done hook
}{
	Init:       0x000F4FD0,
	VoicePathA: 0x000F48E4,
	VoicePathB: 0x000F4914,
	VoiceEnd:   0x000F492C,
	AdvDone:    0x000F4FF0,
}

// HUD weapon name cave

// HUDWeaponLayout is the layout for the HUD weapon name region within the
// broader code cave (shares space with the subtitle cave end).
var HUDWeaponLayout = Layout{
	Start: 0x163400,
	End:   0x163840,
	Regions: []Region{
		{Name: "names", Offset: 0x163400, Size: 0x400},
		{Name: "pointerTable", Offset: 0x165570, Size: 0x5C}, // 23 * 4 bytes
	},
}
