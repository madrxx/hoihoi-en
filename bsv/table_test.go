package bsv

import (
	"testing"
)

func TestTableSRTSPoolStart(t *testing.T) {
	tbl := Table{Base: 0x1000}
	got := tbl.SRTSPoolStart()
	want := uint64(0x1030)
	if got != want {
		t.Errorf("SRTSPoolStart: got 0x%X, want 0x%X", got, want)
	}
}

func TestTableSRTSEnd(t *testing.T) {
	tbl := Table{
		Base:       0x1000,
		SRTSOffset: 0x1100,
		SRTSSize:   0x200,
	}
	got := tbl.SRTSEnd()
	want := uint64(0x1100 + 0x08 + 0x200)
	if got != want {
		t.Errorf("SRTSEnd: got 0x%X, want 0x%X", got, want)
	}
}

func TestTableRowOffset(t *testing.T) {
	tbl := Table{
		Base:        0x1000,
		ATADOffset:  0x2000,
		RowCount:    10,
		RowSize:     0x74,
	}
	// Row 0 starts at ATADOffset + 0x10.
	got := tbl.RowOffset(0)
	want := uint64(0x2010)
	if got != want {
		t.Errorf("RowOffset(0): got 0x%X, want 0x%X", got, want)
	}
	// Row 3 starts at ATADOffset + 0x10 + 3*RowSize.
	got = tbl.RowOffset(3)
	want = uint64(0x2010 + 3*0x74)
	if got != want {
		t.Errorf("RowOffset(3): got 0x%X, want 0x%X", got, want)
	}
}

func TestTableFieldOffset(t *testing.T) {
	tbl := Table{
		Base:        0x1000,
		ATADOffset:  0x2000,
		RowCount:    10,
		RowSize:     0x74,
	}
	// Field 3 of row 0: RowOffset(0) + 3*4 = 0x2010 + 0x0C = 0x201C.
	got := tbl.FieldOffset(0, 3)
	want := uint64(0x201C)
	if got != want {
		t.Errorf("FieldOffset(0, 3): got 0x%X, want 0x%X", got, want)
	}
}

func TestStringPoolNew(t *testing.T) {
	tbl := Table{
		Base:       0x1000,
		SRTSOffset: 0x1100,
		SRTSSize:   0x200,
	}
	pool := NewStringPool(tbl)
	if pool.Cursor() != tbl.SRTSPoolStart() {
		t.Errorf("cursor: got 0x%X, want 0x%X", pool.Cursor(), tbl.SRTSPoolStart())
	}
}

func TestStringPoolDedupAndSeen(t *testing.T) {
	tbl := Table{
		Base:       0x1000,
		SRTSOffset: 0x1100,
		SRTSSize:   0x200,
	}
	pool := NewStringPool(tbl)

	if pool.Seen("hello") {
		t.Error("Seen should return false before Dedup")
	}

	pool.Dedup("hello", 0x10)
	if !pool.Seen("hello") {
		t.Error("Seen should return true after Dedup")
	}
}

func TestStringPoolRelFor(t *testing.T) {
	tbl := Table{
		Base:       0x1000,
		SRTSOffset: 0x1100,
		SRTSSize:   0x200,
	}
	pool := NewStringPool(tbl)
	pool.Dedup("hello", 0x10)
	got := pool.RelFor("hello")
	if got != 0x10 {
		t.Errorf("RelFor: got 0x%X, want 0x10", got)
	}
}

func TestStringPoolCursor(t *testing.T) {
	tbl := Table{
		Base:       0x1000,
		SRTSOffset: 0x1100,
		SRTSSize:   0x200,
	}
	pool := NewStringPool(tbl)

	start := pool.Cursor()
	pool.Advance(20)
	if pool.Cursor() != start+20 {
		t.Errorf("after Advance(20): got 0x%X, want 0x%X", pool.Cursor(), start+20)
	}

	pool.SetCursor(0x1200)
	if pool.Cursor() != 0x1200 {
		t.Errorf("after SetCursor: got 0x%X, want 0x1200", pool.Cursor())
	}
}

func TestStringPoolRange(t *testing.T) {
	tbl := Table{
		Base:       0x1000,
		SRTSOffset: 0x1100,
		SRTSSize:   0x400,
	}
	pool := NewStringPoolRange(tbl, 0x1130, 0x1300)
	if pool.Start() != 0x1130 {
		t.Errorf("Start: got 0x%X, want 0x1130", pool.Start())
	}
	if pool.End() != 0x1300 {
		t.Errorf("End: got 0x%X, want 0x1300", pool.End())
	}
}

func TestEncodePatchText_ASCII(t *testing.T) {
	got := EncodePatchText("A")
	if len(got) == 0 {
		t.Fatal("EncodePatchText returned nil/empty")
	}
	// Should be fullwidth-SJIS for 'A' without null terminator.
	if got[len(got)-1] == 0 {
		t.Error("EncodePatchText should NOT have a trailing null byte")
	}
}

func TestEncodePatchText_Empty(t *testing.T) {
	got := EncodePatchText("")
	if got != nil {
		t.Errorf("EncodePatchText(''): got %v, want nil", got)
	}
}

func TestEncodePatchText_Deterministic(t *testing.T) {
	a := EncodePatchText("Hello")
	b := EncodePatchText("Hello")
	if len(a) != len(b) {
		t.Fatal("EncodePatchText is not deterministic")
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatal("EncodePatchText is not deterministic")
		}
	}
}
