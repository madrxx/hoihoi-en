// Item text patching via BSV tables. Reads the main item table, two item
// list copies, and the outfit relation table from GAME.UFP, applies English
// string replacements, rebuilds string pools, and writes all tables back.
// Also provides DumpItemTexts for inspecting current on-disc text.
package main

import (
	"fmt"
	"log"

	"github.com/madrxx/hoihoi-en/bsv"
	"github.com/madrxx/hoihoi-en/game"
)

type ItemTextPatch struct {
	MainRow int

	Name        *string
	ListName    *string
	HUDName     *string
	Description *string
	Info1       *string
	Info2       *string
	Info3       *string
}

func strptr(s string) *string {
	return &s
}

type ItemTextRecord struct {
	Name        []byte
	Type        []byte
	Info1       []byte
	Info2       []byte
	Info3       []byte
	Description []byte
}

type ItemListRecord struct {
	Name     []byte
	Category []byte
}

type OutfitRelationRecord struct {
	Name     []byte
	NoCond   []byte
	WithCond []byte
}

// readMainItemRecords reads all rows from the main item BSV table, extracting
// name, type, info (3 lines), and description fields for each row.
func readMainItemRecords(p *Patcher, t BSVTable) []ItemTextRecord {
	records := make([]ItemTextRecord, int(t.RowCount))

	for row := 0; row < int(t.RowCount); row++ {
		records[row] = ItemTextRecord{
			Name:        p.ReadBSVString(t, p.ReadFileU32LE(t.FilePath, t.FieldOffset(row, game.ItemFieldName))),
			Type:        p.ReadBSVString(t, p.ReadFileU32LE(t.FilePath, t.FieldOffset(row, game.ItemFieldType))),
			Info1:       p.ReadBSVString(t, p.ReadFileU32LE(t.FilePath, t.FieldOffset(row, game.ItemFieldInfo1))),
			Info2:       p.ReadBSVString(t, p.ReadFileU32LE(t.FilePath, t.FieldOffset(row, game.ItemFieldInfo2))),
			Info3:       p.ReadBSVString(t, p.ReadFileU32LE(t.FilePath, t.FieldOffset(row, game.ItemFieldInfo3))),
			Description: p.ReadBSVString(t, p.ReadFileU32LE(t.FilePath, t.FieldOffset(row, game.ItemFieldDescription))),
		}
	}

	return records
}

// readItemListRecords reads all rows from an item list BSV table, extracting
// name and category fields.
func readItemListRecords(p *Patcher, t BSVTable) []ItemListRecord {
	records := make([]ItemListRecord, int(t.RowCount))

	for row := 0; row < int(t.RowCount); row++ {
		records[row] = ItemListRecord{
			Name:     p.ReadBSVString(t, p.ReadFileU32LE(t.FilePath, t.FieldOffset(row, game.ItemListFieldName))),
			Category: p.ReadBSVString(t, p.ReadFileU32LE(t.FilePath, t.FieldOffset(row, game.ItemListFieldCategory))),
		}
	}

	return records
}

// readOutfitRelationRecords reads all rows from the outfit relation BSV
// table, extracting name, no-condition, and with-condition fields.
func readOutfitRelationRecords(p *Patcher, t BSVTable) []OutfitRelationRecord {
	records := make([]OutfitRelationRecord, int(t.RowCount))

	for row := 0; row < int(t.RowCount); row++ {
		records[row] = OutfitRelationRecord{
			Name:     p.ReadBSVString(t, p.ReadFileU32LE(t.FilePath, t.FieldOffset(row, game.OutfitFieldName))),
			NoCond:   p.ReadBSVString(t, p.ReadFileU32LE(t.FilePath, t.FieldOffset(row, game.OutfitFieldNoCond))),
			WithCond: p.ReadBSVString(t, p.ReadFileU32LE(t.FilePath, t.FieldOffset(row, game.OutfitFieldWithCond))),
		}
	}

	return records
}

// applyMainItemPatches overwrites fields in main item records with English
// text from the patch list. nil patch fields leave the original text untouched.
func applyMainItemPatches(records []ItemTextRecord, patches []ItemTextPatch) {
	seen := map[int]bool{}

	for _, patch := range patches {
		if patch.MainRow < 0 || patch.MainRow >= len(records) {
			log.Fatalf("item patch row out of range: %d", patch.MainRow)
		}
		if seen[patch.MainRow] {
			log.Fatalf("duplicate item patch for main row %d", patch.MainRow)
		}
		seen[patch.MainRow] = true

		record := &records[patch.MainRow]

		if patch.Name != nil {
			record.Name = bsv.EncodePatchText(*patch.Name)
		}
		if patch.Description != nil {
			record.Description = bsv.EncodePatchText(*patch.Description)
		}
		if patch.Info1 != nil {
			record.Info1 = bsv.EncodePatchText(*patch.Info1)
		}
		if patch.Info2 != nil {
			record.Info2 = bsv.EncodePatchText(*patch.Info2)
		}
		if patch.Info3 != nil {
			record.Info3 = bsv.EncodePatchText(*patch.Info3)
		}
	}
}

// mainRowToItemListRow maps main item table rows to item list table rows.
// The two tables use different orderings: main table rows 0-22 (weapons)
// map linearly, but rows 23-33 (ammo) and 11-22 (weapons in list order) are
// interleaved in the item list. Rows 34+ (outfits, accessories) map 1:1.
//
//	Main table layout:  |  Item list layout:
//	0-10   Weapons A    |  0-10   Weapons A (same)
//	11-22  Weapons B    |  11-21  Ammo (from main 23-33)
//	23-33  Ammo         |  22-33  Weapons B (from main 11-22)
//	34+    Outfits/etc  |  34+    Outfits/etc (same)
var mainRowToItemListRow = map[int]int{
	0:  0,
	1:  1,
	2:  2,
	3:  3,
	4:  4,
	5:  5,
	6:  6,
	7:  7,
	8:  8,
	9:  9,
	10: 10,

	23: 11,
	24: 12,
	25: 13,
	26: 14,
	27: 15,
	28: 16,
	29: 17,
	30: 18,
	31: 19,
	32: 20,
	33: 21,

	11: 22,
	12: 23,
	13: 24,
	14: 25,
	15: 26,
	16: 27,
	17: 28,
	18: 29,
	19: 30,
	20: 31,
	21: 32,
	22: 33,

	34: 34,
	35: 35,
	36: 36,
	37: 37,
	38: 38,
	39: 39,
	40: 40,
	41: 41,
	42: 42,
	43: 43,
	44: 44,
	45: 45,
	46: 46,
	47: 47,
	48: 48,
	49: 49,
	50: 50,

	51: 51,
	52: 52,
	53: 53,
	54: 54,
	55: 55,
	56: 56,
	57: 57,
	58: 58,
	59: 59,
	60: 60,
	61: 61,
	62: 62,
	63: 63,
	64: 64,
	65: 65,
	66: 66,
	67: 67,
	68: 68,
	69: 69,
}

// applyItemListPatches updates item list records with English names,
// using the main-to-list row mapping to translate indices.
func applyItemListPatches(records []ItemListRecord, patches []ItemTextPatch) {
	for _, patch := range patches {
		listRow, ok := mainRowToItemListRow[patch.MainRow]
		if !ok {
			continue
		}
		if listRow < 0 || listRow >= len(records) {
			log.Fatalf("item-list row out of range: mainRow=%d listRow=%d", patch.MainRow, listRow)
		}

		name := patch.ListName
		if name == nil {
			name = patch.Name
		}

		if name != nil {
			records[listRow].Name = bsv.EncodePatchText(*name)
		}
	}
}

// writeMainItemRecords rebuilds the main item table's string pool and writes
// all rows back, duplicating the name reference to both name fields.
func writeMainItemRecords(p *Patcher, t BSVTable, records []ItemTextRecord) {
	pw := bsv.NewPoolWriter(p, t)
	bsv.RebuildPool(&pw, p.collectEMANRefs(t), p.collectDNIKRefs(t))

	for row, record := range records {
		nameRel := pw.PutString(record.Name)

		p.WriteFileU32LE(t.FilePath, t.FieldOffset(row, game.ItemFieldName), nameRel)
		p.WriteFileU32LE(t.FilePath, t.FieldOffset(row, game.ItemFieldNameCopy), nameRel)

		p.WriteFileU32LE(t.FilePath, t.FieldOffset(row, game.ItemFieldType), pw.PutString(record.Type))
		p.WriteFileU32LE(t.FilePath, t.FieldOffset(row, game.ItemFieldInfo1), pw.PutString(record.Info1))
		p.WriteFileU32LE(t.FilePath, t.FieldOffset(row, game.ItemFieldInfo2), pw.PutString(record.Info2))
		p.WriteFileU32LE(t.FilePath, t.FieldOffset(row, game.ItemFieldInfo3), pw.PutString(record.Info3))
		p.WriteFileU32LE(t.FilePath, t.FieldOffset(row, game.ItemFieldDescription), pw.PutString(record.Description))
	}
}

// writeItemListRecords rebuilds an item list table's string pool and writes
// all rows back, duplicating the name reference to both name fields.
func writeItemListRecords(p *Patcher, t BSVTable, records []ItemListRecord) {
	pw := bsv.NewPoolWriter(p, t)
	bsv.RebuildPool(&pw, p.collectEMANRefs(t), p.collectDNIKRefs(t))

	for row, record := range records {
		nameRel := pw.PutString(record.Name)

		p.WriteFileU32LE(t.FilePath, t.FieldOffset(row, game.ItemListFieldName), nameRel)
		p.WriteFileU32LE(t.FilePath, t.FieldOffset(row, game.ItemListFieldNameCopy), nameRel)
		p.WriteFileU32LE(t.FilePath, t.FieldOffset(row, game.ItemListFieldCategory), pw.PutString(record.Category))
	}
}

// writeOutfitRelationRecords rebuilds the outfit relation table's string
// pool and writes all rows back.
func writeOutfitRelationRecords(p *Patcher, t BSVTable, records []OutfitRelationRecord) {
	pw := bsv.NewPoolWriter(p, t)
	bsv.RebuildPool(&pw, p.collectEMANRefs(t), p.collectDNIKRefs(t))

	for row, record := range records {
		p.WriteFileU32LE(t.FilePath, t.FieldOffset(row, game.OutfitFieldName), pw.PutString(record.Name))
		p.WriteFileU32LE(t.FilePath, t.FieldOffset(row, game.OutfitFieldNoCond), pw.PutString(record.NoCond))
		p.WriteFileU32LE(t.FilePath, t.FieldOffset(row, game.OutfitFieldWithCond), pw.PutString(record.WithCond))
	}
}


// ItemTexts applies English text patches to all item/weapon/outfit BSV
// tables: main item table, both item list copies, and the outfit relation
// table. Each table's string pool is rebuilt after patching.
func (p *Patcher) ItemTexts(patches []ItemTextPatch) {
	mainTable := p.OpenBSVTable(game.GameUFP, game.BSVMainItem)
	itemList1 := p.OpenBSVTable(game.GameUFP, game.BSVItemList1)
	itemList2 := p.OpenBSVTable(game.GameUFP, game.BSVItemList2)
	outfitTable := p.OpenBSVTable(game.GameUFP, game.BSVOutfitRelation)

	mainRecords := readMainItemRecords(p, mainTable)
	list1Records := readItemListRecords(p, itemList1)
	list2Records := readItemListRecords(p, itemList2)
	outfitRecords := readOutfitRelationRecords(p, outfitTable)

	applyMainItemPatches(mainRecords, patches)
	applyItemListPatches(list1Records, patches)
	applyItemListPatches(list2Records, patches)

	// Optional later: apply outfit-specific patches to outfitRecords.

	writeMainItemRecords(p, mainTable, mainRecords)
	writeItemListRecords(p, itemList1, list1Records)
	writeItemListRecords(p, itemList2, list2Records)
	writeOutfitRelationRecords(p, outfitTable, outfitRecords)
}

func itemDumpText(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	return fromSJIS(raw)
}

// DumpItemTexts prints the current on-disc text for every item/weapon/outfit
// row to stdout for inspection.
func (p *Patcher) DumpItemTexts() {
	mainTable := p.OpenBSVTable(game.GameUFP, game.BSVMainItem)
	itemList1 := p.OpenBSVTable(game.GameUFP, game.BSVItemList1)

	mainRecords := readMainItemRecords(p, mainTable)
	listRecords := readItemListRecords(p, itemList1)

	reverseListRows := make(map[int]int)
	for mainRow, listRow := range mainRowToItemListRow {
		reverseListRows[mainRow] = listRow
	}

	for mainRow, record := range mainRecords {
		fmt.Printf("MainRow: %d\n", mainRow)

		fmt.Printf("Name:        %q\n", itemDumpText(record.Name))

		if listRow, ok := reverseListRows[mainRow]; ok && listRow >= 0 && listRow < len(listRecords) {
			fmt.Printf("ListName:    %q\n", itemDumpText(listRecords[listRow].Name))
		} else {
			fmt.Printf("ListName:    %q\n", "")
		}

		fmt.Printf("Type:        %q\n", itemDumpText(record.Type))
		fmt.Printf("Info1:       %q\n", itemDumpText(record.Info1))
		fmt.Printf("Info2:       %q\n", itemDumpText(record.Info2))
		fmt.Printf("Info3:       %q\n", itemDumpText(record.Info3))
		fmt.Printf("Description: %q\n", itemDumpText(record.Description))
		fmt.Println()
	}
}
