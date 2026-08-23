package connectorcfg

type Validator interface {
	Validate(Config) error
}

type staticValidator struct{ driver string }

func (v *staticValidator) Validate(config Config) error {
	return ValidateDriver(v.driver, config)
}

func Provider(driver string) Validator {
	if driver == "" {
		return nil
	}
	return &staticValidator{driver: driver}
}
