package patcher

import "github.com/madrxx/hoihoi-en/encoding"

// applyContextMenuPatch replaces context menu labels in SLPM_623.91
// at 0x16EAD0-0x170198 (Mission, X, Custom, Index).
func applyContextMenuPatch(p *Patcher) {
	p.WriteFileBytes("SLPM_623.91", 0x16EAD0, encoding.GameText("Mission"))
	p.WriteFileBytes("SLPM_623.91", 0x16EAC0, encoding.GameText("  X"))
	p.WriteFileBytes("SLPM_623.91", 0x170178, encoding.GameText("  X"))
	p.WriteFileBytes("SLPM_623.91", 0x170188, encoding.GameText("Custom"))
	p.WriteFileBytes("SLPM_623.91", 0x16EAE0, encoding.GameText(" Index"))
	p.WriteFileBytes("SLPM_623.91", 0x170198, encoding.GameText(" Index"))
}

// applyPauseMenuPatch replaces pause menu labels in SLPM_623.91
// at 0x177040-0x177060 (Resume, Goals, Abort).
func applyPauseMenuPatch(p *Patcher) {
	p.WriteFileBytes("SLPM_623.91", 0x177060, encoding.GameText("Resume"))
	p.WriteFileBytes("SLPM_623.91", 0x177050, encoding.GameText("Goals"))
	p.WriteFileBytes("SLPM_623.91", 0x177040, encoding.GameText("Abort"))
}
