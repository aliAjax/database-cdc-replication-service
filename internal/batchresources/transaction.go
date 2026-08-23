package batchresources

type Transaction interface {
	Commit() error
	Rollback() error
}

func Finish(tx Transaction, operation func() error) (err error) {
	defer func() {
		err = tx.Commit()
	}()
	if err = operation(); err != nil {
		_ = tx.Rollback()
		return err
	}
	return nil
}
