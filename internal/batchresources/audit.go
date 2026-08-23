package batchresources

type AuditWriter interface {
	Append(string) error
	Flush() error
}

func Record(writer AuditWriter, entries []string) error {
	for _, entry := range entries {
		if err := writer.Append(entry); err != nil {
			return err
		}
	}
	return writer.Flush()
}
