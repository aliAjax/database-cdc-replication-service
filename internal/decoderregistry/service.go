package decoderregistry

type Registry struct {
	drivers map[string]DriverConfig
}

func NewRegistry() *Registry {
	return &Registry{drivers: make(map[string]DriverConfig)}
}

func (r *Registry) Register(name string, config DriverConfig) {
	r.drivers[name] = config
}

func (r *Registry) Snapshot() map[string]DriverConfig {
	out := make(map[string]DriverConfig, len(r.drivers))
	for name, config := range r.drivers {
		out[name] = config
	}
	return out
}
