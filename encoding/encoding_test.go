package encoding

import (
	"testing"
)

func TestToFullWidth_ASCII(t *testing.T) {
	got := ToFullWidth("A")
	if len(got) == 0 {
		t.Fatal("ToFullWidth returned empty string")
	}
	r := []rune(got)[0]
	if r != 0xFF21 {
		t.Errorf("ToFullWidth('A'): got U+%04X, want U+FF21", r)
	}
}

func TestToFullWidth_Space(t *testing.T) {
	got := ToFullWidth(" ")
	r := []rune(got)[0]
	if r != 0x3000 {
		t.Errorf("ToFullWidth(' '): got U+%04X, want U+3000", r)
	}
}

func TestToFullWidth_Yen(t *testing.T) {
	got := ToFullWidth("¥")
	r := []rune(got)[0]
	if r != 0xFFE5 {
		t.Errorf("ToFullWidth('¥'): got U+%04X, want U+FFE5", r)
	}
}

func TestToFullWidth_PassThrough(t *testing.T) {
	tests := []string{
		"\n",
		"テスト",
		"漢字",
	}
	for _, input := range tests {
		got := ToFullWidth(input)
		if got != input {
			t.Errorf("ToFullWidth(%q): got %q, want unchanged", input, got)
		}
	}
}

func TestToFullWidth_Mixed(t *testing.T) {
	got := ToFullWidth("Hello テスト")
	runes := []rune(got)
	if len(runes) != 9 {
		t.Fatalf("ToFullWidth mixed: expected 9 runes, got %d: %q", len(runes), got)
	}
	for i := 0; i < 5; i++ {
		if runes[i] < 0xFF01 || runes[i] > 0xFF5E {
			t.Errorf("rune %d: got U+%04X, expected fullwidth ASCII", i, runes[i])
		}
	}
	if runes[5] != 0x3000 {
		t.Errorf("space: got U+%04X, want U+3000", runes[5])
	}
	if runes[6] != 'テ' || runes[7] != 'ス' || runes[8] != 'ト' {
		t.Errorf("kana: got %q, want テスト", string(runes[6:]))
	}
}

func TestToSJIS_ASCII(t *testing.T) {
	got, err := ToSJIS("ABC", 0)
	if err != nil {
		t.Fatalf("ToSJIS('ABC', 0): unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("ToSJIS('ABC', 0): len %d, want 3", len(got))
	}
	if got[0] != 'A' || got[1] != 'B' || got[2] != 'C' {
		t.Errorf("ToSJIS('ABC', 0): got %v, want [65 66 67]", got)
	}
}

func TestToSJIS_NullBytes(t *testing.T) {
	got, err := ToSJIS("A", 2)
	if err != nil {
		t.Fatalf("ToSJIS('A', 2): unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("ToSJIS('A', 2): len %d, want 3", len(got))
	}
	if got[1] != 0 || got[2] != 0 {
		t.Errorf("ToSJIS('A', 2): last two bytes should be null, got %v", got[1:])
	}
}

func TestToSJIS_ZeroNullBytes(t *testing.T) {
	got, err := ToSJIS("X", 0)
	if err != nil {
		t.Fatalf("ToSJIS('X', 0): unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("ToSJIS('X', 0): len %d, want 1", len(got))
	}
}

func TestToSJIS_Fullwidth(t *testing.T) {
	fullwidthA := ToFullWidth("A")
	got, err := ToSJIS(fullwidthA, 0)
	if err != nil {
		t.Fatalf("ToSJIS(fullwidth 'A', 0): unexpected error: %v", err)
	}
	if len(got) < 2 {
		t.Errorf("ToSJIS(fullwidth 'A', 0): len %d, expected 2+ bytes", len(got))
	}
}

func TestToSJIS_InvalidNullBytes(t *testing.T) {
	_, err := ToSJIS("A", -1)
	if err == nil {
		t.Error("ToSJIS with negative nullBytes should return an error")
	}
}

func TestFromSJIS_ASCII(t *testing.T) {
	input := []byte{'H', 'e', 'l', 'l', 'o'}
	got, err := FromSJIS(input)
	if err != nil {
		t.Fatalf("FromSJIS: unexpected error: %v", err)
	}
	if got != "Hello" {
		t.Errorf("FromSJIS: got %q, want 'Hello'", got)
	}
}

func TestFromSJIS_RoundTrip(t *testing.T) {
	original := "Hello World"
	encoded, err := ToSJIS(original, 0)
	if err != nil {
		t.Fatalf("ToSJIS: unexpected error: %v", err)
	}
	decoded, err := FromSJIS(encoded)
	if err != nil {
		t.Fatalf("FromSJIS: unexpected error: %v", err)
	}
	if decoded != original {
		t.Errorf("round-trip: got %q, want %q", decoded, original)
	}
}

func TestGameText_Output(t *testing.T) {
	got := GameText("A")
	if len(got) == 0 {
		t.Fatal("GameText returned empty slice")
	}
	if got[len(got)-1] != 0 {
		t.Errorf("GameText: last byte is 0x%02X, want 0x00", got[len(got)-1])
	}
}

func TestGameText_Empty(t *testing.T) {
	got := GameText("")
	if len(got) != 1 {
		t.Fatalf("GameText(''): len %d, want 1", len(got))
	}
	if got[0] != 0 {
		t.Errorf("GameText(''): got 0x%02X, want 0x00", got[0])
	}
}

func TestGameText_Deterministic(t *testing.T) {
	a := GameText("Hello")
	b := GameText("Hello")
	if len(a) != len(b) {
		t.Fatalf("GameText is not deterministic: len %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("GameText is not deterministic: byte %d differs", i)
		}
	}
}

func TestGameText_RoundTrip(t *testing.T) {
	original := "Hello"
	encoded := GameText(original)
	sjisBytes := encoded[:len(encoded)-1]
	decoded, err := FromSJIS(sjisBytes)
	if err != nil {
		t.Fatalf("FromSJIS: unexpected error: %v", err)
	}
	expected := ToFullWidth(original)
	if decoded != expected {
		t.Errorf("GameText round-trip: got %q, want %q", decoded, expected)
	}
}
