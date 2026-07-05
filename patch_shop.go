package main

import "github.com/madrxx/hoihoi-en/encoding"

// applyShopDialoguePatch replaces shopkeeper dialogue and shop UI labels in
// SLPM_623.91 at 0x171460-0x171C70 (greetings, category tabs, purchase flow).
func applyShopDialoguePatch(p *Patcher) {
	p.WriteFileBytes("SLPM_623.91", 0x171AD0, encoding.GameText("Welcome!"))
	p.WriteFileBytes("SLPM_623.91", 0x171AB8, encoding.GameText("Come in!"))
	p.WriteFileBytes("SLPM_623.91", 0x171AA8, encoding.GameText("Thanks!"))

	p.WriteFileBytes("SLPM_623.91", 0x171C70, encoding.GameText("Ranged weapons and ammo."))
	p.WriteFileBytes("SLPM_623.91", 0x171C40, encoding.GameText("Close-combat weapons."))
	p.WriteFileBytes("SLPM_623.91", 0x171C10, encoding.GameText("Outfits."))
	p.WriteFileBytes("SLPM_623.91", 0x171BE0, encoding.GameText("Misc. parts."))
	p.WriteFileBytes("SLPM_623.91", 0x171BC0, encoding.GameText("Go back home."))

	p.WriteFileBytes("SLPM_623.91", 0x171460, encoding.GameText("You don't have\nthe gun."))
	p.WriteFileBytes("SLPM_623.91", 0x171490, encoding.GameText("Too expensive."))
	p.WriteFileBytes("SLPM_623.91", 0x1714B3, encoding.GameText("   Purchase?"))
	p.WriteFileBytes("SLPM_623.91", 0x171540, encoding.GameText("How many?"))
	p.WriteFileBytes("SLPM_623.91", 0x1715AA, encoding.GameText("x\n	"))
	p.WriteFileBytes("SLPM_623.91", 0x1715A0, toSJIS(encoding.ToFullWidth(" "), 0))
	p.WriteFileBytes("SLPM_623.91", 0x1715C2, encoding.GameText(" bought"))
}
