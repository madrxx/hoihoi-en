// CLI dispatch for the hoihoi-en translation patcher. Provides subcommands
// for patching, file listing, offset mapping, BSV text dumping, PSS/PSI
// extraction and XOR delta creation, and UFP archive inspection.
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/madrxx/hoihoi-en/disc"
	"github.com/madrxx/hoihoi-en/patcher"
	"github.com/madrxx/hoihoi-en/ufp"
)

// cmd describes one CLI subcommand.
type cmd struct {
	name    string   // primary command name
	aliases []string // alternative names
	desc    string   // one-line description for usage output
	args    string   // human-readable argument list
	minArgs int      // minimum positional arguments after the command name
	run     func([]string)
}

var commands = []cmd{
	{"patch", nil,
		"Apply all patches to a disc image",
		"<input-bin> <output-bin>", 2, runPatch},

	{"list-files", nil,
		"List ISO 9660 directory entries",
		"<bin-path>", 1, runListFiles},

	{"dump-missions", nil,
		"Print mission text from BSV tables",
		"<bin-path>", 1, runDumpMissions},

	{"dump-items", nil,
		"Print item text from BSV tables",
		"<bin-path>", 1, runDumpItems},

	{"bin-offset-to-file-offset", []string{"b2f"},
		"Map a raw .bin offset to file + file offset",
		"<bin-path> <bin-offset>", 2, runBinOffsetToFileOffset},

	{"file-offset-to-bin-offset", []string{"f2b"},
		"Map a file offset to a raw .bin offset",
		"<bin-path> <file-path> <file-offset>", 3, runFileOffsetToBinOffset},

	{"extract-pss", nil,
		"Extract a PSS (FMV) file from the disc image",
		"<bin-path> <disc-path> <out-pss>", 3, runExtractPSS},

	{"make-pss-xor", nil,
		"Create XOR delta from an edited PSS file",
		"<bin-path> <edited-pss> <manifest-json>", 3, runMakePSSXOR},

	{"recover-pss", nil,
		"Recover an edited PSS from disc + XOR delta",
		"<bin-path> <manifest-json> <disc-path> <out-pss>", 4, runRecoverPSS},

	{"list-ufp", nil,
		"List entries in a UFP archive",
		"<bin-path> <ufp-disc-path>", 2, runListUFP},

	{"extract-psi-all", nil,
		"Extract all PSI textures from a UFP to PNGs",
		"<bin-path> <ufp-disc-path> <out-dir>", 3, runExtractPSIAll},

	{"extract-psi", nil,
		"Extract a single PSI texture to PNG",
		"<bin-path> <ufp-disc-path> <ufp-entry-path> <out-png>", 4, runExtractPSI},

	{"make-psi-xor", nil,
		"Create XOR delta from an edited PSI texture",
		"<bin-path> <edited-png> <manifest-json>", 3, runMakePSIXOR},

	{"recover-psi", nil,
		"Recover an edited PSI from disc + XOR delta",
		"<bin-path> <manifest-json> <ufp-entry-path> <out-png>", 4, runRecoverPSI},
}

func printUsage() {
	fmt.Println("usage:")
	for _, c := range commands {
		names := c.name
		for _, a := range c.aliases {
			names += ", " + a
		}
		fmt.Printf("  hoihoi-en %-30s %s\n", names, c.args)
	}
	fmt.Println()
	fmt.Println("examples:")
	fmt.Println("  hoihoi-en patch clean/hoihoi.bin patched/hoihoi.bin")
	fmt.Println("  hoihoi-en b2f clean/hoihoi.bin 0x0645E5EF")
	fmt.Println("  hoihoi-en f2b clean/hoihoi.bin UFP/GAME.UFP 0x1D25F7")
}

func findCommand(name string) (cmd, bool) {
	for _, c := range commands {
		if name == c.name {
			return c, true
		}
		for _, a := range c.aliases {
			if name == a {
				return c, true
			}
		}
	}
	return cmd{}, false
}

func parseOffset(input string) uint64 {
	value, err := strconv.ParseUint(input, 0, 64)
	if err != nil {
		log.Fatalf("invalid offset %q: %v", input, err)
	}
	return value
}

func loadImage(path string) []byte {
	image, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return image
}

// loadPatcher reads a disc image and returns an initialised Patcher.
// Used by most subcommands that take <bin-path> as their first argument.
func loadPatcher(binPath string) *patcher.Patcher {
	return patcher.New(loadImage(binPath))
}

func runBinOffsetToFileOffset(args []string) {
	binPath := args[0]
	binOffset := parseOffset(args[1])

	image := loadImage(binPath)

	info := disc.BinToUser(binOffset)

	fmt.Printf("bin offset:           0x%X\n", info.BinOffset)
	fmt.Printf("sector:               0x%X\n", info.Sector)
	fmt.Printf("offset within sector: 0x%X\n", info.OffsetWithinSector)

	if !info.IsUserData {
		fmt.Println("area:                 not user data")
		return
	}

	fmt.Println("area:                 user data")
	fmt.Printf("user offset:          0x%X\n", info.UserOffset)

	files, err := disc.ListFiles(image)
	if err != nil {
		log.Fatal(err)
	}

	file, fileOffset, ok := disc.FindFileAtOffset(files, info.UserOffset)
	if !ok {
		fmt.Println("file:                 not found")
		return
	}

	fmt.Printf("file:                 %s\n", file.Path)
	fmt.Printf("file offset:          0x%X\n", fileOffset)
}

func runFileOffsetToBinOffset(args []string) {
	binPath := args[0]
	filePath := args[1]
	fileOffset := parseOffset(args[2])

	image := loadImage(binPath)

	files, err := disc.ListFiles(image)
	if err != nil {
		log.Fatal(err)
	}

	file, ok := disc.FindFile(files, filePath)
	if !ok {
		log.Fatalf("file not found: %s", filePath)
	}

	if file.IsDirectory {
		log.Fatalf("path is a directory: %s", filePath)
	}

	if fileOffset >= file.Size {
		log.Fatalf(
			"file offset 0x%X is outside %s size 0x%X",
			fileOffset,
			file.Path,
			file.Size,
		)
	}

	binOffset := disc.FileToBin(file.StartLBA, fileOffset)

	fmt.Printf("file:        %s\n", file.Path)
	fmt.Printf("file LBA:    0x%X\n", file.StartLBA)
	fmt.Printf("file size:   0x%X\n", file.Size)
	fmt.Printf("file offset: 0x%X\n", fileOffset)
	fmt.Printf("bin offset:  0x%X\n", binOffset)
}

func runListFiles(args []string) {
	image := loadImage(args[0])

	files, err := disc.ListFiles(image)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		kind := "file"
		if file.IsDirectory {
			kind = "dir "
		}

		fmt.Printf("%s LBA=0x%X size=0x%X %q\n", kind, file.StartLBA, file.Size, file.Path)
	}
}

func runPatch(args []string) {
	p := loadPatcher(args[0])
	patcher.ApplyAll(p)

	err := os.WriteFile(args[1], p.Image, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func runDumpMissions(args []string) {
	p := loadPatcher(args[0])
	p.DumpMissionTexts()
}

func runDumpItems(args []string) {
	p := loadPatcher(args[0])
	p.DumpItemTexts()
}

func runExtractPSS(args []string) {
	p := loadPatcher(args[0])
	p.ExtractPSS(args[1], args[2])
}

func runMakePSSXOR(args []string) {
	p := loadPatcher(args[0])
	p.MakePSSXOR(args[1], args[2])
}

func runRecoverPSS(args []string) {
	p := loadPatcher(args[0])
	p.RecoverPSS(args[1], args[2], args[3])
}

func runExtractPSI(args []string) {
	p := loadPatcher(args[0])
	p.ExtractPSI(args[1], args[2], args[3])
}

func runMakePSIXOR(args []string) {
	p := loadPatcher(args[0])
	p.MakePSIXOR(args[1], args[2])
}

func runRecoverPSI(args []string) {
	p := loadPatcher(args[0])
	p.RecoverPSI(args[1], args[2], args[3])
}

func runListUFP(args []string) {
	p := loadPatcher(args[0])

	ufpData := p.ReadWholeFile(args[1])
	entries, err := ufp.Parse(ufpData)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range entries {
		fmt.Printf(
			"%3d  offset=0x%08X  size=0x%06X  %s\n",
			entry.Index,
			entry.Offset,
			entry.Size,
			entry.Path,
		)
	}
}

func runExtractPSIAll(args []string) {
	p := loadPatcher(args[0])
	p.ExtractAllPSI(args[1], args[2])
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	commandName := os.Args[1]
	args := os.Args[2:]

	c, ok := findCommand(commandName)
	if !ok {
		log.Fatalf("unknown command: %s", commandName)
	}

	if len(args) < c.minArgs {
		log.Fatalf("%s expects: %s", c.name, c.args)
	}

	c.run(args)
}
