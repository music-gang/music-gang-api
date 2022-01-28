package entity

import "time"

type RevisionNumber uint

type Revision struct {
	ID           int64
	Rev          RevisionNumber
	CreatedAt    time.Time
	RawCode      string
	CompiledCode []byte
}
