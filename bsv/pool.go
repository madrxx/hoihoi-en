package bsv

import "log"

// PoolWriter handles writing strings into a BSV SRTS string pool.
// It wraps a ReaderWriter for disc I/O and a StringPool for deduplication
// and cursor tracking.
type PoolWriter struct {
	RW   ReaderWriter
	Pool *StringPool
}

// NewPoolWriter creates a PoolWriter for the given table using the full
// SRTS area.
func NewPoolWriter(rw ReaderWriter, t Table) PoolWriter {
	pool := NewStringPool(t)
	return PoolWriter{RW: rw, Pool: &pool}
}

// NewPoolWriterRange creates a PoolWriter restricted to [start, end) within
// the SRTS area.
func NewPoolWriterRange(rw ReaderWriter, t Table, start, end uint64) PoolWriter {
	pool := NewStringPoolRange(t, start, end)
	return PoolWriter{RW: rw, Pool: &pool}
}

// PutString writes a null-terminated raw string into the pool and returns
// its relative offset. Returns 0 for empty data. Deduplicates: if the same
// byte sequence was already written, returns the previous offset.
func (pw *PoolWriter) PutString(raw []byte) uint32 {
	if len(raw) == 0 {
		return 0
	}

	key := string(raw)
	if pw.Pool.Seen(key) {
		return pw.Pool.RelFor(key)
	}

	writeSize := uint64(len(raw) + 1)
	end := pw.Pool.Cursor() + writeSize

	if end > pw.Pool.End() {
		overBy := end - pw.Pool.End()
		usedBefore := pw.Pool.Cursor() - pw.Pool.Start()
		capacity := pw.Pool.End() - pw.Pool.Start()

		log.Fatalf(
			"BSV string pool overflow: table=0x%X cursor=0x%X needEnd=0x%X limit=0x%X usedBefore=%d capacity=%d overBy=%d bytes writeSize=%d",
			pw.Pool.Table.Base,
			pw.Pool.Cursor(),
			end,
			pw.Pool.End(),
			usedBefore,
			capacity,
			overBy,
			writeSize,
		)
	}

	rel64 := pw.Pool.Cursor() - pw.Pool.Table.Base
	if rel64 > 0xFFFFFFFF {
		log.Fatalf("BSV relative string offset too large: table=0x%X rel=0x%X", pw.Pool.Table.Base, rel64)
	}

	data := make([]byte, 0, len(raw)+1)
	data = append(data, raw...)
	data = append(data, 0x00)

	pw.RW.WriteFileBytes(pw.Pool.Table.FilePath, pw.Pool.Cursor(), data)

	rel := uint32(rel64)
	pw.Pool.Dedup(key, rel)
	pw.Pool.SetCursor(end)

	return rel
}
