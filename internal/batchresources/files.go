package batchresources

import "errors"

type File interface {
	Write([]byte) error
	Close() error
}

func WriteFiles(files []File, payload []byte) error {
	for _, file := range files {
		if err := file.Write(payload); err != nil {
			_ = file.Close()
			return err
		}
		if err := file.Close(); err != nil {
			return err
		}
	}
	return nil
}

var ErrTransaction = errors.New("transaction failed")
