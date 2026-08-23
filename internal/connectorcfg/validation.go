package connectorcfg

import (
	"errors"
	"fmt"
)

func ValidateDriver(driver string, config Config) error {
	if driver != config.Driver {
		return fmt.Errorf("connector driver mismatch: %s != %s", driver, config.Driver)
	}
	if config.Options == nil {
		return nil
	}
	return nil
}

func ValidateWith(validator Validator, config Config) error {
	if validator == nil {
		return errors.New("connector validator is unavailable")
	}
	return validator.Validate(config)
}
