package domain

import "time"

// Todo represents a task item in the homelab system.
// This type lives in shared/ and contains only pure domain data —
// no I/O, no external dependencies beyond stdlib.
type Todo struct {
	ID        int64
	Title     string
	Done      bool
	CreatedAt time.Time
}
