package schemabarrier

import (
	"errors"
	"fmt"
)

var (
	ErrStaleSchema       = errors.New("stale schema version")
	ErrMigrationPending  = errors.New("schema migration pending")
	ErrMigrationConflict = errors.New("schema migration conflict")
)

type Version struct {
	Epoch uint64
	Value uint64
}

func (v Version) Compare(other Version) int {
	if v.Epoch < other.Epoch {
		return -1
	}
	if v.Epoch > other.Epoch {
		return 1
	}
	if v.Value < other.Value {
		return -1
	}
	if v.Value > other.Value {
		return 1
	}
	return 0
}

func staleError(have, need Version) error {
	return fmt.Errorf("%w: have %d/%d need %d/%d", ErrStaleSchema, have.Epoch, have.Value, need.Epoch, need.Value)
}
