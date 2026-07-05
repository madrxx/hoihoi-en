// Event modal label replacements  -- item names displayed in the event
// reward / NPC showing you an item dialog. Each entry has:
//
//	Base     -- offset within UFP/GAME.UFP of the BSV table
//	OldText  -- the original Japanese fullwidth string to find and replace
//	NewText  -- the replacement English text
//
// The patcher locates the OldText in the BSV string pool, replaces it
// with the encoded NewText, and updates all u32 references. If the
// English text is longer than the Japanese, the SRTS section is grown
// to accommodate it.

package main

var eventModalPatches = []EventModalPatch{
	{
		Base:    0x2FB000,
		OldText: "ポイントカードを受け取った",
		NewText: "Point Card",
	},
	{
		Base:    0x2FB800,
		OldText: "ウェイトレスドレス",
		NewText: "Waitress Dress",
	},
	{
		Base:    0x2FC000,
		OldText: "ホイホイさん移動リフト",
		NewText: "Mobile Lift",
	},
	{
		Base:    0x2FC800,
		OldText: "巫女装束",
		NewText: "Shrine Maiden",
	},
	{
		Base:    0x2FD000,
		OldText: "マジカルドレス",
		NewText: "Magical Dress",
	},
	{
		Base:    0x2FF800,
		OldText: "ポイントシールをもらった",
		NewText: "Point Sticker",
	},
	{
		Base:    0x300000,
		OldText: "ゴシックドレスをもらった",
		NewText: "Gothic Dress",
	},
	{
		Base:    0x300800,
		OldText: "殲滅指令!!コンバットさん本体",
		NewText: "Combat-san",
	},
	{
		Base:    0x301000,
		OldText: "殲滅指令!!コンバットさん本体",
		NewText: "Combat-san",
	},
	{
		Base:    0x301800,
		OldText: "ポイントカード",
		NewText: "Point Card",
	},
	{
		Base:    0x302000,
		OldText: "ネコ耳をもらった",
		NewText: "Cat Ears",
	},
}
