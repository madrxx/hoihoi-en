package patcher

import "github.com/madrxx/hoihoi-en/encoding"

// applyTickerPatches replaces bottom-of-screen help ticker text for main
// menu, option menu, custom menu, mission menu, and context menus in
// SLPM_623.91 at 0x16E4C0-0x170E00.
func applyTickerPatches(p *Patcher) {
	generalTicker := "←/→ to select. ○ to confirm. × to open submenu."
	missionTicker := "Activate Hoihoi-san to exterminate pests. Clear missions to earn funds."
	customTicker := "Customize Hoihoi-san's weapons and outfit."

	p.WriteFileBytes("SLPM_623.91", 0x16F8B2, encoding.GameText(missionTicker+"		　　　　　　　　　　		　　　　　　　　　　		"+generalTicker))
	p.WriteFileBytes("SLPM_623.91", 0x16F792, encoding.GameText("Buy weapons, ammo and outfits for Hoihoi-san."+"		　　　　　　　　　　		　　　　　　　　　　		"+generalTicker))
	p.WriteFileBytes("SLPM_623.91", 0x16F672, encoding.GameText(customTicker+"		　　　　　　　　　　		　　　　　　　　　　		"+generalTicker))
	p.WriteFileBytes("SLPM_623.91", 0x16F572, encoding.GameText("Change settings or save/load."+"		　　　　　　　　　　		　　　　　　　　　　		"+generalTicker))

	p.WriteFileBytes("SLPM_623.91", 0x170E00, encoding.GameText("Save state to memory card."))
	p.WriteFileBytes("SLPM_623.91", 0x170DC0, encoding.GameText("Load state from memory card."))
	p.WriteFileBytes("SLPM_623.91", 0x170D90, encoding.GameText("Set stereo/mono."))
	p.WriteFileBytes("SLPM_623.91", 0x170D60, encoding.GameText("Toggle vibration."))
	p.WriteFileBytes("SLPM_623.91", 0x170D30, encoding.GameText("Change bug model."))
	p.WriteFileBytes("SLPM_623.91", 0x170CF0, encoding.GameText("Select a control scheme."))
	p.WriteFileBytes("SLPM_623.91", 0x170CC0, encoding.GameText("Toggle auto save."))
	p.WriteFileBytes("SLPM_623.91", 0x170C90, encoding.GameText("Return to index."))

	p.WriteFileBytes("SLPM_623.91", 0x16E850, encoding.GameText("		Choose a weapon for Hoihoi-san's left hand.		　　　　　　　　　　		　　　　　　　　　　		←/→ to select. ○ to confirm. × to open submenu."))
	p.WriteFileBytes("SLPM_623.91", 0x16E700, encoding.GameText("		Choose a weapon for Hoihoi-san's right hand.		　　　　　　　　　　		　　　　　　　　　　		←/→ to select. ○ to confirm. × to open submenu."))
	p.WriteFileBytes("SLPM_623.91", 0x16E5E0, encoding.GameText("		Choose an outfit for Hoihoi-san. 		　　　　　　　　　　		　　　　　　　　　　		←/→ to select. ○ to confirm. × to open submenu."))
	p.WriteFileBytes("SLPM_623.91", 0x16E4C0, encoding.GameText("		Check remaining ammunition, and equip antennas and accessories.		　　　　　　　　　　		　　　　　　　　　　		←/→ to select. ○ to confirm. × to open submenu."))

	p.WriteFileBytes("SLPM_623.91", 0x17007D, encoding.GameText("Press × to open submenu. "))

	p.WriteFileBytes("SLPM_623.91", 0x16E9A0, encoding.GameText("Close submenu."))
	p.WriteFileBytes("SLPM_623.91", 0x16E9C0, encoding.GameText(missionTicker))
	p.WriteFileBytes("SLPM_623.91", 0x16EA60, encoding.GameText("Go back home."))
	p.WriteFileBytes("SLPM_623.91", 0x1700B0, encoding.GameText("Close submenu."))
	p.WriteFileBytes("SLPM_623.91", 0x1700D0, encoding.GameText(customTicker))
	p.WriteFileBytes("SLPM_623.91", 0x170130, encoding.GameText("Go back home."))
}
