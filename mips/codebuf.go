package mips

// CodeBuf builds a []uint32 MIPS instruction sequence and provides deferred
// branch resolution. Instead of tracking indices manually and backpatching
// by slice position, callers use AddPlaceholder() to record insertion
// points and resolve branches once the target position is known.
//
// Usage:
//
//	buf := NewCodeBuf()
//	buf.Add(Addiu(29, 29, -0x20))
//	skipBranch := buf.AddPlaceholder()    // placeholder beq
//	buf.Add(Lw(2, 0, 4))
//	// ... more instructions ...
//	target := buf.Here()
//	buf.ResolveBeq(skipBranch, 2, 0, target) // overwrites placeholder
type CodeBuf struct {
	words []uint32
}

// NewCodeBuf returns an empty code buffer.
func NewCodeBuf() *CodeBuf {
	return &CodeBuf{}
}

// Add appends instructions to the buffer and returns the index of the first
// instruction added (the position before the append).
func (b *CodeBuf) Add(insns ...uint32) int {
	pos := len(b.words)
	b.words = append(b.words, insns...)
	return pos
}

// AddPlaceholder appends a zero word and returns its index for later
// resolution.
func (b *CodeBuf) AddPlaceholder() int {
	return b.Add(0)
}

// Here returns the current length (the index where the next Add would place
// its first instruction). Use this as the target for branch resolution.
func (b *CodeBuf) Here() int {
	return len(b.words)
}

// ResolveBeq overwrites the word at index pos with a Beq instruction that
// branches to target. rs and rt are the comparison registers.
func (b *CodeBuf) ResolveBeq(pos int, rs, rt int, target int) {
	b.words[pos] = Beq(rs, rt, BranchOffset(pos, target))
}

// ResolveBne overwrites the word at index pos with a Bne instruction that
// branches to target.
func (b *CodeBuf) ResolveBne(pos int, rs, rt int, target int) {
	b.words[pos] = Bne(rs, rt, BranchOffset(pos, target))
}

// ResolveJ overwrites the word at index pos with a J instruction for the
// given absolute address.
func (b *CodeBuf) ResolveJ(pos int, addr uint32) {
	b.words[pos] = J(addr)
}

// ResolveJal overwrites the word at index pos with a Jal instruction.
func (b *CodeBuf) ResolveJal(pos int, addr uint32) {
	b.words[pos] = Jal(addr)
}

// LoadAddr emits a LUI/ORI pair to load a 32-bit address into a register.
func (b *CodeBuf) LoadAddr(reg int, addr uint32) {
	b.Add(Lui(reg, addr>>16), Ori(reg, reg, addr&0xFFFF))
}

// Words returns the underlying instruction slice.
func (b *CodeBuf) Words() []uint32 {
	return b.words
}
