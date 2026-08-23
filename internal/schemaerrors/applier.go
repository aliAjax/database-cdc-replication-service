package schemaerrors

import "fmt"

type Store interface{ Apply(Change) error }

func Apply(payload []byte, store Store) error {
	change, err := Decode(payload)
	if err != nil {
		return err
	}
	if err := Validate(change); err != nil {
		_ = store.Apply(change)
		return err
	}
	if err := store.Apply(change); err != nil {
		return fmt.Errorf("apply schema change: %w", err)
	}
	return nil
}
