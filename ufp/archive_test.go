package ufp

import (
	"encoding/binary"
	"testing"
)

func TestU32(t *testing.T) {
	data := []byte{0x78, 0x56, 0x34, 0x12}
	got := U32(data, 0)
	want := uint32(0x12345678)
	if got != want {
		t.Errorf("U32: got 0x%08X, want 0x%08X", got, want)
	}
}

func TestU32_Offset(t *testing.T) {
	data := []byte{0x00, 0x00, 0xEF, 0xBE, 0xAD, 0xDE}
	got := U32(data, 2)
	want := uint32(0xDEADBEEF)
	if got != want {
		t.Errorf("U32(offset=2): got 0x%08X, want 0x%08X", got, want)
	}
}

func TestCleanPath(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"2d/menu/icon.psi", "2d/menu/icon.psi"},
		{`2d\menu\icon.psi`, "2d/menu/icon.psi"},
		{"/2d/menu/", "2d/menu"},
		{".", ""},
		{"", ""},
		{"2d//menu//icon.psi", "2d/menu/icon.psi"},
		{"2d/./menu/icon.psi", "2d/menu/icon.psi"},
		{`mixed/slash\backslash`, "mixed/slash/backslash"},
		{"space test/file.psi", "space test/file.psi"},
	}
	for _, tc := range tests {
		got := cleanPath(tc.input)
		if got != tc.want {
			t.Errorf("cleanPath(%q): got %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestFindEntry_ExactMatch(t *testing.T) {
	entries := []Entry{
		{Index: 0, Path: "2d/menu/icon.psi", Offset: 0x1000, Size: 0x400},
		{Index: 1, Path: "2d/hud/bar.psi", Offset: 0x2000, Size: 0x200},
		{Index: 2, Path: "3d/char/model.psi", Offset: 0x3000, Size: 0x800},
	}
	entry, ok := FindEntry(entries, "2d/hud/bar.psi")
	if !ok {
		t.Fatal("FindEntry: expected to find entry")
	}
	if entry.Index != 1 {
		t.Errorf("FindEntry: got index %d, want 1", entry.Index)
	}
	if entry.Offset != 0x2000 {
		t.Errorf("FindEntry: got offset 0x%X, want 0x2000", entry.Offset)
	}
}

func TestFindEntry_CaseInsensitive(t *testing.T) {
	entries := []Entry{
		{Index: 0, Path: "2d/MENU/Icon.PSI", Offset: 0x1000, Size: 0x400},
	}
	_, ok := FindEntry(entries, "2d/menu/icon.psi")
	if !ok {
		t.Fatal("FindEntry: case-insensitive lookup failed")
	}
}

func TestFindEntry_PSIExtensionFallback(t *testing.T) {
	entries := []Entry{
		{Index: 0, Path: "2d/menu/icon.psi", Offset: 0x1000, Size: 0x400},
	}
	entry, ok := FindEntry(entries, "2d/menu/icon")
	if !ok {
		t.Fatal("FindEntry: .psi fallback should find the entry")
	}
	if entry.Path != "2d/menu/icon.psi" {
		t.Errorf("FindEntry: got path %q, want '2d/menu/icon.psi'", entry.Path)
	}
}

func TestFindEntry_NotFound(t *testing.T) {
	entries := []Entry{
		{Index: 0, Path: "2d/menu/icon.psi", Offset: 0x1000, Size: 0x400},
	}
	_, ok := FindEntry(entries, "nonexistent.psi")
	if ok {
		t.Error("FindEntry: should not find nonexistent entry")
	}
}

func TestFindEntry_EmptyEntries(t *testing.T) {
	_, ok := FindEntry(nil, "anything.psi")
	if ok {
		t.Error("FindEntry: should not find anything in empty entries")
	}
}

func TestReadEntry(t *testing.T) {
	ufpData := make([]byte, 0x2000)
	for i := range ufpData {
		ufpData[i] = byte(i & 0xFF)
	}
	entry := Entry{Path: "test.psi", Offset: 0x1000, Size: 0x10}
	got := ReadEntry(ufpData, entry)
	if len(got) != 0x10 {
		t.Fatalf("ReadEntry: got len 0x%X, want 0x10", len(got))
	}
	for i := 0; i < 0x10; i++ {
		want := byte((0x1000 + i) & 0xFF)
		if got[i] != want {
			t.Errorf("ReadEntry[%d]: got 0x%02X, want 0x%02X", i, got[i], want)
		}
	}
}

func TestWriteEntry(t *testing.T) {
	ufpData := make([]byte, 0x2000)
	entry := Entry{Path: "test.psi", Offset: 0x100, Size: 0x10}
	replacement := []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99}
	WriteEntry(ufpData, entry, replacement)
	got := ufpData[0x100 : 0x100+0x10]
	for i := range replacement {
		if got[i] != replacement[i] {
			t.Errorf("WriteEntry[%d]: got 0x%02X, want 0x%02X", i, got[i], replacement[i])
		}
	}
}

func TestBuildPath_Basic(t *testing.T) {
	// String table: index 0 = root "2d", index 1 = "menu", index 2 = "icon", index 3 = ".psi"
	stringsTable := []string{"2d", "menu", "icon", ".psi"}
	// meta parent=0 (no extra parent), then ids: menu=1, icon=2, .psi=3
	got := buildPath(stringsTable, 0, []uint16{1, 2, 3})
	want := "2d/menu/icon.psi"
	if got != want {
		t.Errorf("buildPath: got %q, want %q", got, want)
	}
}

func TestBuildPath_ParentIndex(t *testing.T) {
	// meta high 16 bits = 4 means string table entry 4 is a parent dir
	stringsTable := []string{"2d", "weapon", "sword", ".psi", "items"}
	got := buildPath(stringsTable, 4<<16, []uint16{1, 2, 3})
	want := "2d/items/weapon/sword.psi"
	if got != want {
		t.Errorf("buildPath: got %q, want %q", got, want)
	}
}

func TestBuildPath_SkipsRootReindex(t *testing.T) {
	// ids containing 0 (root index) should skip it
	stringsTable := []string{"2d", "hud", "bar", ".psi"}
	got := buildPath(stringsTable, 0, []uint16{1, 0, 2, 3})
	want := "2d/hud/bar.psi"
	if got != want {
		t.Errorf("buildPath: got %q, want %q", got, want)
	}
}

func TestBuildPath_NoExtension(t *testing.T) {
	stringsTable := []string{"2d", "menu", "icon"}
	got := buildPath(stringsTable, 0, []uint16{1, 2})
	want := "2d/menu/icon"
	if got != want {
		t.Errorf("buildPath: got %q, want %q", got, want)
	}
}

func TestBuildPath_EmptyRoot(t *testing.T) {
	stringsTable := []string{"", "menu", "icon", ".psi"}
	got := buildPath(stringsTable, 0, []uint16{1, 2, 3})
	want := "menu/icon.psi"
	if got != want {
		t.Errorf("buildPath: got %q, want %q", got, want)
	}
}

func TestParseStringTable(t *testing.T) {
	// Build a synthetic string table payload:
	// String 0: "2d" (len=2)
	// String 1: "menu" (len=4)
	// Each string: 1 byte length + data
	stringPayload := []byte{2, '2', 'd', 4, 'm', 'e', 'n', 'u'}
	// Offset table: offset of each string in stringPayload
	offsetPayload := make([]byte, 8)
	binary.LittleEndian.PutUint32(offsetPayload[0:4], 0)
	binary.LittleEndian.PutUint32(offsetPayload[4:8], 3)

	table, err := parseStringTable(offsetPayload, stringPayload)
	if err != nil {
		t.Fatalf("parseStringTable: unexpected error: %v", err)
	}
	if len(table) != 2 {
		t.Fatalf("parseStringTable: got %d strings, want 2", len(table))
	}
	if table[0] != "2d" {
		t.Errorf("parseStringTable[0]: got %q, want '2d'", table[0])
	}
	if table[1] != "menu" {
		t.Errorf("parseStringTable[1]: got %q, want 'menu'", table[1])
	}
}

func TestParseOffsetTable(t *testing.T) {
	payload := make([]byte, 12)
	binary.LittleEndian.PutUint32(payload[0:4], 0x100)
	binary.LittleEndian.PutUint32(payload[4:8], 0x200)
	binary.LittleEndian.PutUint32(payload[8:12], 0x300)

	got := parseOffsetTable(payload, "TEST")
	if len(got) != 3 {
		t.Fatalf("parseOffsetTable: got %d entries, want 3", len(got))
	}
	want := []uint32{0x100, 0x200, 0x300}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("parseOffsetTable[%d]: got 0x%X, want 0x%X", i, got[i], want[i])
		}
	}
}

func TestParseOffsetTable_Empty(t *testing.T) {
	got := parseOffsetTable([]byte{}, "EMPTY")
	if len(got) != 0 {
		t.Errorf("parseOffsetTable: expected empty, got %d entries", len(got))
	}
}
