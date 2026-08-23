package connectorcfg

func Merge(base Config, overrides map[string]string) Config {
	out := Load(base.Driver, base.Options)
	if len(overrides) == 0 {
		out.Options = base.Options
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
