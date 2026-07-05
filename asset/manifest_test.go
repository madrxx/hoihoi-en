package asset

import (
	"testing"
)

func TestSHA256Hex_Known(t *testing.T) {
	// SHA-256 of empty data.
	got := SHA256Hex([]byte{})
	// Known SHA-256 of empty string.
	want := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if got != want {
		t.Errorf("SHA256Hex([]): got %s, want %s", got, want)
	}
}

func TestSHA256Hex_Deterministic(t *testing.T) {
	a := SHA256Hex([]byte("hello"))
	b := SHA256Hex([]byte("hello"))
	if a != b {
		t.Error("SHA256Hex is not deterministic")
	}
	if a == SHA256Hex([]byte("Hello")) {
		t.Error("SHA256Hex should produce different hashes for different input")
	}
}

func TestMakeXOR(t *testing.T) {
	original := []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE}
	replacement := []byte{0x11, 0x22, 0x33, 0x44, 0x55}
	got := MakeXOR(original, replacement)
	// XOR each byte: 0xAA^0x11=0xBB, 0xBB^0x22=0x99, etc.
	want := []byte{0xBB, 0x99, 0xFF, 0x99, 0xBB}
	if len(got) != len(want) {
		t.Fatalf("MakeXOR: len %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("MakeXOR[%d]: got 0x%02X, want 0x%02X", i, got[i], want[i])
		}
	}
}

func TestMakeXOR_SameSize(t *testing.T) {
	// Same-size replacement should produce a delta of the same length.
	original := make([]byte, 100)
	replacement := make([]byte, 100)
	for i := range replacement {
		replacement[i] = byte(i)
	}
	got := MakeXOR(original, replacement)
	if len(got) != 100 {
		t.Errorf("MakeXOR: len %d, want 100", len(got))
	}
}

func TestApplyXOR_RoundTrip(t *testing.T) {
	original := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	replacement := []byte{0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11}
	xor := MakeXOR(original, replacement)
	got := ApplyXOR(original, xor)
	if len(got) != len(replacement) {
		t.Fatalf("round-trip: len %d, want %d", len(got), len(replacement))
	}
	for i := range replacement {
		if got[i] != replacement[i] {
			t.Errorf("round-trip[%d]: got 0x%02X, want 0x%02X", i, got[i], replacement[i])
		}
	}
}

func TestZeroPad_Shorter(t *testing.T) {
	data := []byte{1, 2, 3}
	got := ZeroPad(data, 5)
	want := []byte{1, 2, 3, 0, 0}
	if len(got) != 5 {
		t.Fatalf("ZeroPad: len %d, want 5", len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ZeroPad[%d]: got 0x%02X, want 0x%02X", i, got[i], want[i])
		}
	}
}

func TestZeroPad_ExactSize(t *testing.T) {
	data := []byte{1, 2, 3}
	got := ZeroPad(data, 3)
	if len(got) != 3 {
		t.Fatalf("ZeroPad: len %d, want 3", len(got))
	}
	if got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Error("ZeroPad should not modify data when size matches")
	}
}

func TestUpsertPSS(t *testing.T) {
	manifest := &Manifest{Version: 1}
	patch1 := PSSPatch{DiscPath: "MOVIE/OP.PSS", XORPath: "patches/fmv/OP.PSS.xor"}
	patch2 := PSSPatch{DiscPath: "MOVIE/02.PSS", XORPath: "patches/fmv/02.PSS.xor"}

	UpsertPSS(manifest, patch1)
	if len(manifest.PSS) != 1 {
		t.Fatalf("after first upsert: len %d, want 1", len(manifest.PSS))
	}

	UpsertPSS(manifest, patch2)
	if len(manifest.PSS) != 2 {
		t.Fatalf("after second upsert: len %d, want 2", len(manifest.PSS))
	}

	// Upsert with same disc path should replace, not append.
	patch1Updated := PSSPatch{DiscPath: "MOVIE/OP.PSS", XORPath: "updated.xor"}
	UpsertPSS(manifest, patch1Updated)
	if len(manifest.PSS) != 2 {
		t.Errorf("upsert with same path should replace: len %d, want 2", len(manifest.PSS))
	}
	if manifest.PSS[0].XORPath != "updated.xor" {
		t.Errorf("upsert didn't update: got %q", manifest.PSS[0].XORPath)
	}
}

func TestUpsertPSS_CaseInsensitive(t *testing.T) {
	manifest := &Manifest{Version: 1}
	UpsertPSS(manifest, PSSPatch{DiscPath: "MOVIE/OP.PSS", XORPath: "first.xor"})
	UpsertPSS(manifest, PSSPatch{DiscPath: "movie/op.pss", XORPath: "second.xor"})
	if len(manifest.PSS) != 1 {
		t.Errorf("case-insensitive upsert: len %d, want 1", len(manifest.PSS))
	}
	if manifest.PSS[0].XORPath != "second.xor" {
		t.Errorf("case-insensitive upsert didn't update: got %q", manifest.PSS[0].XORPath)
	}
}

func TestFindPSS(t *testing.T) {
	manifest := Manifest{
		Version: 1,
		PSS: []PSSPatch{
			{DiscPath: "MOVIE/OP.PSS", XORPath: "op.xor"},
			{DiscPath: "MOVIE/02.PSS", XORPath: "02.xor"},
		},
	}

	p, ok := FindPSS(manifest, "MOVIE/OP.PSS")
	if !ok {
		t.Fatal("FindPSS should find existing patch")
	}
	if p.XORPath != "op.xor" {
		t.Errorf("FindPSS: got XORPath %q, want 'op.xor'", p.XORPath)
	}

	// Case-insensitive.
	_, ok = FindPSS(manifest, "movie/op.pss")
	if !ok {
		t.Error("FindPSS should be case-insensitive")
	}

	_, ok = FindPSS(manifest, "NONEXISTENT.PSS")
	if ok {
		t.Error("FindPSS should not find non-existent patch")
	}
}

func TestFindPSI(t *testing.T) {
	manifest := Manifest{
		Version: 1,
		PSI: []PSIPatch{
			{UFPEntryPath: "2d/adv/charaentry_pt_001.psi", XORPath: "001.xor"},
			{UFPEntryPath: "2d/title/title_pt_003.psi", XORPath: "003.xor"},
		},
	}

	p, ok := FindPSI(manifest, "2d/title/title_pt_003.psi")
	if !ok {
		t.Fatal("FindPSI should find existing patch")
	}
	if p.XORPath != "003.xor" {
		t.Errorf("FindPSI: got XORPath %q, want '003.xor'", p.XORPath)
	}

	_, ok = FindPSI(manifest, "NONEXISTENT.psi")
	if ok {
		t.Error("FindPSI should not find non-existent patch")
	}
}

func TestSidecarPath(t *testing.T) {
	got := SidecarPath("extracted/OP.pss")
	want := "extracted/OP.pss.hoihoi.json"
	if got != want {
		t.Errorf("SidecarPath: got %q, want %q", got, want)
	}
}

func TestSafeName(t *testing.T) {
	tests := []struct{ input, want string }{
		{"MOVIE/OP.PSS", "MOVIE_OP.PSS"},
		{"path:with:colons", "path_with_colons"},
		{"file<name>", "file_name_"},
		{"back\\slash", "back_slash"}, // backslash -> slash -> underscore
		{"", "asset"},
	}
	for _, tt := range tests {
		got := SafeName(tt.input)
		if got != tt.want {
			t.Errorf("SafeName(%q): got %q, want %q", tt.input, got, tt.want)
		}
	}
}
