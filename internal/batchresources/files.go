package batchresources

import "errors"

type File interface {
	Write([]byte) error
	Close() error
}

func WriteFiles(files []File, payload []byte) error {
	for _, file := range files {
		defer file.Close()
		if err := file.Write(payload); err != nil {
			return err
		}
	}
	return nil
}

var ErrTransaction = errors.New("transaction failed")
