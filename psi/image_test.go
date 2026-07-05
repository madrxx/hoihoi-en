package psi

import (
	"encoding/binary"
	"testing"
)

func TestIsPSI_Valid(t *testing.T) {
	data := []byte("3ISP\x00\x00\x00\x00")
	if !IsPSI(data) {
		t.Error("IsPSI: expected true for '3ISP' magic")
	}
}

func TestIsPSI_Invalid(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"short", []byte("3I")},
		{"wrong magic", []byte("UFV1\x00\x00\x00\x00")},
		{"garbage", []byte("XXXX")},
	}
	for _, tc := range tests {
		if IsPSI(tc.data) {
			t.Errorf("IsPSI(%s): expected false", tc.name)
		}
	}
}

func TestFindChunks(t *testing.T) {
	chunks := []Chunk{
		{Tag: "GAMI", ChunkOffset: 0x10},
		{Tag: "GAMI", ChunkOffset: 0x420},
		{Tag: "KCOE", ChunkOffset: 0x840},
	}
	gamis := FindChunks(chunks, "GAMI")
	if len(gamis) != 2 {
		t.Fatalf("FindChunks(GAMI): got %d, want 2", len(gamis))
	}
	koes := FindChunks(chunks, "KCOE")
	if len(koes) != 1 {
		t.Fatalf("FindChunks(KCOE): got %d, want 1", len(koes))
	}
}

func TestFindChunks_NoMatch(t *testing.T) {
	chunks := []Chunk{
		{Tag: "GAMI", ChunkOffset: 0x10},
	}
	got := FindChunks(chunks, "XXXX")
	if len(got) != 0 {
		t.Errorf("FindChunks(XXXX): got %d, want 0", len(got))
	}
}

func TestFindChunks_Empty(t *testing.T) {
	got := FindChunks(nil, "GAMI")
	if len(got) != 0 {
		t.Errorf("FindChunks on empty: got %d, want 0", len(got))
	}
}

func TestReorderCLUT(t *testing.T) {
	// Create a 1024-byte palette where each entry's first byte is its index
	raw := make([]byte, 1024)
	for i := 0; i < 256; i++ {
		raw[i*4] = byte(i)
		raw[i*4+1] = 0
		raw[i*4+2] = 0
		raw[i*4+3] = 0xFF
	}

	got := reorderCLUT(raw)

	// After reorder, within each 32-entry block:
	// entries 8-15 should come from original entries 16-23
	// entries 16-23 should come from original entries 8-15
	for base := 0; base < 256; base += 32 {
		for i := 0; i < 8; i++ {
			// Entry (base+8+i) should equal original entry (base+16+i)
			gotVal := got[(base+8+i)*4]
			wantVal := byte(base + 16 + i)
			if gotVal != wantVal {
				t.Errorf("reorderCLUT[%d]: got %d, want %d (from pos %d)", base+8+i, gotVal, wantVal, base+16+i)
			}
			// Entry (base+16+i) should equal original entry (base+8+i)
			gotVal2 := got[(base+16+i)*4]
			wantVal2 := byte(base + 8 + i)
			if gotVal2 != wantVal2 {
				t.Errorf("reorderCLUT[%d]: got %d, want %d (from pos %d)", base+16+i, gotVal2, wantVal2, base+8+i)
			}
		}
	}

	// Unswapped entries should remain in place
	for base := 0; base < 256; base += 32 {
		for i := 0; i < 8; i++ {
			gotVal := got[(base+i)*4]
			if gotVal != byte(base+i) {
				t.Errorf("reorderCLUT[%d]: unswapped got %d, want %d", base+i, gotVal, base+i)
			}
		}
		for i := 24; i < 32; i++ {
			gotVal := got[(base+i)*4]
			if gotVal != byte(base+i) {
				t.Errorf("reorderCLUT[%d]: unswapped got %d, want %d", base+i, gotVal, base+i)
			}
		}
	}
}

func TestNormalisePalette(t *testing.T) {
	// Build a 1024-byte palette: entry 0 = solid black (0,0,0,0xFF), entry 1 = solid white (255,255,255,0xFF)
	raw := make([]byte, 1024)
	// entry 0: black with alpha 0xFF
	raw[3] = 0xFF
	// entry 1: white with alpha 0x80 (gets doubled to 0xFF)
	raw[4] = 0xFF
	raw[5] = 0xFF
	raw[6] = 0xFF
	raw[7] = 0x80

	pal := normalisePalette(raw)
	if len(pal) != 256 {
		t.Fatalf("normalisePalette: got %d entries, want 256", len(pal))
	}

	r0, g0, b0, a0 := pal[0].RGBA()
	if r0 != 0 || g0 != 0 || b0 != 0 || a0 != 0xFFFF {
		t.Errorf("entry 0: got RGBA(%d,%d,%d,%d), want (0,0,0,65535)", r0>>8, g0>>8, b0>>8, a0>>8)
	}

	r1, g1, b1, a1 := pal[1].RGBA()
	if r1 != 0xFFFF || g1 != 0xFFFF || b1 != 0xFFFF || a1 != 0xFFFF {
		t.Errorf("entry 1: got RGBA(%d,%d,%d,%d), want (255,255,255,255)", r1>>8, g1>>8, b1>>8, a1>>8)
	}
}

func TestNormalisePalette_AlphaClamp(t *testing.T) {
	// Alpha values <= 0x80 get doubled, but clamped to 255
	raw := make([]byte, 1024)
	raw[3] = 0x90 // alpha 144 - above 0x80, so stays 144 (does not double)

	pal := normalisePalette(raw)
	_, _, _, a := pal[0].RGBA()
	// 0x90 > 0x80, so not doubled: stays 0x90
	if a != uint32(0x90)<<8|uint32(0x90) {
		t.Errorf("alpha: got 0x%04X, want 0x%04X (0x90)", a>>8, uint32(0x90))
	}
}

func TestParseChunks_Basic(t *testing.T) {
	// Build a minimal 3ISP file with two chunks of known size.
	// PSI uses 16-byte alignment for all chunks including the root magic.
	//
	// Layout:
	//   0x00: "3ISP" + 0x00 0x00 0x00 0x00  (8 bytes, root)
	//   0x08: 8 bytes padding                 (to reach 16-byte boundary)
	//   0x10: "GAMI" + size(24) + 24B payload (32 bytes total, ends at 0x30)
	//   0x30: "KCOE" + size(8) + 8B payload   (ends at 0x40)
	payload1 := make([]byte, 24)
	for i := range payload1 {
		payload1[i] = byte(i)
	}
	payload2 := make([]byte, 8)
	for i := range payload2 {
		payload2[i] = byte(0xAA + i)
	}

	var data []byte
	data = append(data, []byte("3ISP")...)
	data = append(data, 0x00, 0x00, 0x00, 0x00) // root size = 0
	data = append(data, make([]byte, 8)...)      // 16-byte alignment after root

	// GAMI chunk at 0x10
	data = append(data, []byte("GAMI")...)
	data = append(data, byte(24), 0x00, 0x00, 0x00) // size = 24
	data = append(data, payload1...)

	// KCOE chunk at 0x30 (already 16-byte aligned: 0x30 % 16 = 0)
	data = append(data, []byte("KCOE")...)
	data = append(data, byte(8), 0x00, 0x00, 0x00)
	data = append(data, payload2...)

	chunks := ParseChunks(data)
	if len(chunks) < 3 {
		t.Fatalf("ParseChunks: got %d chunks, want at least 3", len(chunks))
	}
	// Root "3ISP" chunk comes first (size 0, no payload).
	if chunks[0].Tag != "3ISP" {
		t.Errorf("chunk 0 tag: got %q, want '3ISP'", chunks[0].Tag)
	}
	if chunks[1].Tag != "GAMI" {
		t.Errorf("chunk 1 tag: got %q, want 'GAMI'", chunks[1].Tag)
	}
	if chunks[2].Tag != "KCOE" {
		t.Errorf("chunk 2 tag: got %q, want 'KCOE'", chunks[2].Tag)
	}
	if len(chunks[1].Payload) != 24 {
		t.Errorf("GAMI payload: got len %d, want 24", len(chunks[1].Payload))
	}
}

func TestImageDimensions(t *testing.T) {
	// Valid GAMI image header: PSM=0x1300, unknown 2 bytes, width=64, height=32
	header := make([]byte, 8)
	binary.LittleEndian.PutUint16(header[0:2], 0x1300) // PSM
	binary.LittleEndian.PutUint16(header[4:6], 64)     // width
	binary.LittleEndian.PutUint16(header[6:8], 32)     // height

	w, h := imageDimensions(header)
	if w != 64 {
		t.Errorf("width: got %d, want 64", w)
	}
	if h != 32 {
		t.Errorf("height: got %d, want 32", h)
	}
}
