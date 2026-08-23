package connectorcfg

func Merge(base Config, overrides map[string]string) Config {
	out := Load(base.Driver, base.Options)
	if out.Options == nil {
		out.Options = make(map[string]string)
	}
	for key, value := range overrides {
		out.Options[key] = value
	}
	return out
}

func WithDefault(config Config, key, value string) Config {
	out := Merge(config, nil)
	if _, ok := out.Options[key]; !ok {
		out.Options[key] = value
	}
	return out
}
