// Package game provides named constants for the game's reverse-engineered
// data structures: BSV table bases, field indices, executable layout, and
// file paths on the disc. These replace magic numbers in the patching code
// with names that reflect the game's actual architecture as revealed by
// Ghidra analysis of the PS2 binary.
package game

// ---- Disc file paths ----

const (
	ExecutableFile = "SLPM_623.91"
	GameUFP        = "UFP/GAME.UFP"
)

// ---- BSV table bases (file offsets within GameUFP) ----

// These correspond to BSV tables loaded by CHoiBsvManager::search() at
// runtime. Each table has the standard SRTS/EMAN/DNIK/ATAD section layout
// parsed by the bsv package.
const (
	BSVMainItem       uint64 = 0x1D2000 // main item/weapon/outfit table (80 rows)
	BSVItemList1      uint64 = 0x1C5000 // item list copy 1 (70 rows)
	BSVItemList2      uint64 = 0x1C9800 // item list copy 2 (70 rows)
	BSVOutfitRelation uint64 = 0x1C8000 // outfit relation/condition table
	BSVMission        uint64 = 0x1D8800 // mission records (37 rows, 20 used)
)

// ---- Mission record layout ----

// Each mission record is 0x74 bytes. Fields are u32 string offsets relative
// to the BSV base. The game reads these via CHoiMissionManager::loadMission,
// which calls GetBsvData(bsv, row, fieldIndex) where fieldIndex is a column
// index (not a byte offset) into 12-byte field descriptors within the row.
const (
	MissionRecordSize  uint64 = 0x74
	MissionRecordCount int    = 20 // first 20 of 37 rows are mission records

	MissionFieldNumber      uint64 = 0x0C // mission stage number ("1-1")
	MissionFieldID          uint64 = 0x10 // mission identifier
	MissionFieldTitle       uint64 = 0x18 // mission title string
	MissionFieldDescription uint64 = 0x1C // mission description string
	MissionFieldConfirm1    uint64 = 0x60 // confirmation objective line 1
	MissionFieldConfirm2    uint64 = 0x64 // confirmation objective line 2
	MissionFieldConfirm3    uint64 = 0x68 // confirmation objective line 3
	MissionFieldConfirm4    uint64 = 0x6C // confirmation objective line 4
)

// Mission text pool boundaries within the mission BSV's SRTS section.
// Only this sub-range is cleared and rebuilt; earlier column name/enum
// strings and later non-mission strings are preserved.
const (
	MissionTextStart uint64 = 0x1D8B21
	MissionTextEnd   uint64 = 0x1D956E
)

// ---- Item table field indices ----

// Field indices within the main item table (BSVMainItem). These are word
// offsets within each row used by p.ReadFileU32LE(t.FilePath, t.FieldOffset(row, field)).
const (
	ItemFieldName        = 3
	ItemFieldNameCopy    = 4 // duplicate name reference
	ItemFieldType        = 6
	ItemFieldInfo1       = 8
	ItemFieldInfo2       = 9
	ItemFieldInfo3       = 10
	ItemFieldDescription = 11
)

// Field indices within item list tables (BSVItemList1, BSVItemList2).
const (
	ItemListFieldName     = 3
	ItemListFieldNameCopy = 4 // duplicate name reference
	ItemListFieldCategory = 6
)

// Field indices within the outfit relation table (BSVOutfitRelation).
const (
	OutfitFieldName     = 3
	OutfitFieldNoCond   = 4
	OutfitFieldWithCond = 5
)

// ---- HUD weapon name region ----

const (
	// HUDNamesOffset is the start of the weapon name string blob in the
	// executable's code cave (shares space with the subtitle cave tail).
	HUDNamesOffset uint64 = 0x163400
	// HUDNamesEndOffset is one past the last byte of the weapon name region.
	HUDNamesEndOffset uint64 = 0x163840
	// HUDPointerTableOffset is the pointer table for the weapon names.
	HUDPointerTableOffset uint64 = 0x165570

	// HUDWeaponCount is the number of weapon entries (rows 0-22) displayed
	// in the in-game HUD.
	HUDWeaponCount = 23
)
