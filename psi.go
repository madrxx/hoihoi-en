// PSI texture extraction, XOR-based patching, and UFP-level batch
// application. Provides ExtractPSI for dumping textures to indexed PNGs,
// MakePSIXOR for creating XOR deltas from edited PNGs, ApplyPSIPatches for
// batch application (grouped by UFP archive), and RecoverPSI for
// reconstructing textures from disc originals and XOR deltas.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/madrxx/hoihoi-en/asset"
	"github.com/madrxx/hoihoi-en/psi"
	"github.com/madrxx/hoihoi-en/ufp"
)

// ExtractPSI extracts a single PSI texture from a UFP archive and writes it
// as an indexed PNG, with a sidecar JSON recording the original checksum.
func (p *Patcher) ExtractPSI(ufpDiscPath, ufpEntryPath, outPNG string) {
	ufpData := p.ReadWholeFile(ufpDiscPath)
	entries, err := ufp.Parse(ufpData)
	if err != nil {
		log.Fatal(err)
	}

	entry, ok := ufp.FindEntry(entries, ufpEntryPath)
	if !ok {
		log.Fatalf("UFP entry not found: %s", ufpEntryPath)
	}

	psiData := ufp.ReadEntry(ufpData, entry)
	imageInfo := psi.ParseImage(psiData)
	indices := psi.ExtractIndices(psiData, imageInfo)

	psi.WriteIndexedPNG(outPNG, indices, imageInfo)

	sidecar := AssetSidecar{
		Type:               "psi",
		UFPDiscPath:        ufpDiscPath,
		UFPEntryPath:       entry.Path,
		Width:              imageInfo.Width,
		Height:             imageInfo.Height,
		SourceStride:       imageInfo.SourceStride,
		ImageDataOffset:    imageInfo.ImageDataOffset,
		VisiblePixelLength: uint64(imageInfo.Width * imageInfo.Height),
		OriginalSHA256:     asset.SHA256Hex(indices),
	}
	asset.WriteJSON(asset.SidecarPath(outPNG), sidecar)
}

// MakePSIXOR reads an edited PSI PNG, verifies its sidecar checksum against
// the disc original, creates an XOR delta for the pixel indices, and upserts
// it into the manifest.
func (p *Patcher) MakePSIXOR(editedPNG, manifestPath string) {
	var sidecar AssetSidecar
	asset.ReadJSON(asset.SidecarPath(editedPNG), &sidecar)

	if sidecar.Type != "psi" {
		log.Fatalf("sidecar %s is type %q, not psi", asset.SidecarPath(editedPNG), sidecar.Type)
	}

	ufpData := p.ReadWholeFile(sidecar.UFPDiscPath)
	entries, err := ufp.Parse(ufpData)
	if err != nil {
		log.Fatal(err)
	}

	entry, ok := ufp.FindEntry(entries, sidecar.UFPEntryPath)
	if !ok {
		log.Fatalf("UFP entry not found: %s", sidecar.UFPEntryPath)
	}

	psiData := ufp.ReadEntry(ufpData, entry)
	imageInfo := psi.ParseImage(psiData)

	if sidecar.Width != imageInfo.Width ||
		sidecar.Height != imageInfo.Height ||
		sidecar.SourceStride != imageInfo.SourceStride ||
		sidecar.ImageDataOffset != imageInfo.ImageDataOffset ||
		sidecar.VisiblePixelLength != uint64(imageInfo.Width*imageInfo.Height) {
		log.Fatalf("PSI sidecar metadata no longer matches parsed PSI: %s", sidecar.UFPEntryPath)
	}

	originalIndices := psi.ExtractIndices(psiData, imageInfo)
	if got := asset.SHA256Hex(originalIndices); got != sidecar.OriginalSHA256 {
		log.Fatalf("PSI original checksum mismatch for %s: got=%s expected=%s", sidecar.UFPEntryPath, got, sidecar.OriginalSHA256)
	}

	replacementIndices := psi.ReadIndexedPNGIndices(editedPNG, imageInfo.Width, imageInfo.Height)
	xor := asset.MakeXOR(originalIndices, replacementIndices)

	if len(xor) != len(originalIndices) {
		log.Fatalf("PSI xor length mismatch: got=0x%X expected=0x%X", len(xor), len(originalIndices))
	}

	xorPath := asset.DefaultXORPathForPSI(sidecar.UFPEntryPath)
	asset.WriteFile(xorPath, xor)

	manifest := asset.LoadManifest(manifestPath)
	asset.UpsertPSI(&manifest, PSIPatch{
		UFPDiscPath:        sidecar.UFPDiscPath,
		UFPEntryPath:       sidecar.UFPEntryPath,
		XORPath:            filepath.ToSlash(xorPath),
		Width:              imageInfo.Width,
		Height:             imageInfo.Height,
		SourceStride:       imageInfo.SourceStride,
		ImageDataOffset:    imageInfo.ImageDataOffset,
		VisiblePixelLength: uint64(imageInfo.Width * imageInfo.Height),
		OriginalSHA256:     asset.SHA256Hex(originalIndices),
		ReplacementSHA256:  asset.SHA256Hex(replacementIndices),
		XORSHA256:          asset.SHA256Hex(xor),
	})
	asset.SaveManifest(manifestPath, manifest)
}

// ApplyPSIPatchToUFP applies a single PSI XOR patch to a UFP archive's
// pixel indices in-place, verifying all checksums along the way.
func ApplyPSIPatchToUFP(ufpData []byte, entries []ufp.Entry, patch PSIPatch) {
	entry, ok := ufp.FindEntry(entries, patch.UFPEntryPath)
	if !ok {
		log.Fatalf("UFP entry not found: %s", patch.UFPEntryPath)
	}

	psiData := ufp.ReadEntry(ufpData, entry)
	imageInfo := psi.ParseImage(psiData)
	validatePSIMetadata(patch, imageInfo)

	originalIndices := psi.ExtractIndices(psiData, imageInfo)
	if got := asset.SHA256Hex(originalIndices); got != patch.OriginalSHA256 {
		log.Fatalf("PSI original checksum mismatch for %s: got=%s expected=%s", patch.UFPEntryPath, got, patch.OriginalSHA256)
	}

	xor := asset.ReadFile(patch.XORPath)
	if uint64(len(xor)) != patch.VisiblePixelLength {
		log.Fatalf("PSI xor length mismatch for %s: got=0x%X expected=0x%X", patch.XORPath, len(xor), patch.VisiblePixelLength)
	}
	if got := asset.SHA256Hex(xor); got != patch.XORSHA256 {
		log.Fatalf("PSI xor checksum mismatch for %s: got=%s expected=%s", patch.XORPath, got, patch.XORSHA256)
	}

	replacementIndices := asset.ApplyXOR(originalIndices, xor)
	if got := asset.SHA256Hex(replacementIndices); got != patch.ReplacementSHA256 {
		log.Fatalf("PSI replacement checksum mismatch for %s: got=%s expected=%s", patch.UFPEntryPath, got, patch.ReplacementSHA256)
	}

	psi.ReplaceIndices(psiData, imageInfo, replacementIndices)
	ufp.WriteEntry(ufpData, entry, psiData)
}

// ApplyPSIPatches groups PSI patches by UFP file, parses each archive once,
// and applies all patches to the disc image.
func (p *Patcher) ApplyPSIPatches(patches []PSIPatch) {
	byUFP := map[string][]PSIPatch{}

	for _, patch := range patches {
		key := strings.ToLower(patch.UFPDiscPath)
		byUFP[key] = append(byUFP[key], patch)
	}

	for _, group := range byUFP {
		ufpDiscPath := group[0].UFPDiscPath
		ufpData := p.ReadWholeFile(ufpDiscPath)
		entries, err := ufp.Parse(ufpData)
	if err != nil {
		log.Fatal(err)
	}

		seen := map[string]bool{}
		for _, patch := range group {
			key := strings.ToLower(patch.UFPEntryPath)
			if seen[key] {
				log.Fatalf("duplicate PSI patch for %s", patch.UFPEntryPath)
			}
			seen[key] = true

			ApplyPSIPatchToUFP(ufpData, entries, patch)
		}

		p.WriteFileBytes(ufpDiscPath, 0, ufpData)
	}
}

// RecoverPSI reconstructs an edited PSI texture from the disc original and
// the XOR delta recorded in the manifest, writing the result as a PNG.
func (p *Patcher) RecoverPSI(manifestPath, ufpEntryPath, outPNG string) {
	manifest := asset.LoadManifest(manifestPath)

	patch, ok := asset.FindPSI(manifest, ufpEntryPath)
	if !ok {
		log.Fatalf("PSI patch not found for %s in %s", ufpEntryPath, manifestPath)
	}

	ufpData := p.ReadWholeFile(patch.UFPDiscPath)
	entries, err := ufp.Parse(ufpData)
	if err != nil {
		log.Fatal(err)
	}

	entry, ok := ufp.FindEntry(entries, patch.UFPEntryPath)
	if !ok {
		log.Fatalf("UFP entry not found: %s", patch.UFPEntryPath)
	}

	psiData := ufp.ReadEntry(ufpData, entry)
	imageInfo := psi.ParseImage(psiData)
	validatePSIMetadata(patch, imageInfo)

	originalIndices := psi.ExtractIndices(psiData, imageInfo)
	if got := asset.SHA256Hex(originalIndices); got != patch.OriginalSHA256 {
		log.Fatalf("PSI original checksum mismatch for %s: got=%s expected=%s", patch.UFPEntryPath, got, patch.OriginalSHA256)
	}

	xor := asset.ReadFile(patch.XORPath)
	replacementIndices := asset.ApplyXOR(originalIndices, xor)
	if got := asset.SHA256Hex(replacementIndices); got != patch.ReplacementSHA256 {
		log.Fatalf("PSI recovered replacement checksum mismatch for %s: got=%s expected=%s", patch.UFPEntryPath, got, patch.ReplacementSHA256)
	}

	psi.WriteIndexedPNG(outPNG, replacementIndices, imageInfo)

	sidecar := AssetSidecar{
		Type:               "psi",
		UFPDiscPath:        patch.UFPDiscPath,
		UFPEntryPath:       patch.UFPEntryPath,
		Width:              patch.Width,
		Height:             patch.Height,
		SourceStride:       patch.SourceStride,
		ImageDataOffset:    patch.ImageDataOffset,
		VisiblePixelLength: patch.VisiblePixelLength,
		OriginalSHA256:     patch.OriginalSHA256,
	}
	asset.WriteJSON(asset.SidecarPath(outPNG), sidecar)
}

// ExtractAllPSI extracts every PSI texture from a UFP archive as indexed
// PNGs, writing a manifest.tsv index file alongside them.
func (p *Patcher) ExtractAllPSI(ufpDiscPath, outDir string) {
	ufpData := p.ReadWholeFile(ufpDiscPath)
	entries, err := ufp.Parse(ufpData)
	if err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatal(err)
	}

	var manifest []byte
	manifest = append(manifest, []byte("index\tpath\twidth\theight\tpng\n")...)

	for _, entry := range entries {
		if !strings.EqualFold(filepath.Ext(entry.Path), ".psi") {
			continue
		}

		psiData := ufp.ReadEntry(ufpData, entry)

		if !psi.IsPSI(psiData) {
			first := ""
			if len(psiData) >= 4 {
				first = fmt.Sprintf("% X", psiData[:4])
			}
			log.Printf("skipping non-3ISP entry: index=%d path=%s first=%s", entry.Index, entry.Path, first)
			continue
		}

		imageInfo := psi.ParseImage(psiData)
		indices := psi.ExtractIndices(psiData, imageInfo)

		name := fmt.Sprintf(
			"%03d_%s.png",
			entry.Index,
			asset.SafeName(entry.Path),
		)
		outPath := filepath.Join(outDir, name)

		psi.WriteIndexedPNG(outPath, indices, imageInfo)

		manifest = append(
			manifest,
			[]byte(fmt.Sprintf(
				"%d\t%s\t%d\t%d\t%s\n",
				entry.Index,
				entry.Path,
				imageInfo.Width,
				imageInfo.Height,
				name,
			))...,
		)
	}

	asset.WriteFile(filepath.Join(outDir, "manifest.tsv"), manifest)
}

func validatePSIMetadata(patch PSIPatch, imageInfo psi.Image) {
	if patch.Width != imageInfo.Width ||
		patch.Height != imageInfo.Height ||
		patch.SourceStride != imageInfo.SourceStride ||
		patch.ImageDataOffset != imageInfo.ImageDataOffset ||
		patch.VisiblePixelLength != uint64(imageInfo.Width*imageInfo.Height) {
		log.Fatalf(
			"PSI metadata mismatch for %s: manifest=%dx%d stride=%d dataOff=0x%X visible=0x%X parsed=%dx%d stride=%d dataOff=0x%X visible=0x%X",
			patch.UFPEntryPath,
			patch.Width,
			patch.Height,
			patch.SourceStride,
			patch.ImageDataOffset,
			patch.VisiblePixelLength,
			imageInfo.Width,
			imageInfo.Height,
			imageInfo.SourceStride,
			imageInfo.ImageDataOffset,
			uint64(imageInfo.Width*imageInfo.Height),
		)
	}
}
