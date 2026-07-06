package patcher

import "github.com/madrxx/hoihoi-en/encoding"

// applyTutorialPatch replaces the three tutorial screen texts in SLPM_623.91
// at 0x165A98-0x16626C (basic controls, tiptoe, dash slash).
func applyTutorialPatch(p *Patcher) {
	p.WriteFileBytes("SLPM_623.91", 0x177540, encoding.GameText("R1"))
	p.WriteFileBytes("SLPM_623.91", 0x177550, encoding.GameText("△"))
	p.WriteFileBytes("SLPM_623.91", 0x177560, encoding.GameText("○"))
	p.WriteFileBytes("SLPM_623.91", 0x177570, encoding.GameText("LS"))
	p.WriteFileBytes("SLPM_623.91", 0x1775B0, encoding.GameText("    :   Dash"))
	p.WriteFileBytes("SLPM_623.91", 0x1775D0, encoding.GameText("    :   L Atk"))
	p.WriteFileBytes("SLPM_623.91", 0x1775F0, encoding.GameText("    :   R Atk"))
	p.WriteFileBytes("SLPM_623.91", 0x177610, encoding.GameText("    :   Move"))

	p.WriteFileBytes("SLPM_623.91", 0x1773D0, encoding.GameText("Tilt    gently"))
	p.WriteFileBytes("SLPM_623.91", 0x177318, encoding.GameText("     LS"))
	p.WriteFileBytes("SLPM_623.91", 0x1773A0, encoding.GameText("or hold    to"))
	p.WriteFileBytes("SLPM_623.91", 0x177308, encoding.GameText("    R2"))
	p.WriteFileBytes("SLPM_623.91", 0x177370, encoding.GameText("      . If enemies"))
	p.WriteFileBytes("SLPM_623.91", 0x1772E0, encoding.GameText("tiptoe"))
	p.WriteFileBytes("SLPM_623.91", 0x177350, encoding.GameText("run away, sneak"))
	p.WriteFileBytes("SLPM_623.91", 0x177330, encoding.GameText("up on them."))

	p.WriteFileBytes("SLPM_623.91", 0x177160, encoding.GameText("Dash then melee to"))
	p.WriteFileBytes("SLPM_623.91", 0x1770F0, encoding.GameText("           before"))
	p.WriteFileBytes("SLPM_623.91", 0x1770C8, encoding.GameText("Dash Slash"))
	p.WriteFileBytes("SLPM_623.91", 0x177120, encoding.GameText("insects flee."))
	// Pointer fixups for relocated text lines.
	p.WriteFileU32LE("SLPM_623.91", 0x0F12E4, 0x24A57070)
	p.WriteFileU32LE("SLPM_623.91", 0x0F267C, 0x24A57070)
	p.WriteFileU32LE("SLPM_623.91", 0x0F1324, 0x24A57060)
	p.WriteFileU32LE("SLPM_623.91", 0x0F26BC, 0x24A57060)
	p.WriteFileU32LE("SLPM_623.91", 0x0F1304, 0x24A57060)
	p.WriteFileU32LE("SLPM_623.91", 0x0F269C, 0x24A57060)
}
