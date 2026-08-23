package schemaerrors

import (
	"errors"
	"fmt"
)

func Validate(change Change) error {
	if change.Table == "" || len(change.Columns) == 0 {
		return fmt.Errorf("validate schema change: %w", ErrSchemaCorrupt)
	}
	return nil
}

func IsCorrupt(err error) bool { return errors.Is(err, ErrSchemaCorrupt) }
