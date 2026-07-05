// Package disc provides MODE2/2352 CD sector layout helpers and ISO 9660
// file-system parsing for the Hoihoi-san CD image.
//
// The CD image (.bin) stores data in 2352-byte raw sectors. Each sector
// contains 2048 bytes of user data starting at offset 24. The ISO 9660
// file system on top of that user-data stream maps file paths to Logical
// Block Addresses (LBAs) and byte sizes.
//
// Two coordinate systems are used throughout:
//
//   - Raw offset: byte position in the .bin image (includes sector headers/ECC).
//   - User offset: byte position in the logical user-data stream.
//
// Conversion between the two is handled by OffsetInfo.
package disc

import (
	"encoding/binary"
	"fmt"
	"path"
	"strings"
)

// ----
// Raw MODE2/2352 sector layout / offset conversion
// ----

const (
	// RawSectorSize is the full MODE2/2352 sector size in bytes.
	RawSectorSize uint64 = 2352
	// UserSectorSize is the user-data portion of a sector (2048 bytes).
	UserSectorSize uint64 = 2048
	// Mode2DataStart is the byte offset within a raw sector where user data begins.
	Mode2DataStart uint64 = 24
	// Mode2DataEnd is one past the last user-data byte within a raw sector.
	Mode2DataEnd uint64 = Mode2DataStart + UserSectorSize
)

// OffsetInfo describes where a raw .bin offset lands inside the CD sector
// structure and, if applicable, the corresponding position in the user-data
// stream.
type OffsetInfo struct {
	// The original offset in the .bin.
	BinOffset uint64
	// The CD sector number containing that offset.
	Sector uint64
	// Byte position within the 2352-byte sector.
	OffsetWithinSector uint64
	// Whether it falls within the 2048-byte user data area.
	IsUserData bool
	// If IsUserData, the position within the whole user-data stream.
	UserOffset uint64
}

// BinToUser maps a raw .bin offset to the corresponding user-data offset by
// accounting for sector headers and ECC data.
func BinToUser(binOffset uint64) OffsetInfo {
	sector := binOffset / RawSectorSize
	offsetWithinSector := binOffset % RawSectorSize

	info := OffsetInfo{
		BinOffset:          binOffset,
		Sector:             sector,
		OffsetWithinSector: offsetWithinSector,
		IsUserData:         offsetWithinSector >= Mode2DataStart && offsetWithinSector < Mode2DataEnd,
	}

	if info.IsUserData {
		info.UserOffset = sector*UserSectorSize + (offsetWithinSector - Mode2DataStart)
	}

	return info
}

// UserToBin converts a user-data offset back to a raw .bin offset.
func UserToBin(userOffset uint64) uint64 {
	sector := userOffset / UserSectorSize
	offsetWithinSector := userOffset % UserSectorSize

	return sector*RawSectorSize + Mode2DataStart + offsetWithinSector
}

// FileToBin converts a file-relative offset (file identified by its start LBA)
// to a raw .bin offset.
func FileToBin(fileStartLBA uint64, fileOffset uint64) uint64 {
	fileStartUserOffset := fileStartLBA * UserSectorSize
	return UserToBin(fileStartUserOffset + fileOffset)
}

// ----
// Low level sector and extent reading
// ----

// readUserSector returns the 2048-byte user-data area from a sector of the
// raw image.
func readUserSector(image []byte, lba uint64) ([]byte, error) {
	rawStart := lba*RawSectorSize + Mode2DataStart
	rawEnd := rawStart + UserSectorSize

	if rawEnd > uint64(len(image)) {
		return nil, fmt.Errorf("sector 0x%X is outside image", lba)
	}

	return image[rawStart:rawEnd], nil
}

// readExtent returns a whole file extent as a continuous byte slice by reading
// consecutive sectors.
func readExtent(image []byte, startLBA uint64, size uint64) ([]byte, error) {
	output := make([]byte, 0, size)

	sectorCount := (size + UserSectorSize - 1) / UserSectorSize

	for i := uint64(0); i < sectorCount; i++ {
		sector, err := readUserSector(image, startLBA+i)
		if err != nil {
			return nil, err
		}

		remaining := size - uint64(len(output))

		if remaining >= UserSectorSize {
			output = append(output, sector...)
		} else {
			output = append(output, sector[:remaining]...)
		}
	}
	return output, nil
}

// ----
// ISO9660 directory model/parsing
// ----

// File represents a single ISO 9660 directory entry (file or subdirectory).
type File struct {
	// Full path from disc root, e.g. "UFP/GAME.UFP".
	Path string
	// First LBA (sector number) of the file.
	StartLBA uint64
	// File size in bytes, or directory extent size.
	Size uint64
	// True if the entry is a subdirectory.
	IsDirectory bool
}

// cleanISOName normalises current-dir / parent-dir labels and strips the
// ISO 9660 version suffix (e.g. ";1").
func cleanISOName(name string) string {
	if name == "\x00" {
		return "."
	}
	if name == "\x01" {
		return ".."
	}

	if semicolon := strings.LastIndex(name, ";"); semicolon != -1 {
		suffix := name[semicolon+1:]
		if suffix != "" {
			allDigits := true
			for _, r := range suffix {
				if r < '0' || r > '9' {
					allDigits = false
					break
				}
			}
			if allDigits {
				return name[:semicolon]
			}
		}
	}

	return name
}

// parseDirectoryRecord decodes one ISO 9660 directory record from data,
// returning the File, number of bytes consumed, or an error.
func parseDirectoryRecord(data []byte) (File, int, error) {
	if len(data) < 34 {
		return File{}, 0, fmt.Errorf("directory record too short")
	}

	recordLength := int(data[0])
	if recordLength == 0 {
		return File{}, 0, nil
	}

	if recordLength > len(data) {
		return File{}, 0, fmt.Errorf("directory record truncated")
	}

	startLBA := binary.LittleEndian.Uint32(data[2:6])
	size := binary.LittleEndian.Uint32(data[10:14])
	flags := data[25]

	nameLength := int(data[32])
	nameStart := 33
	nameEnd := nameStart + nameLength

	if nameEnd > recordLength {
		return File{}, 0, fmt.Errorf("directory record name truncated")
	}

	name := cleanISOName(string(data[nameStart:nameEnd]))

	file := File{
		Path:        name,
		StartLBA:    uint64(startLBA),
		Size:        uint64(size),
		IsDirectory: flags&0x02 != 0,
	}

	return file, recordLength, nil
}

// readRootDirectoryRecord reads the ISO 9660 primary volume descriptor
// (sector 16) and returns the root directory record.
func readRootDirectoryRecord(image []byte) (File, error) {
	sector, err := readUserSector(image, 16)
	if err != nil {
		return File{}, err
	}

	if string(sector[1:6]) != "CD001" {
		return File{}, fmt.Errorf("sector 16 does not contain CD001")
	}

	root, _, err := parseDirectoryRecord(sector[156:])
	if err != nil {
		return File{}, err
	}

	return root, nil
}

// ----
// Directory traversal
// ----

// readDirectory returns the immediate children of a directory.
func readDirectory(image []byte, dir File) ([]File, error) {
	data, err := readExtent(image, dir.StartLBA, dir.Size)
	if err != nil {
		return nil, err
	}

	var entries []File

	offset := 0
	for offset < len(data) {
		recordLength := int(data[offset])

		if recordLength == 0 {
			nextSector := ((offset / int(UserSectorSize)) + 1) * int(UserSectorSize)
			offset = nextSector
			continue
		}

		entry, consumed, err := parseDirectoryRecord(data[offset:])
		if err != nil {
			return nil, err
		}

		offset += consumed

		if entry.Path == "." || entry.Path == ".." {
			continue
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// walkDirectory recursively reads directories, appending all Files to the
// provided slice.
func walkDirectory(image []byte, dir File, prefix string, files *[]File) error {
	entries, err := readDirectory(image, dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		entry.Path = path.Join(prefix, entry.Path)

		*files = append(*files, entry)

		if entry.IsDirectory {
			if err := walkDirectory(image, entry, entry.Path, files); err != nil {
				return err
			}
		}
	}

	return nil
}

// ListFiles returns all files and directories on the disc.
func ListFiles(image []byte) ([]File, error) {
	root, err := readRootDirectoryRecord(image)
	if err != nil {
		return nil, err
	}

	var files []File

	if err := walkDirectory(image, root, "", &files); err != nil {
		return nil, err
	}

	return files, nil
}

// ----
// Lookup helpers
// ----

// FindFile searches for a disc file by path (case-insensitive, forward-slash
// normalised). Returns the File and true if found.
func FindFile(files []File, wantedPath string) (File, bool) {
	wantedPath = strings.ReplaceAll(wantedPath, "\\", "/")
	wantedPath = strings.TrimPrefix(path.Clean(wantedPath), "/")

	for _, file := range files {
		if strings.EqualFold(file.Path, wantedPath) {
			return file, true
		}
	}

	return File{}, false
}

// FindFileAtOffset returns the file that contains the given user-data offset,
// together with the byte offset within that file.
func FindFileAtOffset(files []File, userOffset uint64) (File, uint64, bool) {
	for _, file := range files {
		if file.IsDirectory {
			continue
		}

		fileStart := file.StartLBA * UserSectorSize
		fileEnd := fileStart + file.Size

		if userOffset >= fileStart && userOffset < fileEnd {
			fileOffset := userOffset - fileStart
			return file, fileOffset, true
		}
	}

	return File{}, 0, false
}
