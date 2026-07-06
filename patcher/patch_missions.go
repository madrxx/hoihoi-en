package patcher

import (
	"github.com/madrxx/hoihoi-en/encoding"
	"github.com/madrxx/hoihoi-en/text"
)

// applyMissionPatch replaces the online ranking label in SLPM_623.91 at
// 0x1701E0 and applies all 20 mission title/description/objective patches
// via MissionTexts, which rebuilds the mission BSV table's string pool at
// offsets 0x1D99D4-0x1DA000 in GAME.UFP.
func applyMissionPatch(p *Patcher) {
	p.WriteFileBytes("SLPM_623.91", 0x1701E0, encoding.GameText("Online Ranking PW"))
	p.MissionTexts(text.Missions)
}
