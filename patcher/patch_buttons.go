package patcher

import "github.com/madrxx/hoihoi-en/encoding"

// applyButtonPatches replaces controller button labels (L1, □, ×, ○) in
// SLPM_623.91 at 0x1773F8-0x177580.
func applyButtonPatches(p *Patcher) {
	p.WriteFileBytes("SLPM_623.91", 0x1774C8, encoding.GameText("L1"))
	p.WriteFileBytes("SLPM_623.91", 0x177478, encoding.GameText("L1"))
	p.WriteFileBytes("SLPM_623.91", 0x1774F8, encoding.GameText("□"))
	p.WriteFileBytes("SLPM_623.91", 0x177408, encoding.GameText("□"))
	p.WriteFileBytes("SLPM_623.91", 0x177488, encoding.GameText("□"))
	p.WriteFileBytes("SLPM_623.91", 0x177508, encoding.GameText("×"))
	p.WriteFileBytes("SLPM_623.91", 0x177418, encoding.GameText("×"))
	p.WriteFileBytes("SLPM_623.91", 0x177448, encoding.GameText("×"))
	p.WriteFileBytes("SLPM_623.91", 0x177498, encoding.GameText("×"))
	p.WriteFileBytes("SLPM_623.91", 0x1773F8, encoding.GameText("○"))
	p.WriteFileBytes("SLPM_623.91", 0x177580, encoding.GameText("○ Next page"))
}
