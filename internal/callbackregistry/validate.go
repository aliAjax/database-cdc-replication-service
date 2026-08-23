package callbackregistry

import "fmt"

func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("empty event name")
	}
	return nil
}
