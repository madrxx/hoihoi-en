package main

import "github.com/madrxx/hoihoi-en/encoding"

// applyMemoryCardPatch replaces memory card UI strings in SLPM_623.91:
// checking/saving/loading/formatting/error messages (~35 strings at
// 0x178453-0x178991).
func applyMemoryCardPatch(p *Patcher) {
	p.WriteFileBytes("SLPM_623.91", 0x178453, encoding.GameText("Checking"))
	p.WriteFileBytes("SLPM_623.91", 0x17846C, encoding.GameText("Slot 1."))
	p.WriteFileBytes("SLPM_623.91", 0x178481, encoding.GameText("Do not remove"))
	p.WriteFileBytes("SLPM_623.91", 0x1784A0, encoding.GameText("memory card."))

	p.WriteFileBytes("SLPM_623.91", 0x1784B9, encoding.GameText("Memory card"))
	p.WriteFileBytes("SLPM_623.91", 0x1784D2, encoding.GameText("not inserted"))
	p.WriteFileBytes("SLPM_623.91", 0x1784F1, encoding.GameText("in Slot 1."))

	p.WriteFileBytes("SLPM_623.91", 0x178506, encoding.GameText("27kB required"))
	p.WriteFileBytes("SLPM_623.91", 0x178529, encoding.GameText("to save progress."))

	p.WriteFileBytes("SLPM_623.91", 0x17854C, encoding.GameText("Loading."))
	p.WriteFileBytes("SLPM_623.91", 0x17855F, encoding.GameText("Loaded."))

	p.WriteFileBytes("SLPM_623.91", 0x178570, encoding.GameText("Saving game"))
	p.WriteFileBytes("SLPM_623.91", 0x17858B, encoding.GameText("data will not"))
	p.WriteFileBytes("SLPM_623.91", 0x1785AA, encoding.GameText("be possible."))

	p.WriteFileBytes("SLPM_623.91", 0x1785C5, encoding.GameText("Check again?"))
	p.WriteFileBytes("SLPM_623.91", 0x1785E4, encoding.GameText("Start the game?"))

	p.WriteFileBytes("SLPM_623.91", 0x178603, encoding.GameText("A save exists"))
	p.WriteFileBytes("SLPM_623.91", 0x17861E, encoding.GameText("for Hoihoi."))

	p.WriteFileBytes("SLPM_623.91", 0x178639, encoding.GameText("Overwrite it?"))

	p.WriteFileBytes("SLPM_623.91", 0x17865C, encoding.GameText("Please insert"))
	p.WriteFileBytes("SLPM_623.91", 0x178677, encoding.GameText("another card."))

	p.WriteFileBytes("SLPM_623.91", 0x178696, encoding.GameText("Create a"))
	p.WriteFileBytes("SLPM_623.91", 0x1786AB, encoding.GameText("new save?"))

	p.WriteFileBytes("SLPM_623.91", 0x1786C6, encoding.GameText("Saving."))
	p.WriteFileBytes("SLPM_623.91", 0x1789FF, encoding.GameText("Saved."))

	p.WriteFileBytes("SLPM_623.91", 0x1786D9, encoding.GameText("Save not found."))
	p.WriteFileBytes("SLPM_623.91", 0x1786FE, encoding.GameText("Load error."))
	p.WriteFileBytes("SLPM_623.91", 0x178715, encoding.GameText("Save error."))

	p.WriteFileBytes("SLPM_623.91", 0x17872C, encoding.GameText("Not formatted."))
	p.WriteFileBytes("SLPM_623.91", 0x178749, encoding.GameText("Format it?"))

	p.WriteFileBytes("SLPM_623.91", 0x178760, encoding.GameText("Formatting."))

	p.WriteFileBytes("SLPM_623.91", 0x178779, encoding.GameText("Format failed."))

	p.WriteFileBytes("SLPM_623.91", 0x178796, encoding.GameText("Insufficient"))
	p.WriteFileBytes("SLPM_623.91", 0x1787B5, encoding.GameText("space."))

	p.WriteFileBytes("SLPM_623.91", 0x1787D0, encoding.GameText("Memory card replaced."))

	p.WriteFileBytes("SLPM_623.91", 0x17892D, encoding.GameText("Loading from"))
	p.WriteFileBytes("SLPM_623.91", 0x178946, encoding.GameText("Slot 1 will"))
	p.WriteFileBytes("SLPM_623.91", 0x178963, encoding.GameText("lose any"))
	p.WriteFileBytes("SLPM_623.91", 0x178976, encoding.GameText("current data."))
	p.WriteFileBytes("SLPM_623.91", 0x178991, encoding.GameText("Continue?"))
}
