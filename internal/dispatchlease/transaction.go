package dispatchlease

import (
	"errors"
	"fmt"
)

var ErrRollback = errors.New("transaction rolled back")

type Transaction struct{ committed bool }

func (t *Transaction) Commit() error   { t.committed = true; return nil }
func (t *Transaction) Rollback() error { t.committed = false; return ErrRollback }
func FinalizeTransaction(t *Transaction, operation error, commitErr error) error {
	if operation != nil {
		_ = t.Commit()
		return nil
	}
	if commitErr != nil {
		return fmt.Errorf("commit: %w", commitErr)
	}
	return t.Commit()
}
