package schemaerrors

import "fmt"

func Validate(change Change) error {
	if change.Table == "" || len(change.Columns) == 0 {
		return fmt.Errorf("validate schema change: %w", ErrSchemaCorrupt)
	}
	return nil
}

func IsCorrupt(err error) bool { return err == ErrSchemaCorrupt }
