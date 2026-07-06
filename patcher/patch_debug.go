package patcher

import "github.com/madrxx/hoihoi-en/encoding"

// applyDebugMenuPatch replaces debug menu strings (BSV selection, 3D
// viewer, sounds, cutscene names) in SLPM_623.91 at 0x16C831-0x16DB29.
func applyDebugMenuPatch(p *Patcher) {
	p.WriteFileBytes("SLPM_623.91", 0x16C831, encoding.GameText("<<< Debug Menu >>>"))
	p.WriteFileBytes("SLPM_623.91", 0x16C922, encoding.GameText("<<< BSV Selection >>>"))

	p.WriteFileBytes("SLPM_623.91", 0x16C85E, encoding.GameText("Start"))
	p.WriteFileBytes("SLPM_623.91", 0x16C875, encoding.GameText("Start(Unlocked)"))
	p.WriteFileBytes("SLPM_623.91", 0x16C89A, encoding.GameText("3D Viewer"))
	p.WriteFileBytes("SLPM_623.91", 0x16C8AD, encoding.GameText("Sounds"))
	p.WriteFileBytes("SLPM_623.91", 0x16C8BC, encoding.GameText("Cutscenes"))

	p.WriteFileBytes("SLPM_623.91", 0x16D9B9, encoding.GameText("OP.PSS"))
	p.WriteFileBytes("SLPM_623.91", 0x16D9CE, encoding.GameText("02.PSS"))
	p.WriteFileBytes("SLPM_623.91", 0x16D9E9, encoding.GameText("04.PSS"))
	p.WriteFileBytes("SLPM_623.91", 0x16D9FE, encoding.GameText("05.PSS"))
	p.WriteFileBytes("SLPM_623.91", 0x16DA0B, encoding.GameText("06.PSS"))
	p.WriteFileBytes("SLPM_623.91", 0x16DA34, encoding.GameText("07.PSS"))
	p.WriteFileBytes("SLPM_623.91", 0x16DA41, encoding.GameText("08.PSS"))

	p.WriteFileBytes("SLPM_623.91", 0x16DA60, encoding.GameText("0x07"))
	p.WriteFileBytes("SLPM_623.91", 0x16DA6D, encoding.GameText("0x08"))
	p.WriteFileBytes("SLPM_623.91", 0x16DA7A, encoding.GameText("0x09"))
	p.WriteFileBytes("SLPM_623.91", 0x16DA89, encoding.GameText("0x0A"))
	p.WriteFileBytes("SLPM_623.91", 0x16DA96, encoding.GameText("0x0B"))
	p.WriteFileBytes("SLPM_623.91", 0x16DAA3, encoding.GameText("0x0C"))
	p.WriteFileBytes("SLPM_623.91", 0x16DAB0, encoding.GameText("0x0D"))
	p.WriteFileBytes("SLPM_623.91", 0x16DABD, encoding.GameText("0x0E"))
	p.WriteFileBytes("SLPM_623.91", 0x16DACA, encoding.GameText("0x0F"))

	p.WriteFileBytes("SLPM_623.91", 0x16DAD7, encoding.GameText("0x10"))
	p.WriteFileBytes("SLPM_623.91", 0x16DAE4, encoding.GameText("0x11"))
	p.WriteFileBytes("SLPM_623.91", 0x16DAF1, encoding.GameText("0x12"))
	p.WriteFileBytes("SLPM_623.91", 0x16DAFE, encoding.GameText("0x13"))
	p.WriteFileBytes("SLPM_623.91", 0x16DB0B, encoding.GameText("0x14"))
	p.WriteFileBytes("SLPM_623.91", 0x16DB1A, encoding.GameText("0x15"))
	p.WriteFileBytes("SLPM_623.91", 0x16DB29, encoding.GameText("0x16"))
}
