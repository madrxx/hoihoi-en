package disc

import "testing"

func TestBinToUser_FirstUserDataByte(t *testing.T) {
	// Offset 24 is the first byte of user data in sector 0.
	info := BinToUser(24)
	if !info.IsUserData {
		t.Error("offset 24 should be user data")
	}
	if info.Sector != 0 {
		t.Errorf("sector: got %d, want 0", info.Sector)
	}
	if info.UserOffset != 0 {
		t.Errorf("userOffset: got %d, want 0", info.UserOffset)
	}
}

func TestBinToUser_SyncBytes(t *testing.T) {
	// Offset 0 is the sync pattern  -- not user data.
	info := BinToUser(0)
	if info.IsUserData {
		t.Error("offset 0 should NOT be user data (sync bytes)")
	}
}

func TestBinToUser_LastHeaderByte(t *testing.T) {
	// Offset 23 is the last byte before user data.
	info := BinToUser(23)
	if info.IsUserData {
		t.Error("offset 23 should NOT be user data (mode byte)")
	}
}

func TestBinToUser_SecondSector(t *testing.T) {
	// First user byte of sector 1.
	info := BinToUser(RawSectorSize + Mode2DataStart)
	if !info.IsUserData {
		t.Error("second sector user data should be valid")
	}
	if info.Sector != 1 {
		t.Errorf("sector: got %d, want 1", info.Sector)
	}
	expectedUserOffset := UserSectorSize // 2048 bytes from sector 0
	if info.UserOffset != expectedUserOffset {
		t.Errorf("userOffset: got %d, want %d", info.UserOffset, expectedUserOffset)
	}
}

func TestBinToUser_ECCBytes(t *testing.T) {
	// Bytes at the end of a sector (ECC area) are not user data.
	info := BinToUser(RawSectorSize - 1)
	if info.IsUserData {
		t.Error("ECC bytes should not be user data")
	}
}

func TestUserToBin_RoundTrip(t *testing.T) {
	tests := []uint64{
		Mode2DataStart,                 // sector 0, first user byte
		Mode2DataStart + 100,           // sector 0, mid-sector
		RawSectorSize + Mode2DataStart, // sector 1, first user byte
	}
	for _, binOffset := range tests {
		info := BinToUser(binOffset)
		if !info.IsUserData {
			t.Errorf("BinToUser(0x%X): expected user data", binOffset)
			continue
		}
		got := UserToBin(info.UserOffset)
		if got != binOffset {
			t.Errorf("round-trip: UserToBin(BinToUser(0x%X).UserOffset) = 0x%X", binOffset, got)
		}
	}
}

func TestFileToBin(t *testing.T) {
	// File at LBA 100 with offset 0.
	got := FileToBin(100, 0)
	// LBA 100 * 2048 = user offset 204800, then convert to bin offset:
	// sector = 204800 / 2048 = 100, offset within sector = 0
	// bin = 100*2352 + 24 + 0 = 235224
	want := uint64(100*RawSectorSize + Mode2DataStart)
	if got != want {
		t.Errorf("FileToBin(100, 0): got 0x%X, want 0x%X", got, want)
	}
}

func TestFileToBin_WithOffset(t *testing.T) {
	// File at LBA 50 with file offset 100.
	got := FileToBin(50, 100)
	// user offset = 50*2048 + 100 = 102500
	// sector = 102500 / 2048 = 50, remainder = 100
	// bin = 50*2352 + 24 + 100 = 117724
	want := uint64(50*RawSectorSize + Mode2DataStart + 100)
	if got != want {
		t.Errorf("FileToBin(50, 100): got 0x%X, want 0x%X", got, want)
	}
}

func TestFileToBin_CrossSector(t *testing.T) {
	// File at LBA 0 with an offset that crosses into the next sector.
	got := FileToBin(0, 2048)
	// user offset = 2048, sector = 1, remainder = 0
	// bin = 1*2352 + 24 + 0 = 2376
	want := uint64(RawSectorSize + Mode2DataStart)
	if got != want {
		t.Errorf("FileToBin(0, 2048): got 0x%X, want 0x%X", got, want)
	}
}

func TestCleanISOName_CurrentDir(t *testing.T) {
	got := cleanISOName("\x00")
	if got != "." {
		t.Errorf("cleanISOName(\\x00): got %q, want '.'", got)
	}
}

func TestCleanISOName_ParentDir(t *testing.T) {
	got := cleanISOName("\x01")
	if got != ".." {
		t.Errorf("cleanISOName(\\x01): got %q, want '..'", got)
	}
}

func TestCleanISOName_VersionSuffix(t *testing.T) {
	tests := []struct{ input, want string }{
		{"FILE.EXT;1", "FILE.EXT"},
		{"LONGNAME.;123", "LONGNAME."},
		{"NO.VERSION", "NO.VERSION"},
		{"FILE.EXT;ABC", "FILE.EXT;ABC"}, // non-digit suffix preserved
		{"FOO;1;2", "FOO;1"},             // only last semicolon stripped
	}
	for _, tt := range tests {
		got := cleanISOName(tt.input)
		if got != tt.want {
			t.Errorf("cleanISOName(%q): got %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCleanISOName_NoSemicolon(t *testing.T) {
	got := cleanISOName("README.TXT")
	if got != "README.TXT" {
		t.Errorf("cleanISOName(README.TXT): got %q", got)
	}
}

func TestFindFile(t *testing.T) {
	files := []File{
		{Path: "UFP/GAME.UFP", StartLBA: 100, Size: 500000},
		{Path: "SLPM_623.91", StartLBA: 50, Size: 7000000},
		{Path: "MOVIE/OP.PSS", StartLBA: 200, Size: 50000000},
	}

	// Exact match.
	f, ok := FindFile(files, "SLPM_623.91")
	if !ok {
		t.Fatal("FindFile should find SLPM_623.91")
	}
	if f.StartLBA != 50 {
		t.Errorf("StartLBA: got %d, want 50", f.StartLBA)
	}

	// Case-insensitive.
	f, ok = FindFile(files, "slpm_623.91")
	if !ok {
		t.Error("FindFile should be case-insensitive")
	}

	// Not found.
	_, ok = FindFile(files, "NONEXISTENT.FILE")
	if ok {
		t.Error("FindFile should not find non-existent file")
	}
}

func TestFindFileAtOffset(t *testing.T) {
	files := []File{
		{Path: "FILE_A.BIN", StartLBA: 10, Size: 100},
		{Path: "FILE_B.BIN", StartLBA: 20, Size: 200},
	}

	// User offset 10*2048 + 50 should be in FILE_A.BIN.
	f, _, ok := FindFileAtOffset(files, 10*UserSectorSize+50)
	if !ok {
		t.Fatal("FindFileAtOffset should find file at offset")
	}
	if f.Path != "FILE_A.BIN" {
		t.Errorf("got %q, want FILE_A.BIN", f.Path)
	}

	// User offset 20*2048 should be in FILE_B.BIN.
	f, _, ok = FindFileAtOffset(files, 20*UserSectorSize)
	if !ok {
		t.Fatal("FindFileAtOffset should find file at start of FILE_B")
	}
	if f.Path != "FILE_B.BIN" {
		t.Errorf("got %q, want FILE_B.BIN", f.Path)
	}
}
