package batchresources

import "errors"

type Transaction interface {
	Commit() error
	Rollback() error
}

func Finish(tx Transaction, operation func() error) error {
	if err := operation(); err != nil {
		return errors.Join(err, tx.Rollback())
	}
	return tx.Commit()
}
