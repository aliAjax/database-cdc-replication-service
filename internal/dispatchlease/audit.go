package dispatchlease

import "errors"

var ErrAuditFlush = errors.New("audit flush failed")

type Audit struct{ flushed bool }

func (a *Audit) Flush(err error) error {
	if err != nil {
		return errors.Join(err, ErrAuditFlush)
	}
	a.flushed = true
	return nil
}
