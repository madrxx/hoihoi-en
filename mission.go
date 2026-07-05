// Mission text patching via BSV tables. Reads mission records from
// GAME.UFP's mission BSV table, applies English title, description, and
// confirmation line replacements, rebuilds the string pool, and writes
// all records back. Also provides DumpMissionTexts for inspection.
package main

import (
	"fmt"
	"log"

	"github.com/madrxx/hoihoi-en/bsv"
	"github.com/madrxx/hoihoi-en/game"
)

const (
	// Mission record data starts at this offset within the mission BSV.
	missionRecordStart uint64 = 0x1D99D4
)

type MissionPatch struct {
	// Zero-based mission index.
	Index int

	// Mission title. If empty, keep existing title.
	Title string
	// Mission title. If empty, keep existing description.
	Description string

	// Mission objectives. nil keeps all existing text, non-nil replaces all lines. Up to four lines are used, missing lines become blank.
	ConfirmLines []string
}

type missionRecordText struct {
	Number      []byte
	ID          []byte
	Title       []byte
	Description []byte
	Confirm     [4][]byte
}

func missionRecordOffset(index int) uint64 {
	return missionRecordStart + uint64(index)*game.MissionRecordSize
}

func missionFieldRelOffset(p *Patcher, index int, fieldOffset uint64) uint32 {
	return p.ReadFileU32LE(game.GameUFP, missionRecordOffset(index)+fieldOffset)
}

func readMissionRawString(p *Patcher, table BSVTable, relOffset uint32) []byte {
	return p.ReadBSVString(table, relOffset)
}

// readMissionRecords reads all mission records from the mission BSV table,
// extracting number, ID, title, description, and four confirmation lines.
func readMissionRecords(p *Patcher, table BSVTable) []missionRecordText {
	records := make([]missionRecordText, game.MissionRecordCount)

	for i := 0; i < game.MissionRecordCount; i++ {
		records[i] = missionRecordText{
			Number:      readMissionRawString(p, table, missionFieldRelOffset(p, i, game.MissionFieldNumber)),
			ID:          readMissionRawString(p, table, missionFieldRelOffset(p, i, game.MissionFieldID)),
			Title:       readMissionRawString(p, table, missionFieldRelOffset(p, i, game.MissionFieldTitle)),
			Description: readMissionRawString(p, table, missionFieldRelOffset(p, i, game.MissionFieldDescription)),
			Confirm: [4][]byte{
				readMissionRawString(p, table, missionFieldRelOffset(p, i, game.MissionFieldConfirm1)),
				readMissionRawString(p, table, missionFieldRelOffset(p, i, game.MissionFieldConfirm2)),
				readMissionRawString(p, table, missionFieldRelOffset(p, i, game.MissionFieldConfirm3)),
				readMissionRawString(p, table, missionFieldRelOffset(p, i, game.MissionFieldConfirm4)),
			},
		}
	}

	return records
}


// applyMissionPatches overwrites mission record fields with English text
// from the patch list. Empty Title/Description and nil ConfirmLines leave
// the original text untouched.
func applyMissionPatches(records []missionRecordText, patches []MissionPatch) {
	seen := make(map[int]bool)

	for _, patch := range patches {
		if patch.Index < 0 || patch.Index >= len(records) {
			log.Fatalf("mission patch index out of range: %d", patch.Index)
		}

		if seen[patch.Index] {
			log.Fatalf("duplicate mission patch for index %d", patch.Index)
		}
		seen[patch.Index] = true

		record := &records[patch.Index]

		if patch.Title != "" {
			record.Title = bsv.EncodePatchText(patch.Title)
		}

		if patch.Description != "" {
			record.Description = bsv.EncodePatchText(patch.Description)
		}

		if patch.ConfirmLines != nil {
			if len(patch.ConfirmLines) > 4 {
				log.Fatalf(
					"mission patch index %d has %d confirmation lines; maximum is 4",
					patch.Index,
					len(patch.ConfirmLines),
				)
			}

			for i := 0; i < 4; i++ {
				if i < len(patch.ConfirmLines) {
					record.Confirm[i] = bsv.EncodePatchText(patch.ConfirmLines[i])
				} else {
					record.Confirm[i] = nil
				}
			}
		}
	}
}

func writeMissionStringOffset(p *Patcher, index int, fieldOffset uint64, relOffset uint32) {
	p.WriteFileU32LE(game.GameUFP, missionRecordOffset(index)+fieldOffset, relOffset)
}

// writeMissionRecords rebuilds the mission table's string pool (clearing
// only the mission text sub-range, not column names/enums or tail data) and
// writes all rows back.
func writeMissionRecords(p *Patcher, table BSVTable, records []missionRecordText) {
	// Clear only the mission text area.
	//
	// Do not clear the whole SRTS payload: earlier column names/enums and later
	// non-mission strings should remain intact.
	p.WriteFileBytes(game.GameUFP, game.MissionTextStart, make([]byte, game.MissionTextEnd-game.MissionTextStart))

	pw := bsv.NewPoolWriterRange(p, table, game.MissionTextStart, game.MissionTextEnd)

	for i, record := range records {
		writeMissionStringOffset(p, i, game.MissionFieldNumber, pw.PutString(record.Number))
		writeMissionStringOffset(p, i, game.MissionFieldID, pw.PutString(record.ID))
		writeMissionStringOffset(p, i, game.MissionFieldTitle, pw.PutString(record.Title))
		writeMissionStringOffset(p, i, game.MissionFieldDescription, pw.PutString(record.Description))
		writeMissionStringOffset(p, i, game.MissionFieldConfirm1, pw.PutString(record.Confirm[0]))
		writeMissionStringOffset(p, i, game.MissionFieldConfirm2, pw.PutString(record.Confirm[1]))
		writeMissionStringOffset(p, i, game.MissionFieldConfirm3, pw.PutString(record.Confirm[2]))
		writeMissionStringOffset(p, i, game.MissionFieldConfirm4, pw.PutString(record.Confirm[3]))
	}
}

// MissionTexts applies English title, description, and confirmation line
// patches to the mission BSV table.
func (p *Patcher) MissionTexts(patches []MissionPatch) {
	table := p.OpenBSVTable(game.GameUFP, game.BSVMission)

	records := readMissionRecords(p, table)
	applyMissionPatches(records, patches)
	writeMissionRecords(p, table, records)
}

func missionDumpText(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}

	return fromSJIS(raw)
}

// DumpMissionTexts prints the current on-disc text for every mission record
// to stdout for inspection.
func (p *Patcher) DumpMissionTexts() {
	table := p.OpenBSVTable(game.GameUFP, game.BSVMission)
	records := readMissionRecords(p, table)

	for i, record := range records {
		fmt.Printf("Index: %d", i)

		number := missionDumpText(record.Number)
		if number != "" {
			fmt.Printf(" // %s", number)
		}

		fmt.Println()
		fmt.Printf("Title: %q\n", missionDumpText(record.Title))
		fmt.Printf("Description: %q\n", missionDumpText(record.Description))
		fmt.Println("ConfirmLines:")

		for _, line := range record.Confirm {
			fmt.Printf("  %q\n", missionDumpText(line))
		}

		fmt.Println()
	}
}
