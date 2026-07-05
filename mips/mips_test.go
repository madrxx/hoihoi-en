package mips

import "testing"

func TestJ(t *testing.T) {
	got := J(0x002597C4)
	want := uint32(0x080965F1)
	if got != want {
		t.Errorf("J(0x002597C4): got 0x%08X, want 0x%08X", got, want)
	}
}

func TestJal(t *testing.T) {
	got := Jal(0x00159844)
	want := uint32(0x0C056611)
	if got != want {
		t.Errorf("Jal(0x00159844): got 0x%08X, want 0x%08X", got, want)
	}
}

func TestLui(t *testing.T) {
	got := Lui(4, 0x3F80)
	want := uint32(0x3C043F80)
	if got != want {
		t.Errorf("Lui(4, 0x3F80): got 0x%08X, want 0x%08X", got, want)
	}
}

func TestOri(t *testing.T) {
	got := Ori(5, 5, 0x0000)
	want := uint32(0x34A50000)
	if got != want {
		t.Errorf("Ori(5, 5, 0): got 0x%08X, want 0x%08X", got, want)
	}
	got = Ori(4, 4, 0x97C4)
	want = uint32(0x348497C4)
	if got != want {
		t.Errorf("Ori(4, 4, 0x97C4): got 0x%08X, want 0x%08X", got, want)
	}
}

func TestAddiu(t *testing.T) {
	got := Addiu(4, 0, 0x10)
	want := uint32(0x24040010)
	if got != want {
		t.Errorf("Addiu(4, 0, 0x10): got 0x%08X, want 0x%08X", got, want)
	}
}

func TestLw(t *testing.T) {
	got := Lw(4, 0x10, 29)
	want := uint32(0x8FA40010)
	if got != want {
		t.Errorf("Lw(4, 0x10, 29): got 0x%08X, want 0x%08X", got, want)
	}
}

func TestSw(t *testing.T) {
	got := Sw(4, 0x10, 29)
	want := uint32(0xAFA40010)
	if got != want {
		t.Errorf("Sw(4, 0x10, 29): got 0x%08X, want 0x%08X", got, want)
	}
}

func TestSb(t *testing.T) {
	got := Sb(5, 0, 4)
	want := uint32(0xA0850000)
	if got != want {
		t.Errorf("Sb(5, 0, 4): got 0x%08X, want 0x%08X", got, want)
	}
}

func TestSh(t *testing.T) {
	got := Sh(6, 2, 5)
	want := uint32(0xA4A60002)
	if got != want {
		t.Errorf("Sh(6, 2, 5): got 0x%08X, want 0x%08X", got, want)
	}
}

func TestLhu(t *testing.T) {
	got := Lhu(3, 8, 7)
	want := uint32(0x94E30008)
	if got != want {
		t.Errorf("Lhu(3, 8, 7): got 0x%08X, want 0x%08X", got, want)
	}
}

func TestAndi(t *testing.T) {
	got := Andi(4, 5, 0xFF)
	want := uint32(0x30A400FF)
	if got != want {
		t.Errorf("Andi(4, 5, 0xFF): got 0x%08X, want 0x%08X", got, want)
	}
}

func TestSll(t *testing.T) {
	got := Sll(4, 5, 2)
	want := uint32(0x00052080)
	if got != want {
		t.Errorf("Sll(4, 5, 2): got 0x%08X, want 0x%08X", got, want)
	}
}

func TestSrl(t *testing.T) {
	got := Srl(4, 5, 2)
	want := uint32(0x00052082)
	if got != want {
		t.Errorf("Srl(4, 5, 2): got 0x%08X, want 0x%08X", got, want)
	}
}

func TestAddu(t *testing.T) {
	got := Addu(4, 5, 6)
	want := uint32(0x00A62021)
	if got != want {
		t.Errorf("Addu(4, 5, 6): got 0x%08X, want 0x%08X", got, want)
	}
}

func TestDaddu(t *testing.T) {
	got := Daddu(4, 5, 6)
	want := uint32(0x00A6202D)
	if got != want {
		t.Errorf("Daddu(4, 5, 6): got 0x%08X, want 0x%08X", got, want)
	}
}

func TestBeq(t *testing.T) {
	got := Beq(4, 0, 5)
	want := uint32(0x10800005)
	if got != want {
		t.Errorf("Beq(4, 0, 5): got 0x%08X, want 0x%08X", got, want)
	}
}

func TestBne(t *testing.T) {
	got := Bne(4, 5, -3)
	want := uint32(0x1485FFFD)
	if got != want {
		t.Errorf("Bne(4, 5, -3): got 0x%08X, want 0x%08X", got, want)
	}
}

func TestRuntimeAddr(t *testing.T) {
	got := RuntimeAddr(0x00159844)
	want := uint32(0x002597C4)
	if got != want {
		t.Errorf("RuntimeAddr(0x00159844): got 0x%08X, want 0x%08X", got, want)
	}
}

func TestBranchOffset(t *testing.T) {
	got := BranchOffset(10, 15)
	want := 4
	if got != want {
		t.Errorf("BranchOffset(10, 15): got %d, want %d", got, want)
	}
	got = BranchOffset(20, 17)
	want = -4
	if got != want {
		t.Errorf("BranchOffset(20, 17): got %d, want %d", got, want)
	}
}

func TestLoadAddr(t *testing.T) {
	buf := NewCodeBuf()
	buf.LoadAddr(4, 0x002597C4)
	words := buf.Words()
	if len(words) != 2 {
		t.Fatalf("LoadAddr: expected 2 words, got %d", len(words))
	}
	wantLUI := uint32(0x3C040025)
	if words[0] != wantLUI {
		t.Errorf("LoadAddr[0] LUI: got 0x%08X, want 0x%08X", words[0], wantLUI)
	}
	wantORI := uint32(0x348497C4)
	if words[1] != wantORI {
		t.Errorf("LoadAddr[1] ORI: got 0x%08X, want 0x%08X", words[1], wantORI)
	}
}

func TestWriteU32sLE(t *testing.T) {
	input := []uint32{0x01020304, 0xDEADBEEF}
	got := WriteU32sLE(input)
	want := []byte{0x04, 0x03, 0x02, 0x01, 0xEF, 0xBE, 0xAD, 0xDE}
	if len(got) != len(want) {
		t.Fatalf("WriteU32sLE: len %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("WriteU32sLE[%d]: got 0x%02X, want 0x%02X", i, got[i], want[i])
		}
	}
}

func TestRegisterRange(t *testing.T) {
	for reg := 0; reg <= 31; reg++ {
		_ = Lui(reg, 0)
		_ = Ori(reg, reg, 0)
		_ = Addiu(reg, reg, 0)
		_ = Lw(reg, 0, reg)
		_ = Sw(reg, 0, reg)
		_ = Sb(reg, 0, reg)
		_ = Sh(reg, 0, reg)
		_ = Lhu(reg, 0, reg)
		_ = Andi(reg, reg, 0)
		_ = Sll(reg, reg, 0)
		_ = Srl(reg, reg, 0)
		_ = Addu(reg, reg, reg)
		_ = Daddu(reg, reg, reg)
		_ = Beq(reg, reg, 0)
		_ = Bne(reg, reg, 0)
	}
}

func TestCodeBuf(t *testing.T) {
	buf := NewCodeBuf()
	buf.Add(Addiu(29, 29, -0x20))
	skip := buf.AddPlaceholder()
	buf.Add(Lw(2, 0, 4))
	target := buf.Here()
	buf.ResolveBeq(skip, 2, 0, target)
	words := buf.Words()
	if len(words) != 3 {
		t.Fatalf("CodeBuf: expected 3 words, got %d", len(words))
	}
	wantBeq := Beq(2, 0, BranchOffset(skip, target))
	if words[skip] != wantBeq {
		t.Errorf("CodeBuf.ResolveBeq: got 0x%08X, want 0x%08X", words[skip], wantBeq)
	}
}
