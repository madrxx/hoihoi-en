package bsv

// RebuildPool clears the SRTS string pool and rewrites all EMAN and DNIK
// string references through the given PoolWriter. After this call the pool
// cursor is positioned after the last reference string, ready for new data.
//
// emanRefs and dnikRefs are pairs of (fileOffset, rawBytes) describing the
// existing string references that must be preserved. Each reference's u32
// pointer is updated to the new pool position.
func RebuildPool(pw *PoolWriter, emanRefs, dnikRefs []FieldRef) {
	t := pw.Pool.Table

	pw.RW.WriteFileBytes(
		t.FilePath,
		t.SRTSPoolStart(),
		make([]byte, t.SRTSEnd()-t.SRTSPoolStart()),
	)

	for _, ref := range emanRefs {
		rel := pw.PutString(ref.Raw)
		pw.RW.WriteFileU32LE(t.FilePath, ref.FileOffset, rel)
	}
	for _, ref := range dnikRefs {
		rel := pw.PutString(ref.Raw)
		pw.RW.WriteFileU32LE(t.FilePath, ref.FileOffset, rel)
	}
}
