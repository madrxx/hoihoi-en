// Package psi parses PSI (3ISP) texture files  -- used for UI sprites, menu
// icons, and texture atlases.
//
// A PSI file is a tagged chunk format (similar to RIFF but with 16-byte
// alignment). Each file contains at least two GAMI chunks: a palette chunk
// and an indexed pixel-data chunk. The palette uses 256-entry 32-bit RGBA
// with a PS2 CLUT reorder pass.
//
// Typical usage:
//
//	data := readFromUFP(...)
//	img := psi.ParseImage(data)
//	indices := psi.ExtractIndices(data, img)
//	psi.WriteIndexedPNG("output.png", indices, img)
package psi

import (
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/madrxx/hoihoi-en/ufp"
)

// ----
// Chunk parsing
// ----

// Chunk is a tagged data section within a PSI file.
type Chunk struct {
	Tag           string
	ChunkOffset   uint64
	PayloadOffset uint64
	Payload       []byte
}

// IsPSI reports whether data starts with the "3ISP" magic.
func IsPSI(data []byte) bool {
	return len(data) >= 4 && string(data[:4]) == "3ISP"
}

// ParseChunks splits a PSI file into its tagged chunks. Chunks are 16-byte
// aligned; parsing stops at the KCOE end marker.
func ParseChunks(data []byte) []Chunk {
	if len(data) < 8 || string(data[:4]) != "3ISP" {
		log.Fatalf("not a 3ISP/PSI file")
	}

	var chunks []Chunk
	pos := 0

	for pos+8 <= len(data) {
		tag := string(data[pos : pos+4])
		size := int(ufp.U32(data, pos+4))
		payloadStart := pos + 8
		payloadEnd := payloadStart + size

		if size < 0 || payloadEnd > len(data) {
			break
		}

		chunks = append(chunks, Chunk{
			Tag:           tag,
			ChunkOffset:   uint64(pos),
			PayloadOffset: uint64(payloadStart),
			Payload:       data[payloadStart:payloadEnd],
		})

		pos = payloadEnd
		if pos%16 != 0 {
			pos += 16 - (pos % 16)
		}

		if tag == "KCOE" {
			break
		}
	}

	return chunks
}

// FindChunks returns all chunks matching the given tag.
func FindChunks(chunks []Chunk, tag string) []Chunk {
	var out []Chunk
	for _, c := range chunks {
		if c.Tag == tag {
			out = append(out, c)
		}
	}
	return out
}

// ----
// Constants
// ----

const (
	// PaletteHeaderSize is the 8-byte header before palette data in a GAMI chunk.
	PaletteHeaderSize = 8
	// ImageHeaderSize is the 8-byte header before pixel data in a GAMI chunk.
	ImageHeaderSize = 8
	// PaletteSize is the fixed 256-entry × 4-byte RGBA palette.
	PaletteSize = 1024
	// UseCLUTReorder enables the PS2 8-bit CLUT swizzle pass when reading palettes.
	UseCLUTReorder = true
)

// ----
// Image metadata
// ----

// Image describes the dimensions and layout of a PSI texture.
type Image struct {
	Width           int
	Height          int
	SourceStride    int
	ImageDataOffset uint64
	Palette         []byte
}

// ParseImage extracts image metadata (dimensions, palette, data offset) from
// a PSI file. The first GAMI chunk is expected to be the palette, the second
// the pixel data.
func ParseImage(data []byte) Image {
	chunks := ParseChunks(data)
	gamis := gamiChunks(chunks)

	palChunk := gamis[0]
	imgChunk := gamis[1]

	width, height := imageDimensions(imgChunk.Payload)

	imageDataOffset := imgChunk.PayloadOffset + ImageHeaderSize
	imageDataLength := uint64(width * height)
	imageDataEnd := imageDataOffset + imageDataLength

	if imageDataEnd > uint64(len(data)) {
		log.Fatalf(
			"PSI image data out of range: %dx%d dataOff=0x%X dataEnd=0x%X psiSize=0x%X",
			width,
			height,
			imageDataOffset,
			imageDataEnd,
			len(data),
		)
	}

	return Image{
		Width:           width,
		Height:          height,
		SourceStride:    width,
		ImageDataOffset: imageDataOffset,
		Palette:         readPalette(data, palChunk),
	}
}

// ----
// Pixel index extraction / replacement
// ----

// ExtractIndices reads the raw 8-bit pixel indices from a PSI file,
// de-swizzling row-by-row.
func ExtractIndices(data []byte, image Image) []byte {
	size := image.Width * image.Height
	out := make([]byte, size)

	for y := 0; y < image.Height; y++ {
		src := int(image.ImageDataOffset) + y*image.SourceStride
		dst := y * image.Width

		if src+image.Width > len(data) {
			log.Fatalf(
				"PSI row read out of range: y=%d src=0x%X width=0x%X psiSize=0x%X",
				y,
				src,
				image.Width,
				len(data),
			)
		}

		copy(out[dst:dst+image.Width], data[src:src+image.Width])
	}

	return out
}

// ReplaceIndices overwrites the pixel indices in a PSI file with new data.
// The indices slice must be exactly width × height bytes.
func ReplaceIndices(data []byte, image Image, indices []byte) {
	expected := image.Width * image.Height
	if len(indices) != expected {
		log.Fatalf(
			"PSI indices length mismatch: got=0x%X expected=0x%X",
			len(indices),
			expected,
		)
	}

	for y := 0; y < image.Height; y++ {
		dst := int(image.ImageDataOffset) + y*image.SourceStride
		src := y * image.Width

		if dst+image.Width > len(data) {
			log.Fatalf(
				"PSI row write out of range: y=%d dst=0x%X width=0x%X psiSize=0x%X",
				y,
				dst,
				image.Width,
				len(data),
			)
		}

		copy(data[dst:dst+image.Width], indices[src:src+image.Width])
	}
}

// ----
// PNG import / export
// ----

// WriteIndexedPNG saves 8-bit pixel indices as an indexed PNG using the
// palette from img.
func WriteIndexedPNG(path string, indices []byte, img Image) {
	if len(indices) != img.Width*img.Height {
		log.Fatalf("indexed PNG data length mismatch: got=0x%X expected=0x%X", len(indices), img.Width*img.Height)
	}

	paletted := image.NewPaletted(
		image.Rect(0, 0, img.Width, img.Height),
		normalisePalette(img.Palette),
	)

	for y := 0; y < img.Height; y++ {
		copy(
			paletted.Pix[y*paletted.Stride:y*paletted.Stride+img.Width],
			indices[y*img.Width:y*img.Width+img.Width],
		)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := png.Encode(f, paletted); err != nil {
		log.Fatal(err)
	}
}

// ReadIndexedPNGIndices reads an indexed PNG and returns the raw 8-bit pixel
// indices, verifying the dimensions match.
func ReadIndexedPNGIndices(path string, width, height int) []byte {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	paletted, ok := img.(*image.Paletted)
	if !ok {
		log.Fatalf("expected indexed/paletted PNG: %s", path)
	}

	bounds := paletted.Bounds()
	if bounds.Dx() != width || bounds.Dy() != height {
		log.Fatalf("PNG dimensions mismatch for %s: got=%dx%d expected=%dx%d", path, bounds.Dx(), bounds.Dy(), width, height)
	}

	out := make([]byte, width*height)
	for y := 0; y < height; y++ {
		src := (bounds.Min.Y+y)*paletted.Stride + bounds.Min.X
		dst := y * width
		copy(out[dst:dst+width], paletted.Pix[src:src+width])
	}

	return out
}

// ----
// Internal helpers
// ----

// gamiChunks returns the GAMI chunks from a PSI file, fataling if there
// are fewer than two (palette + image).
func gamiChunks(chunks []Chunk) []Chunk {
	gamis := FindChunks(chunks, "GAMI")
	if len(gamis) < 2 {
		log.Fatalf("PSI does not contain at least two GAMI chunks")
	}
	return gamis
}

// imageDimensions extracts width and height from a GAMI image payload header.
// Expects PSM value 0x1300 (PSMT8  -- 8-bit indexed).
func imageDimensions(payload []byte) (int, int) {
	if len(payload) < ImageHeaderSize {
		log.Fatalf("PSI image GAMI too small for header: got=0x%X", len(payload))
	}

	psm := binary.LittleEndian.Uint16(payload[0:2])
	width := int(binary.LittleEndian.Uint16(payload[4:6]))
	height := int(binary.LittleEndian.Uint16(payload[6:8]))

	if psm != 0x1300 {
		log.Fatalf("unexpected PSI image PSM/header value: got=0x%X expected=0x1300", psm)
	}
	if width <= 0 || height <= 0 || width > 4096 || height > 4096 {
		log.Fatalf("bad PSI image dimensions: %dx%d", width, height)
	}

	return width, height
}

// readPalette extracts and optionally reorders the 256-entry RGBA palette
// from a GAMI palette chunk.
func readPalette(data []byte, palChunk Chunk) []byte {
	paletteDataOffset := palChunk.PayloadOffset + PaletteHeaderSize
	paletteDataEnd := paletteDataOffset + PaletteSize

	if paletteDataEnd > uint64(len(data)) {
		log.Fatalf(
			"PSI palette data out of range: dataOff=0x%X dataEnd=0x%X psiSize=0x%X",
			paletteDataOffset,
			paletteDataEnd,
			len(data),
		)
	}

	raw := make([]byte, PaletteSize)
	copy(raw, data[paletteDataOffset:paletteDataEnd])

	if UseCLUTReorder {
		raw = reorderCLUT(raw)
	}

	return raw
}

// normalisePalette converts the raw 32-bit RGBA palette bytes to a Go
// color.Palette suitable for image.NewPaletted.
func normalisePalette(raw []byte) color.Palette {
	if len(raw) != PaletteSize {
		log.Fatalf("palette size mismatch: got=0x%X expected=0x%X", len(raw), PaletteSize)
	}

	pal := make(color.Palette, 256)

	for i := 0; i < 256; i++ {
		off := i * 4

		r := raw[off+0]
		g := raw[off+1]
		b := raw[off+2]
		aRaw := raw[off+3]

		a := int(aRaw)
		if a <= 0x80 {
			a *= 2
			if a > 255 {
				a = 255
			}
		}

		pal[i] = color.RGBA{
			R: r,
			G: g,
			B: b,
			A: uint8(a),
		}
	}

	return pal
}

// reorderCLUT applies the PS2 8-bit CLUT swizzle: within each 32-entry
// block, entries 8-15 are swapped with 16-23.
func reorderCLUT(raw []byte) []byte {
	if len(raw) < PaletteSize {
		log.Fatalf("palette too small for PS2 CLUT reorder: got=0x%X", len(raw))
	}

	out := make([]byte, PaletteSize)
	copy(out, raw[:PaletteSize])

	for base := 0; base < 256; base += 32 {
		for i := 0; i < 8; i++ {
			dstA := (base + 8 + i) * 4
			srcA := (base + 16 + i) * 4
			dstB := (base + 16 + i) * 4
			srcB := (base + 8 + i) * 4

			copy(out[dstA:dstA+4], raw[srcA:srcA+4])
			copy(out[dstB:dstB+4], raw[srcB:srcB+4])
		}
	}

	return out
}
