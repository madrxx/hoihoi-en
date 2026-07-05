package cave

import (
	"strings"
	"testing"
)

func TestRegion_Found(t *testing.T) {
	layout := Layout{
		Start: 0x1000,
		End:   0x2000,
		Regions: []Region{
			{Name: "alpha", Offset: 0x1000, Size: 0x400},
			{Name: "beta", Offset: 0x1400, Size: 0x200},
			{Name: "gamma", Offset: 0x1600, Size: 0xA00},
		},
	}
	r := layout.Region("beta")
	if r.Name != "beta" {
		t.Errorf("Region: got name %q, want 'beta'", r.Name)
	}
	if r.Offset != 0x1400 {
		t.Errorf("Region: got offset 0x%X, want 0x1400", r.Offset)
	}
	if r.Size != 0x200 {
		t.Errorf("Region: got size 0x%X, want 0x200", r.Size)
	}
}

func TestRegion_NotFound(t *testing.T) {
	layout := Layout{
		Start:   0x1000,
		End:     0x2000,
		Regions: []Region{{Name: "alpha", Offset: 0x1000, Size: 0x400}},
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Error("Region: expected panic for missing region")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "nonexistent") {
			t.Errorf("Region: panic message should mention region name, got: %v", r)
		}
	}()
	layout.Region("nonexistent")
}

func TestVerify_Valid(t *testing.T) {
	layout := Layout{
		Start: 0x1000,
		End:   0x2000,
		Regions: []Region{
			{Name: "alpha", Offset: 0x1000, Size: 0x400},
			{Name: "beta", Offset: 0x1400, Size: 0x200},
			{Name: "gamma", Offset: 0x1600, Size: 0xA00},
		},
	}
	// Should not panic
	layout.Verify()
}

func TestVerify_ExactFit(t *testing.T) {
	layout := Layout{
		Start: 0x1000,
		End:   0x2000,
		Regions: []Region{
			{Name: "full", Offset: 0x1000, Size: 0x1000},
		},
	}
	layout.Verify()
}

func TestVerify_BeforeStart(t *testing.T) {
	layout := Layout{
		Start: 0x1000,
		End:   0x2000,
		Regions: []Region{
			{Name: "bad", Offset: 0x0FFF, Size: 0x100},
		},
	}
	defer func() {
		r := recover()
		if r == nil {
			t.Error("Verify: expected panic for region before start")
		}
	}()
	layout.Verify()
}

func TestVerify_BeyondEnd(t *testing.T) {
	layout := Layout{
		Start: 0x1000,
		End:   0x2000,
		Regions: []Region{
			{Name: "bad", Offset: 0x1F00, Size: 0x200},
		},
	}
	defer func() {
		r := recover()
		if r == nil {
			t.Error("Verify: expected panic for region extending beyond end")
		}
	}()
	layout.Verify()
}

func TestVerify_Overlap(t *testing.T) {
	layout := Layout{
		Start: 0x1000,
		End:   0x2000,
		Regions: []Region{
			{Name: "alpha", Offset: 0x1000, Size: 0x400},
			{Name: "beta", Offset: 0x1300, Size: 0x400},
		},
	}
	defer func() {
		r := recover()
		if r == nil {
			t.Error("Verify: expected panic for overlapping regions")
		}
	}()
	layout.Verify()
}

func TestVerify_Adjacent(t *testing.T) {
	// Adjacent regions (end of A == start of B) should NOT overlap
	layout := Layout{
		Start: 0x1000,
		End:   0x2000,
		Regions: []Region{
			{Name: "alpha", Offset: 0x1000, Size: 0x400},
			{Name: "beta", Offset: 0x1400, Size: 0xC00},
		},
	}
	layout.Verify()
}

func TestAdvSubtitleLayout_Verify(t *testing.T) {
	AdvSubtitleLayout.Verify()
}
