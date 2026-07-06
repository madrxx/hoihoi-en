package patcher

import "github.com/madrxx/hoihoi-en/encoding"

// applyModalPatches replaces generic Yes/No/Cancel/Start/Okay buttons,
// confirmation dialogs, item-obtained popups, and lift battery messages
// in SLPM_623.91 at 0x164780-0x178991 (~35 strings).
func applyModalPatches(p *Patcher) {
	p.WriteFileBytes("SLPM_623.91", 0x16FA10, encoding.GameText("Yes"))
	p.WriteFileBytes("SLPM_623.91", 0x171420, encoding.GameText("Yes"))
	p.WriteFileBytes("SLPM_623.91", 0x172378, encoding.GameText("Yes"))
	p.WriteFileBytes("SLPM_623.91", 0x176030, encoding.GameText("Yes"))
	p.WriteFileBytes("SLPM_623.91", 0x1767E8, encoding.GameText("Yes"))
	p.WriteFileBytes("SLPM_623.91", 0x177038, encoding.GameText("Yes"))
	p.WriteFileBytes("SLPM_623.91", 0x1776E8, encoding.GameText("Yes"))
	p.WriteFileBytes("SLPM_623.91", 0x178432, encoding.GameText("Yes"))

	p.WriteFileBytes("SLPM_623.91", 0x16FA08, encoding.GameText("No"))
	p.WriteFileBytes("SLPM_623.91", 0x171418, encoding.GameText("No"))
	p.WriteFileBytes("SLPM_623.91", 0x172370, encoding.GameText("No"))
	p.WriteFileBytes("SLPM_623.91", 0x176028, encoding.GameText("No"))
	p.WriteFileBytes("SLPM_623.91", 0x1767E0, encoding.GameText("No"))
	p.WriteFileBytes("SLPM_623.91", 0x177030, encoding.GameText("No"))
	p.WriteFileBytes("SLPM_623.91", 0x1776E0, encoding.GameText("No"))
	p.WriteFileBytes("SLPM_623.91", 0x178439, encoding.GameText("No"))

	p.WriteFileBytes("SLPM_623.91", 0x170558, encoding.GameText("Cancel"))
	p.WriteFileBytes("SLPM_623.91", 0x170640, encoding.GameText("Cancel"))
	p.WriteFileBytes("SLPM_623.91", 0x170568, encoding.GameText("Start"))
	p.WriteFileBytes("SLPM_623.91", 0x170650, encoding.GameText("Start"))

	p.WriteFileBytes("SLPM_623.91", 0x17843E, encoding.GameText("Okay"))
	p.WriteFileBytes("SLPM_623.91", 0x178447, encoding.GameText(" No"))

	p.WriteFileBytes("SLPM_623.91", 0x177640, encoding.GameText("Repeat tutorial?"))
	p.WriteFileBytes("SLPM_623.91", 0x177670, encoding.GameText("Repeat tutorial?"))

	p.WriteFileBytes("SLPM_623.91", 0x1776A0, encoding.GameText("Tutorial?"))
	p.WriteFileBytes("SLPM_623.91", 0x1776C0, encoding.GameText("Tutorial?"))

	p.WriteFileBytes("SLPM_623.91", 0x177010, encoding.GameText("Are you sure?"))

	p.WriteFileBytes("SLPM_623.91", 0x16FA20, encoding.GameText("Back to title?"))

	p.WriteFileBytes("SLPM_623.91", 0x16FA49, encoding.GameText("100% complete!"))

	p.WriteFileBytes("SLPM_623.91", 0x16FA73, encoding.GameText("Bonus:"))

	p.WriteFileBytes("SLPM_623.91", 0x16FA8C, encoding.GameText("¥100000"))

	p.WriteFileBytes("SLPM_623.91", 0x16DD40, append([]byte{0x25, 0x73, 0x0A /* "%s\n" */}, encoding.GameText("    equipped")...))

	p.WriteFileBytes("SLPM_623.91", 0x16EA90, encoding.GameText("Cannot use\nthis outfit."))
	p.WriteFileBytes("SLPM_623.91", 0x16FAB0, encoding.GameText("Cannot use\nthis outfit."))

	p.WriteFileBytes("SLPM_623.91", 0x164780, encoding.GameText("		No ammo in L Hand:\n　　　　　　　　　　\n"))
	p.WriteFileBytes("SLPM_623.91", 0x164880, encoding.GameText("		No ammo in R Hand:\n　　　　　　　　　　\n"))
	p.WriteFileBytes("SLPM_623.91", 0x170585, encoding.GameText("Continue?\n"))
	p.WriteFileBytes("SLPM_623.91", 0x1705B3, encoding.GameText("Continue?\n"))
	p.WriteFileBytes("SLPM_623.91", 0x1705E3, encoding.GameText("Continue?\n"))

	p.WriteFileBytes("SLPM_623.91", 0x176D50, encoding.GameText("Item obtained!!"))
	p.WriteFileBytes("SLPM_623.91", 0x176E00, encoding.GameText("Item obtained!!"))
	p.WriteFileBytes("SLPM_623.91", 0x176E20, encoding.GameText("Item obtained!!"))

	p.WriteFileBytes("SLPM_623.91", 0x176D20, encoding.GameText("Battery obtained!!"))

	p.WriteFileBytes("SLPM_623.91", 0x176D00, encoding.GameText("Red lifts"))
	p.WriteFileBytes("SLPM_623.91", 0x176CE0, encoding.GameText("Green lifts"))
	p.WriteFileBytes("SLPM_623.91", 0x176CC0, encoding.GameText("Blue lifts"))
	p.WriteFileBytes("SLPM_623.91", 0x176CA0, encoding.GameText("now available."))

	p.WriteFileBytes("SLPM_623.91", 0x176D70, append(toSJIS(encoding.ToFullWidth("Stickers: "), 0), []byte{0x25, 0x73, 0x00}...))
	p.WriteFileBytes("SLPM_623.91", 0x176DA0, append(toSJIS(encoding.ToFullWidth("Puzzle pcs: "), 0), []byte{0x25, 0x73, 0x00}...))
	p.WriteFileBytes("SLPM_623.91", 0x176DD0, append(toSJIS(encoding.ToFullWidth("Money: ¥"), 0), []byte{0x25, 0x73, 0x00}...))
}
