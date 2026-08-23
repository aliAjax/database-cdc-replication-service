package slotlease

import (
	"errors"
	"time"
)

var (
	ErrLeaseHeld    = errors.New("replication slot lease is held")
	ErrLeaseExpired = errors.New("replication slot lease expired")
	ErrLeaseOwner   = errors.New("replication slot lease owner mismatch")
)

type Lease struct {
	Slot       string
	Owner      string
	Generation uint64
	ExpiresAt  time.Time
}

func (l Lease) Active(now time.Time) bool {
	return l.Owner != "" && now.Before(l.ExpiresAt)
}

func (l Lease) OwnedBy(owner string, generation uint64, now time.Time) bool {
	return l.Active(now) && l.Owner == owner && l.Generation == generation
}
