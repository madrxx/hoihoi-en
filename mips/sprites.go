package mips

// Sprite field offsets for the game's 2D primitive class hierarchy.
//
// These encode the reverse-engineered memory layouts of:
//
//   CJL2DPrimitiveTypeS  -- container / base 2D primitive
//   CJLPolygonTypeR      -- filled rectangle (extends CJL2DPrimitiveTypeR)
//   CJLSpriteTypeS        -- textured sprite (extends CJL2DPrimitiveTypeS,
//                           embeds CJLSpriteInterface at +0xAC)
//   CJLStringSprite       -- text sprite
//
// The subtitle background overlay is constructed as a sprite hierarchy
// rooted at bgRoot (a CJL2DPrimitiveTypeS allocated in the code cave):
//
//   bgRoot (0x000)         -- CJL2DPrimitiveTypeS, container
//   ├── bgFill (0x0AC)     -- CJLPolygonTypeR, black translucent fill
//   ├── bgTop  (0x180)     -- CJLSpriteTypeS, top cap (iface at +0xAC)
//   └── bgBottom (0x254)   -- CJLSpriteTypeS, bottom cap (iface at +0xAC)
//
// Offsets are root-relative for direct use with the bgRoot base register ($3).

// ---- Root container (CJL2DPrimitiveTypeS at +0x000) ----

const (
	RootVisible = 0x0D // byte: visibility flag
	RootZIndex  = 0x10 // u32: render priority
	RootPosX    = 0x90 // u32: x position
	RootPosY    = 0x94 // u32: y position
	RootWidth   = 0x98 // u32: width
	RootHeight  = 0x9C // u32: height
)

// ---- Fill (CJLPolygonTypeR at +0x0AC) ----

const (
	FillVisible  = 0xB9 // byte: visibility (fill base + 0x0D)
	FillPriority = 0xBC // u32: render priority (fill base + 0x10)
	FillColorR   = 0xF5 // byte: red channel (fill base + 0x49)
	FillColorG   = 0xF6 // byte: green channel (fill base + 0x4A)
	FillColorB   = 0xF7 // byte: blue channel (fill base + 0x4B)
	FillAlpha    = 0xF8 // byte: alpha channel (fill base + 0x4C)
	FillLeft     = 0x13C // u32: left edge (fill base + 0x90)
	FillTop      = 0x140 // u32: top edge (fill base + 0x94)
	FillRight    = 0x144 // u32: right edge (fill base + 0x98)
	FillBottom   = 0x148 // u32: bottom edge (fill base + 0x9C)
)

// ---- Top cap sprite (CJLSpriteTypeS at +0x180, CJLSpriteInterface at +0x22C) ----

const (
	TopCapVisible = 0x18D // byte: visibility (top cap + 0x0D)
	TopCapZIndex  = 0x190 // u32: render priority (top cap + 0x10)
	TopCapAlpha   = 0x1CC // byte: alpha channel (top cap + 0x4C)
	TopCapPosX    = 0x210 // u32: x position (top cap + 0x90)
	TopCapPosY    = 0x214 // u32: y position (top cap + 0x94)
	TopCapWidth   = 0x218 // u32: width (top cap + 0x98)
	TopCapHeight  = 0x21C // u32: height (top cap + 0x9C)
	TopCapScaleX  = 0x220 // u32 (float): x scale (top cap + 0xA0)
	TopCapScaleY  = 0x224 // u32 (float): y scale (top cap + 0xA4)
)

// ---- Bottom cap sprite (CJLSpriteTypeS at +0x254, CJLSpriteInterface at +0x300) ----

const (
	BotCapVisible = 0x261 // byte: visibility (bottom cap + 0x0D)
	BotCapZIndex  = 0x264 // u32: render priority (bottom cap + 0x10)
	BotCapAlpha   = 0x2A0 // byte: alpha channel (bottom cap + 0x4C)
	BotCapPosX    = 0x2E4 // u32: x position (bottom cap + 0x90)
	BotCapPosY    = 0x2E8 // u32: y position (bottom cap + 0x94)
	BotCapWidth   = 0x2EC // u32: width (bottom cap + 0x98)
	BotCapHeight  = 0x2F0 // u32: height (bottom cap + 0x9C)
	BotCapScaleX  = 0x2F4 // u32 (float): x scale (bottom cap + 0xA0)
	BotCapScaleY  = 0x2F8 // u32 (float): y scale (bottom cap + 0xA4)
)

// ---- Color field offsets (on CJLStringSprite, CJLPolygonTypeR, etc.) ----

const (
	StringColorR = 0x49 // byte: red
	StringColorG = 0x4A // byte: green
	StringColorB = 0x4B // byte: blue
)
