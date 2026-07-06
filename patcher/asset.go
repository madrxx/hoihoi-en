// Asset patching  -- PSS and PSI extraction, XOR delta creation, and
// manifest-driven application. Provides ExtractPSS/ExtractPSI for dumping
// assets with sidecar checksums, MakePSSXOR/MakePSIXOR for creating XOR
// deltas from edited files, and ApplyPSSPatch/ApplyPSIPatches/RecoverPSS/
// RecoverPSI for applying and reconstructing assets.

package patcher

import (
	"log"
	"os"

	"github.com/madrxx/hoihoi-en/asset"
)

// Type aliases re-export asset package types into the main package so call
// sites in other root-package files can use short names without importing asset.
type (
	AssetManifest = asset.Manifest
	PSSPatch      = asset.PSSPatch
	PSIPatch      = asset.PSIPatch
	AssetSidecar  = asset.Sidecar
)

const defaultAssetManifestPath = asset.DefaultManifestPath

// ExtractPSS reads a PSS (FMV) file from the disc image and writes it to
// outPath, along with a sidecar JSON file recording the original checksum.
func (p *Patcher) ExtractPSS(discPath, outPath string) {
	original := p.ReadWholeFile(discPath)
	asset.WriteFile(outPath, original)

	sidecar := AssetSidecar{
		Type:           "pss",
		DiscPath:       discPath,
		OriginalSize:   uint64(len(original)),
		OriginalSHA256: asset.SHA256Hex(original),
	}
	asset.WriteJSON(asset.SidecarPath(outPath), sidecar)
}

// MakePSSXOR reads an edited PSS file, verifies its sidecar checksum against
// the disc original, creates an XOR delta, and upserts it into the manifest.
func (p *Patcher) MakePSSXOR(editedPath, manifestPath string) {
	var sidecar AssetSidecar
	asset.ReadJSON(asset.SidecarPath(editedPath), &sidecar)

	if sidecar.Type != "pss" {
		log.Fatalf("sidecar %s is type %q, not pss", asset.SidecarPath(editedPath), sidecar.Type)
	}

	original := p.ReadWholeFile(sidecar.DiscPath)
	if uint64(len(original)) != sidecar.OriginalSize {
		log.Fatalf("original PSS size mismatch for %s: got=0x%X expected=0x%X", sidecar.DiscPath, len(original), sidecar.OriginalSize)
	}
	if got := asset.SHA256Hex(original); got != sidecar.OriginalSHA256 {
		log.Fatalf("original PSS checksum mismatch for %s: got=%s expected=%s", sidecar.DiscPath, got, sidecar.OriginalSHA256)
	}

	replacement := asset.ReadFile(editedPath)
	xor := asset.MakeXOR(original, replacement)
	xorPath := asset.DefaultXORPathForPSS(sidecar.DiscPath)
	asset.WriteFile(xorPath, xor)

	manifest := asset.LoadManifest(manifestPath)
	asset.UpsertPSS(&manifest, PSSPatch{
		DiscPath:          sidecar.DiscPath,
		XORPath:           xorPath,
		OriginalSize:      uint64(len(original)),
		ReplacementSize:   uint64(len(replacement)),
		OriginalSHA256:    asset.SHA256Hex(original),
		ReplacementSHA256: asset.SHA256Hex(replacement),
		XORSHA256:         asset.SHA256Hex(xor),
	})
	asset.SaveManifest(manifestPath, manifest)
}

// ApplyPSSPatch applies a single PSS XOR patch to the disc image after
// verifying all checksums (original, XOR delta, recovered replacement).
func (p *Patcher) ApplyPSSPatch(patch PSSPatch) {
	original := p.ReadWholeFile(patch.DiscPath)

	if uint64(len(original)) != patch.OriginalSize {
		log.Fatalf("PSS size mismatch for %s: got=0x%X expected=0x%X", patch.DiscPath, len(original), patch.OriginalSize)
	}
	if got := asset.SHA256Hex(original); got != patch.OriginalSHA256 {
		log.Fatalf("PSS original checksum mismatch for %s: got=%s expected=%s", patch.DiscPath, got, patch.OriginalSHA256)
	}

	xor := asset.ReadFile(patch.XORPath)
	if uint64(len(xor)) != patch.ReplacementSize {
		log.Fatalf("PSS xor length mismatch for %s: got=0x%X expected=0x%X", patch.XORPath, len(xor), patch.ReplacementSize)
	}
	if got := asset.SHA256Hex(xor); got != patch.XORSHA256 {
		log.Fatalf("PSS xor checksum mismatch for %s: got=%s expected=%s", patch.XORPath, got, patch.XORSHA256)
	}

	replacement := asset.ApplyXOR(original, xor)
	if got := asset.SHA256Hex(replacement); got != patch.ReplacementSHA256 {
		log.Fatalf("PSS recovered replacement checksum mismatch for %s: got=%s expected=%s", patch.DiscPath, got, patch.ReplacementSHA256)
	}

	padded := asset.ZeroPad(replacement, len(original))
	p.WriteFileBytes(patch.DiscPath, 0, padded)
}

// RecoverPSS reconstructs an edited PSS from the disc original and the XOR
// delta recorded in the manifest, writing the result to outPath.
func (p *Patcher) RecoverPSS(manifestPath, discPath, outPath string) {
	manifest := asset.LoadManifest(manifestPath)
	patch, ok := asset.FindPSS(manifest, discPath)
	if !ok {
		log.Fatalf("PSS patch not found for %s in %s", discPath, manifestPath)
	}

	original := p.ReadWholeFile(patch.DiscPath)
	if got := asset.SHA256Hex(original); got != patch.OriginalSHA256 {
		log.Fatalf("PSS original checksum mismatch for %s: got=%s expected=%s", patch.DiscPath, got, patch.OriginalSHA256)
	}

	xor := asset.ReadFile(patch.XORPath)
	replacement := asset.ApplyXOR(original, xor)
	if got := asset.SHA256Hex(replacement); got != patch.ReplacementSHA256 {
		log.Fatalf("PSS recovered checksum mismatch for %s: got=%s expected=%s", patch.DiscPath, got, patch.ReplacementSHA256)
	}

	asset.WriteFile(outPath, replacement)

	sidecar := AssetSidecar{
		Type:           "pss",
		DiscPath:       patch.DiscPath,
		OriginalSize:   patch.OriginalSize,
		OriginalSHA256: patch.OriginalSHA256,
	}
	asset.WriteJSON(asset.SidecarPath(outPath), sidecar)
}

// ApplyAssetManifest reads the manifest at manifestPath and applies all PSS
// and PSI patches it contains. Missing manifest is not an error.
func (p *Patcher) ApplyAssetManifest(manifestPath string) {
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return
	}

	manifest := asset.LoadManifest(manifestPath)

	if len(manifest.PSI) > 0 {
		p.ApplyPSIPatches(manifest.PSI)
	}

	for _, patch := range manifest.PSS {
		p.ApplyPSSPatch(patch)
	}
}
