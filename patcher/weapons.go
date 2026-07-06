// HUD weapon name injection into the executable's code cave, combined with
// item/weapon/outfit text patching. Writes HUD display names into the
// pointer table at the HUD weapon name cave region and applies all item
// text patches to the BSV tables.

package patcher

import (
	"encoding/binary"
	"log"

	"github.com/madrxx/hoihoi-en/encoding"
	"github.com/madrxx/hoihoi-en/game"
	"github.com/madrxx/hoihoi-en/text"
)

// writeHUDWeaponNames encodes weapon names to game text, writes them into
// the HUD weapon name cave, and builds their runtime pointer table.
func (p *Patcher) writeHUDWeaponNames(names []string) {
	const runtimeBase uint32 = 0x000FFF80

	var blob []byte
	pointers := make([]byte, len(names)*4)

	for i, name := range names {
		ptr := uint32(game.HUDNamesOffset) + runtimeBase + uint32(len(blob))
		binary.LittleEndian.PutUint32(pointers[i*4:i*4+4], ptr)
		blob = append(blob, encoding.GameText(name)...)
	}

	capacity := game.HUDNamesEndOffset - game.HUDNamesOffset
	if uint64(len(blob)) > capacity {
		log.Fatalf("HUD weapon names overflow: len=0x%X capacity=0x%X overBy=0x%X",
			len(blob), capacity, uint64(len(blob))-capacity)
	}

	p.WriteFileBytes(game.ExecutableFile, game.HUDNamesOffset, make([]byte, int(capacity)))
	p.WriteFileBytes(game.ExecutableFile, game.HUDNamesOffset, blob)
	p.WriteFileBytes(game.ExecutableFile, game.HUDPointerTableOffset, pointers)
}

// hudNamesFromItemPatches extracts HUD display names from item patches,
// falling back from HUDName to Name for each main table row.
func hudNamesFromItemPatches(patches []text.ItemTextPatch) []string {
	names := make([]string, game.HUDWeaponCount)
	for _, patch := range patches {
		if patch.MainRow < 0 || patch.MainRow >= len(names) {
			continue
		}
		if patch.HUDName != nil {
			names[patch.MainRow] = *patch.HUDName
		} else if patch.Name != nil {
			names[patch.MainRow] = *patch.Name
		}
	}
	return names
}

// ItemAndWeaponTexts applies item/weapon/outfit text patches to BSV tables
// and writes HUD weapon name strings into the executable's code cave.
func (p *Patcher) ItemAndWeaponTexts(patches []text.ItemTextPatch) {
	p.ItemTexts(patches)
	p.writeHUDWeaponNames(hudNamesFromItemPatches(patches))
}
