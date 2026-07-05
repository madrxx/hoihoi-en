// Package mips provides MIPS IV (R5900) instruction encoding helpers and
// a code buffer abstraction for generating runtime patches.
package mips

import "encoding/binary"

// ----
// Instruction encoders
// ----

// Instruction encoders for the MIPS IV (R5900) instruction set.
//
// Each function takes register numbers and immediate values and returns the
// encoded 32-bit instruction word. These are pure functions with no side
// effects  -- they don't allocate, don't mutate state, and produce the same
// output for the same inputs every time.

// J encodes an unconditional jump to a 26-bit absolute address.
func J(addr uint32) uint32 { return 0x08000000 | ((addr >> 2) & 0x03FFFFFF) }

// Jal encodes a jump-and-link: jumps to addr and stores the return address in $31.
func Jal(addr uint32) uint32 { return 0x0C000000 | ((addr >> 2) & 0x03FFFFFF) }

// Lui loads a 16-bit immediate into the upper half of register rt.
func Lui(rt int, imm uint32) uint32 { return 0x3C000000 | (uint32(rt) << 16) | (imm & 0xFFFF) }

// Ori bitwise-ORs register rs with a 16-bit zero-extended immediate, storing the result in rt.
func Ori(rt, rs int, imm uint32) uint32 {
	return 0x34000000 | (uint32(rs) << 21) | (uint32(rt) << 16) | (imm & 0xFFFF)
}

// Addiu adds a 16-bit sign-extended immediate to register rs, storing the result in rt.
func Addiu(rt, rs int, imm int32) uint32 {
	return 0x24000000 | (uint32(rs) << 21) | (uint32(rt) << 16) | (uint32(imm) & 0xFFFF)
}

// Lw loads a 32-bit word from memory at [rs + off] into register rt.
func Lw(rt, off, rs int) uint32 {
	return 0x8C000000 | (uint32(rs) << 21) | (uint32(rt) << 16) | (uint32(off) & 0xFFFF)
}

// Sw stores a 32-bit word from register rt to memory at [rs + off].
func Sw(rt, off, rs int) uint32 {
	return 0xAC000000 | (uint32(rs) << 21) | (uint32(rt) << 16) | (uint32(off) & 0xFFFF)
}

// Sb stores the low byte of register rt to memory at [rs + off].
func Sb(rt, off, rs int) uint32 {
	return 0xA0000000 | (uint32(rs) << 21) | (uint32(rt) << 16) | (uint32(off) & 0xFFFF)
}

// Sh stores the low halfword of register rt to memory at [rs + off].
func Sh(rt, off, rs int) uint32 {
	return 0xA4000000 | (uint32(rs) << 21) | (uint32(rt) << 16) | (uint32(off) & 0xFFFF)
}

// Lhu loads a 16-bit halfword from memory at [rs + off] into rt, zero-extending.
func Lhu(rt, off, rs int) uint32 {
	return 0x94000000 | (uint32(rs) << 21) | (uint32(rt) << 16) | (uint32(off) & 0xFFFF)
}

// Andi bitwise-ANDs register rs with a 16-bit zero-extended immediate, storing the result in rt.
func Andi(rt, rs int, imm uint32) uint32 {
	return 0x30000000 | (uint32(rs) << 21) | (uint32(rt) << 16) | (imm & 0xFFFF)
}

// Sll shifts register rt left by sh bits, storing the result in rd.
func Sll(rd, rt, sh int) uint32 { return (uint32(rt) << 16) | (uint32(rd) << 11) | (uint32(sh) << 6) }

// Srl shifts register rt right by sh bits (logical), storing the result in rd.
func Srl(rd, rt, sh int) uint32 {
	return (uint32(rt) << 16) | (uint32(rd) << 11) | (uint32(sh) << 6) | 0x02
}

// Addu adds registers rs and rt (unsigned, no overflow trap), storing the result in rd.
func Addu(rd, rs, rt int) uint32 {
	return (uint32(rs) << 21) | (uint32(rt) << 16) | (uint32(rd) << 11) | 0x21
}

// Daddu adds registers rs and rt (64-bit unsigned, no overflow trap), storing the result in rd.
func Daddu(rd, rs, rt int) uint32 {
	return (uint32(rs) << 21) | (uint32(rt) << 16) | (uint32(rd) << 11) | 0x2D
}

// Beq branches to PC+4+off*4 if registers rs and rt are equal.
func Beq(rs, rt int, off int) uint32 {
	return 0x10000000 | (uint32(rs) << 21) | (uint32(rt) << 16) | (uint32(off) & 0xFFFF)
}

// Bne branches to PC+4+off*4 if registers rs and rt are not equal.
func Bne(rs, rt int, off int) uint32 {
	return 0x14000000 | (uint32(rs) << 21) | (uint32(rt) << 16) | (uint32(off) & 0xFFFF)
}

// ----
// Helpers
// ----

// WriteU32sLE encodes a slice of uint32 values as little-endian bytes.
func WriteU32sLE(values []uint32) []byte {
	out := make([]byte, len(values)*4)
	for i, value := range values {
		binary.LittleEndian.PutUint32(out[i*4:i*4+4], value)
	}
	return out
}

// RuntimeAddr converts a file offset to a runtime (KSEG0) address.
// The game loads the main executable at 0x000FFF80.
func RuntimeAddr(fileOffset uint64) uint32 { return uint32(fileOffset + 0x000FFF80) }

// BranchOffset returns the signed offset (in instructions) from branchIndex
// to targetIndex, for use in MIPS branch instruction immediates.
func BranchOffset(branchIndex int, targetIndex int) int { return targetIndex - (branchIndex + 1) }
