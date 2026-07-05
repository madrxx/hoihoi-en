// Patcher struct and sector-aware disc I/O layer. The Patcher holds the
// in-memory disc image and parsed ISO 9660 directory, and its methods
// translate between file-relative offsets and raw .bin offsets, crossing
// 2352-byte CD sector boundaries transparently.

package main

import (
	"encoding/binary"
	"log"

	"github.com/madrxx/hoihoi-en/disc"
	"github.com/madrxx/hoihoi-en/encoding"
)

// Patcher holds the in-memory disc image and parsed ISO 9660 directory.
// Its methods translate between file-relative offsets and raw .bin offsets,
// crossing 2352-byte CD sector boundaries transparently.
type Patcher struct {
	Image []byte      // loaded CD image
	Files []disc.File // parsed ISO 9660 file list
}

func newPatcher(image []byte) Patcher {
	files, err := disc.ListFiles(image)
	if err != nil {
		log.Fatal(err)
	}

	return Patcher{
		Image: image,
		Files: files,
	}
}

// WriteRawBytes writes data at a raw .bin offset, checking bounds.
func (p *Patcher) WriteRawBytes(offset uint64, data []byte) {
	end := offset + uint64(len(data))

	if end > uint64(len(p.Image)) {
		log.Fatalf("write out of range: offset=0x%X len=0x%X", offset, len(data))
	}

	copy(p.Image[int(offset):int(end)], data)
}

// WriteFileBytes writes data at a file-relative offset within a disc file,
// splitting writes across CD sector boundaries as needed.
func (p *Patcher) WriteFileBytes(filePath string, fileOffset uint64, data []byte) {
	file, ok := disc.FindFile(p.Files, filePath)
	if !ok {
		log.Fatalf("file not found: %s", filePath)
	}

	if file.IsDirectory {
		log.Fatalf("path is a directory: %s", filePath)
	}

	if fileOffset+uint64(len(data)) > file.Size {
		log.Fatalf(
			"file patch out of range: %s offset=0x%X len=0x%X fileSize=0x%X",
			file.Path,
			fileOffset,
			len(data),
			file.Size,
		)
	}

	remaining := uint64(len(data))
	dataOffset := uint64(0)
	currentFileOffset := fileOffset

	for remaining > 0 {
		offsetWithinUserSector := currentFileOffset % disc.UserSectorSize
		bytesUntilSectorEnd := disc.UserSectorSize - offsetWithinUserSector

		chunkSize := bytesUntilSectorEnd
		if chunkSize > remaining {
			chunkSize = remaining
		}

		rawOffset := disc.FileToBin(file.StartLBA, currentFileOffset)

		p.WriteRawBytes(
			rawOffset,
			data[int(dataOffset):int(dataOffset+chunkSize)],
		)

		currentFileOffset += chunkSize
		dataOffset += chunkSize
		remaining -= chunkSize
	}
}

// ReadFileBytes reads size bytes from a file-relative offset within a disc
// file, reassembling data that spans CD sector boundaries.
func (p *Patcher) ReadFileBytes(filePath string, fileOffset uint64, size uint64) []byte {
	file, ok := disc.FindFile(p.Files, filePath)
	if !ok {
		log.Fatalf("file not found: %s", filePath)
	}

	if file.IsDirectory {
		log.Fatalf("path is a directory: %s", filePath)
	}

	if fileOffset+size > file.Size {
		log.Fatalf(
			"file read out of range: %s offset=0x%X len=0x%X fileSize=0x%X",
			file.Path,
			fileOffset,
			size,
			file.Size,
		)
	}

	output := make([]byte, 0, size)

	remaining := size
	currentFileOffset := fileOffset

	for remaining > 0 {
		offsetWithinUserSector := currentFileOffset % disc.UserSectorSize
		bytesUntilSectorEnd := disc.UserSectorSize - offsetWithinUserSector

		chunkSize := bytesUntilSectorEnd
		if chunkSize > remaining {
			chunkSize = remaining
		}

		rawOffset := disc.FileToBin(file.StartLBA, currentFileOffset)

		output = append(
			output,
			p.Image[int(rawOffset):int(rawOffset+chunkSize)]...,
		)

		currentFileOffset += chunkSize
		remaining -= chunkSize
	}

	return output
}

// WriteFileU32LE writes a 32-bit little-endian value at a file-relative offset.
func (p *Patcher) WriteFileU32LE(filePath string, fileOffset uint64, value uint32) {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, value)

	p.WriteFileBytes(filePath, fileOffset, data)
}

// ReadFileU32LE reads a 32-bit little-endian value from a file-relative offset.
func (p *Patcher) ReadFileU32LE(filePath string, fileOffset uint64) uint32 {
	data := p.ReadFileBytes(filePath, fileOffset, 4)
	return binary.LittleEndian.Uint32(data)
}

// ReadWholeFile returns the entire contents of a disc file.
func (p *Patcher) ReadWholeFile(filePath string) []byte {
	file, ok := disc.FindFile(p.Files, filePath)
	if !ok {
		log.Fatalf("file not found: %s", filePath)
	}
	if file.IsDirectory {
		log.Fatalf("path is a directory: %s", filePath)
	}
	return p.ReadFileBytes(filePath, 0, file.Size)
}

// toSJIS encodes a string to fullwidth Shift-JIS with nullBytes null bytes
// appended. Fatals on error  -- all callers pass ASCII literals that cannot
// fail SJIS encoding.
func toSJIS(input string, nullBytes int) []byte {
	out, err := encoding.ToSJIS(input, nullBytes)
	if err != nil {
		log.Fatal(err)
	}
	return out
}

// fromSJIS decodes a Shift-JIS byte slice to a Go string. Fatals on
// error  -- the game disc data is known-valid Shift-JIS.
func fromSJIS(raw []byte) string {
	out, err := encoding.FromSJIS(raw)
	if err != nil {
		log.Fatal(err)
	}
	return out
}
