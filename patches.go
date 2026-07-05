// Patch orchestration  -- defines the ordered list of 17 patch groups and
// the applyPatches() driver that runs them all. The order is fixed and
// must match the golden reference binary.

package main

// patchGroup describes one patch step in the translation pipeline.
type patchGroup struct {
	name  string         // human-readable label for debugging / skip flags
	apply func(*Patcher) // the patch function to call
}

// patchGroups defines all patch steps in the order they are applied.
// The order is fixed and must match the golden reference binary.
var patchGroups = []patchGroup{
	{"AdvSubtitle", (*Patcher).AdvSubtitlePatch},
	{"MemoryCard", applyMemoryCardPatch},
	{"NewGame", applyNewGamePatch},
	{"Tickers", applyTickerPatches},
	{"ShopDialogue", applyShopDialoguePatch},
	{"Buttons", applyButtonPatches},
	{"Tutorial", applyTutorialPatch},
	{"KeyConfig", applyKeyConfigPatch},
	{"EventModals", (*Patcher).EventModalLabels},
	{"Modals", applyModalPatches},
	{"ContextMenu", applyContextMenuPatch},
	{"PauseMenu", applyPauseMenuPatch},
	{"Missions", applyMissionPatch},
	{"Weapons", applyWeaponPatch},
	{"Assets", func(p *Patcher) { p.ApplyAssetManifest(defaultAssetManifestPath) }},
	{"DebugMenu", applyDebugMenuPatch},
}

// applyPatches runs every patch group against the patcher, in order.
func applyPatches(p *Patcher) {
	for _, g := range patchGroups {
		g.apply(p)
	}
}
