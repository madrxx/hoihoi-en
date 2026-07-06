package patcher

import (
	"github.com/madrxx/hoihoi-en/encoding"
	"github.com/madrxx/hoihoi-en/text"
)

// applyWeaponPatch replaces HUD weapon display strings, card/point/puzzle
// format strings, and ammo counters in SLPM_623.91 at 0x16E346-0x16E40B,
// then applies all item and weapon name patches.
func applyWeaponPatch(p *Patcher) {
	p.WriteFileBytes("SLPM_623.91", 0x16E380, encoding.GameText("Press □\nto play\naudio."))
	p.ItemAndWeaponTexts(text.Items)

	pointCardText := toSJIS(encoding.ToFullWidth("Card "), 0)
	pointCardText = append(pointCardText, []byte("%s")...)
	pointCardText = append(pointCardText, toSJIS(encoding.ToFullWidth("\n"), 0)...)
	pointCardText = append(pointCardText, []byte("%s")...)
	pointCardText = append(pointCardText, encoding.GameText(" point")...)
	p.WriteFileBytes("SLPM_623.91", 0x16E3C3, pointCardText)

	puzzleText := []byte("%s")
	puzzleText = append(puzzleText, toSJIS(encoding.ToFullWidth("%\n"), 0)...)
	puzzleText = append(puzzleText, []byte("%s")...)
	puzzleText = append(puzzleText, encoding.GameText(" remain")...)
	p.WriteFileBytes("SLPM_623.91", 0x16E3E3, puzzleText)

	boxText := []byte("%s")
	boxText = append(boxText, encoding.GameText("x")...)
	p.WriteFileBytes("SLPM_623.91", 0x16E40B, boxText)

	ammoText := []byte("%s")
	ammoText = append(ammoText, toSJIS(encoding.ToFullWidth("x "), 0)...)
	ammoText = append(ammoText, []byte("%s")...)
	ammoText = append(ammoText, encoding.GameText("r")...)
	p.WriteFileBytes("SLPM_623.91", 0x16E346, ammoText)
}
