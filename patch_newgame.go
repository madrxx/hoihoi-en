package main

import "github.com/madrxx/hoihoi-en/encoding"

// applyNewGamePatch replaces new-game and purchase confirmation dialogs in
// SLPM_623.91 at 0x172380-0x176ADF (name entry prompts, purchase flow).
func applyNewGamePatch(p *Patcher) {
	p.WriteFileBytes("SLPM_623.91", 0x176A24, encoding.GameText("Please fill in\nall fields."))
	p.WriteFileBytes("SLPM_623.91", 0x176A5C, encoding.GameText("Overwriting\nexisting user.\nIs that okay?"))
	p.WriteFileBytes("SLPM_623.91", 0x176AAF, encoding.GameText("Submit these\ndetails?"))
	p.WriteFileBytes("SLPM_623.91", 0x176ADF, encoding.GameText("Return to the\ntitle screen?"))

	p.WriteFileBytes("SLPM_623.91", 0x172390, encoding.GameText("1-Hit Bug Killer\"Hoihoi-san\"\n\n¥29800"))
	p.WriteFileBytes("SLPM_623.91", 0x172380, encoding.GameText("Buy it?"))
}
