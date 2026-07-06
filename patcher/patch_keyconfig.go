package patcher

import "github.com/madrxx/hoihoi-en/encoding"

// applyKeyConfigPatch replaces key configuration labels in SLPM_623.91
// at 0x164C40-0x164CE0 (FP View, Tiptoe, Refocus, Dash, L/R Atk).
func applyKeyConfigPatch(p *Patcher) {
	p.WriteFileBytes("SLPM_623.91", 0x164C40, encoding.GameText("FP View"))
	p.WriteFileBytes("SLPM_623.91", 0x164C60, encoding.GameText("Tiptoe"))
	p.WriteFileBytes("SLPM_623.91", 0x164C80, encoding.GameText("Refocus"))
	p.WriteFileBytes("SLPM_623.91", 0x164CA0, encoding.GameText("Dash"))
	p.WriteFileBytes("SLPM_623.91", 0x164CC0, encoding.GameText("L Atk"))
	p.WriteFileBytes("SLPM_623.91", 0x164CE0, encoding.GameText("R Atk"))
}
