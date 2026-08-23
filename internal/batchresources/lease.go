package batchresources

import "errors"

type Lease interface{ Release() error }

func WithLease(lease Lease, operation func() error) error {
	operationErr := operation()
	releaseErr := lease.Release()
	return errors.Join(operationErr, releaseErr)
}
