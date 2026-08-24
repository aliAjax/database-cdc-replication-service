package eventhandlers

import "fmt"

func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: empty name", ErrUnknownEvent)
	}
	return nil
}
