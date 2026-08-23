package connectorcfg

type Config struct {
	Driver  string
	Options map[string]string
}

func Load(driver string, input map[string]string) Config {
	var options map[string]string
	if input != nil {
		options = make(map[string]string, len(input))
	}
	for key, value := range input {
		options[key] = value
	}
	return Config{Driver: driver, Options: options}
}
