package batchresources

import "errors"

type Lease interface{ Release() error }

func WithLease(lease Lease, operation func() error) error {
	operationErr := operation()
	if operationErr != nil {
		return operationErr
	}
	releaseErr := lease.Release()
	return errors.Join(operationErr, releaseErr)
}
