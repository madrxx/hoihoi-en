// Package asset manages the XOR-patch manifest for PSS (FMV) and PSI
// (texture) assets.
//
// The manifest (patches/assets.json) records every replaced asset along
// with SHA-256 checksums of the original data, replacement data, and XOR
// delta. This lets the patcher verify integrity before applying any patch
// and supports round-tripping: extract -> edit -> make-xor -> apply.
//
// Each asset also has a .hoihoi.json sidecar file that stores metadata
// needed to re-derive the XOR without re-parsing the disc image.
package asset

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// DefaultManifestPath is the conventional location for the asset manifest.
const DefaultManifestPath = "patches/assets.json"

// ----
// Manifest types
// ----

// Manifest records all PSS and PSI asset patches.
type Manifest struct {
	Version int        `json:"version"`
	PSS     []PSSPatch `json:"pss,omitempty"`
	PSI     []PSIPatch `json:"psi,omitempty"`
}

// PSSPatch describes a single PSS (FMV) replacement via XOR.
type PSSPatch struct {
	DiscPath          string `json:"discPath"`
	XORPath           string `json:"xorPath"`
	OriginalSize      uint64 `json:"originalSize"`
	ReplacementSize   uint64 `json:"replacementSize"`
	OriginalSHA256    string `json:"originalSha256"`
	ReplacementSHA256 string `json:"replacementSha256"`
	XORSHA256         string `json:"xorSha256"`
}

// PSIPatch describes a single PSI (texture) replacement via XOR.
type PSIPatch struct {
	UFPDiscPath        string `json:"ufpDiscPath"`
	UFPEntryPath       string `json:"ufpEntryPath"`
	XORPath            string `json:"xorPath"`
	Width              int    `json:"width"`
	Height             int    `json:"height"`
	SourceStride       int    `json:"sourceStride"`
	ImageDataOffset    uint64 `json:"imageDataOffset"`
	VisiblePixelLength uint64 `json:"visiblePixelLength"`

	OriginalSHA256    string `json:"originalIndicesSha256"`
	ReplacementSHA256 string `json:"replacementIndicesSha256"`
	XORSHA256         string `json:"xorSha256"`
}

// Sidecar stores per-asset metadata alongside the extracted/edited file.
// The .hoihoi.json sidecar is written next to the extracted asset so that
// make-xor and recover commands can reconstruct patches without re-parsing
// the disc image.
type Sidecar struct {
	Type string `json:"type"`

	DiscPath string `json:"discPath,omitempty"`

	UFPDiscPath        string `json:"ufpDiscPath,omitempty"`
	UFPEntryPath       string `json:"ufpEntryPath,omitempty"`
	Width              int    `json:"width,omitempty"`
	Height             int    `json:"height,omitempty"`
	SourceStride       int    `json:"sourceStride,omitempty"`
	ImageDataOffset    uint64 `json:"imageDataOffset,omitempty"`
	VisiblePixelLength uint64 `json:"visiblePixelLength,omitempty"`

	OriginalSize   uint64 `json:"originalSize,omitempty"`
	OriginalSHA256 string `json:"originalSha256"`
}

// ----
// File I/O helpers
// ----

// ReadFile reads an entire file, fataling on error.
func ReadFile(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

// WriteFile writes data to path, creating parent directories as needed.
func WriteFile(path string, data []byte) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Fatal(err)
	}
}

// ReadJSON unmarshals a JSON file into v, fataling on error.
func ReadJSON(path string, v any) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		log.Fatalf("could not parse JSON %s: %v", path, err)
	}
}

// WriteJSON marshals v as indented JSON to path.
func WriteJSON(path string, v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	data = append(data, '\n')
	WriteFile(path, data)
}

// ----
// Manifest I/O
// ----

// LoadManifest reads an asset manifest from path. If the file does not exist,
// an empty version-1 manifest is returned.
func LoadManifest(path string) Manifest {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Manifest{Version: 1}
	}
	if err != nil {
		log.Fatal(err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		log.Fatalf("could not parse asset manifest %s: %v", path, err)
	}
	if manifest.Version != 1 {
		log.Fatalf("unsupported asset manifest version %d in %s", manifest.Version, path)
	}
	return manifest
}

// SaveManifest writes an asset manifest to path, forcing version 1.
func SaveManifest(path string, manifest Manifest) {
	manifest.Version = 1
	WriteJSON(path, manifest)
}

// ----
// Path helpers
// ----

// SidecarPath returns the conventional .hoihoi.json sidecar path for an
// extracted asset file.
func SidecarPath(assetPath string) string {
	return assetPath + ".hoihoi.json"
}

// SafeName sanitises a string for use as a filename component.
func SafeName(s string) string {
	s = strings.ReplaceAll(s, "\\", "/")
	s = strings.Trim(s, "/")
	replacer := strings.NewReplacer("/", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_")
	s = replacer.Replace(s)
	if s == "" {
		return "asset"
	}
	return s
}

// DefaultXORPathForPSS returns the conventional .xor path for a PSS file.
func DefaultXORPathForPSS(discPath string) string {
	return filepath.ToSlash(filepath.Join("patches", "fmv", SafeName(discPath)+".xor"))
}

// DefaultXORPathForPSI returns the conventional .xor path for a PSI entry.
func DefaultXORPathForPSI(ufpEntryPath string) string {
	return filepath.ToSlash(filepath.Join("patches", "psi", SafeName(ufpEntryPath)+".indices.xor"))
}

// ----
// XOR operations
// ----

// SHA256Hex returns the hex-encoded SHA-256 hash of data.
func SHA256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// MakeXOR produces an XOR delta from original -> replacement. The
// replacement must not be larger than the original.
func MakeXOR(original, replacement []byte) []byte {
	if len(replacement) > len(original) {
		log.Fatalf("replacement is larger than original: replacement=0x%X original=0x%X", len(replacement), len(original))
	}
	out := make([]byte, len(replacement))
	for i := range replacement {
		out[i] = original[i] ^ replacement[i]
	}
	return out
}

// ApplyXOR recovers the replacement by XORing original with the delta.
func ApplyXOR(original, xor []byte) []byte {
	if len(xor) > len(original) {
		log.Fatalf("xor is larger than original: xor=0x%X original=0x%X", len(xor), len(original))
	}
	out := make([]byte, len(xor))
	for i := range xor {
		out[i] = original[i] ^ xor[i]
	}
	return out
}

// ZeroPad returns a slice of length size, with data copied to the front
// and the remainder zero-filled.
func ZeroPad(data []byte, size int) []byte {
	if len(data) > size {
		log.Fatalf("cannot pad: data len=0x%X size=0x%X", len(data), size)
	}
	out := make([]byte, size)
	copy(out, data)
	return out
}

// ----
// Manifest mutation helpers
// ----

// UpsertPSS adds or replaces a PSS patch in the manifest, matched by
// disc path (case-insensitive).
func UpsertPSS(manifest *Manifest, patch PSSPatch) {
	for i := range manifest.PSS {
		if strings.EqualFold(manifest.PSS[i].DiscPath, patch.DiscPath) {
			manifest.PSS[i] = patch
			return
		}
	}
	manifest.PSS = append(manifest.PSS, patch)
}

// UpsertPSI adds or replaces a PSI patch in the manifest, matched by
// UFP disc path and entry path (case-insensitive).
func UpsertPSI(manifest *Manifest, patch PSIPatch) {
	for i := range manifest.PSI {
		if strings.EqualFold(manifest.PSI[i].UFPDiscPath, patch.UFPDiscPath) &&
			strings.EqualFold(manifest.PSI[i].UFPEntryPath, patch.UFPEntryPath) {
			manifest.PSI[i] = patch
			return
		}
	}
	manifest.PSI = append(manifest.PSI, patch)
}

// FindPSS returns the PSS patch for discPath, if present.
func FindPSS(manifest Manifest, discPath string) (PSSPatch, bool) {
	for _, patch := range manifest.PSS {
		if strings.EqualFold(patch.DiscPath, discPath) {
			return patch, true
		}
	}
	return PSSPatch{}, false
}

// FindPSI returns the PSI patch for ufpEntryPath, if present.
func FindPSI(manifest Manifest, ufpEntryPath string) (PSIPatch, bool) {
	for _, patch := range manifest.PSI {
		if strings.EqualFold(patch.UFPEntryPath, ufpEntryPath) {
			return patch, true
		}
	}
	return PSIPatch{}, false
}
