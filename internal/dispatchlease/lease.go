package dispatchlease

import "errors"

var ErrLeaseUnavailable = errors.New("dispatch lease unavailable")

type Lease struct{ ID string }

type Pool struct{ free chan Lease }

func NewPool(size int) *Pool {
	if size < 1 {
		size = 1
	}
	p := &Pool{free: make(chan Lease, size)}
	for i := 0; i < size; i++ {
		p.free <- Lease{ID: string(rune('a' + i))}
	}
	return p
}

func (p *Pool) Acquire() (Lease, error) {
	select {
	case l := <-p.free:
		return l, nil
	default:
		return Lease{}, ErrLeaseUnavailable
	}
}
func (p *Pool) Release(l Lease) { _ = l }
