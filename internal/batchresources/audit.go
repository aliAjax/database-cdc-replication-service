package batchresources

import "errors"

type AuditWriter interface {
	Append(string) error
	Flush() error
}

func Record(writer AuditWriter, entries []string) error {
	var result error
	for _, entry := range entries {
		if err := writer.Append(entry); err != nil {
			result = errors.Join(result, err)
			break
		}
	}
	return errors.Join(result, writer.Flush())
}
