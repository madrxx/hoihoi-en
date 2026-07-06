// Package text provides the English translation data and shared types for
// the Hoihoi-san patch. The types describe the structure of each translation
// patch (items, missions, modals, subtitles), and the data files contain the
// actual English text.
package text

// ---- Shared helper ----

// StrPtr returns a pointer to s. Used when constructing translation patch
// literals where nil means "keep the original text."
func StrPtr(s string) *string {
	return &s
}

// ---- Item types ----

// ItemTextPatch describes an English replacement for a single row in the
// main item table (weapon, outfit, ammo, antenna, accessory, or treasure).
// Pointer fields use nil to mean "keep the original Japanese text."
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

// ItemTextRecord holds the decoded text fields from one main item table row.
type ItemTextRecord struct {
	Name        []byte
	Type        []byte
	Info1       []byte
	Info2       []byte
	Info3       []byte
	Description []byte
}

// ItemListRecord holds the decoded text fields from one item list table row.
type ItemListRecord struct {
	Name     []byte
	Category []byte
}

// OutfitRelationRecord holds the decoded text fields from one outfit
// relation table row.
type OutfitRelationRecord struct {
	Name     []byte
	NoCond   []byte
	WithCond []byte
}

// ---- Mission types ----

// MissionPatch describes an English replacement for a single mission record.
// Index is zero-based. Empty Title/Description means "keep existing."
// ConfirmLines replaces objective text; nil leaves existing untouched,
// empty strings suppress that line. Up to four lines are used.
type MissionPatch struct {
	Index        int
	Title        string
	Description  string
	ConfirmLines []string
}

// MissionRecordText holds the decoded text fields from one mission record.
type MissionRecordText struct {
	Number      []byte
	ID          []byte
	Title       []byte
	Description []byte
	Confirm     [4][]byte
}

// ---- Event modal types ----

// EventModalPatch replaces a known Japanese label in an event BSV table's
// string pool with English text.
type EventModalPatch struct {
	Base    uint64
	OldText string
	NewText string
}

// EventModalSRTSString is a string found in an event BSV table's SRTS pool.
type EventModalSRTSString struct {
	Rel uint32
	Raw []byte
}

// ---- ADV subtitle types ----

// AdvSubtitlePatch is a single ADV cutscene subtitle entry. EventIndex and
// VoiceID identify the voice line to subtitle. Text may contain newlines
// (split into up to 4 subtitle lines at runtime).
type AdvSubtitlePatch struct {
	EventIndex uint32
	VoiceID    uint32
	Text       string
}
