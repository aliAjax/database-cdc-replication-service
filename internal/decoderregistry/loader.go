package decoderregistry

type DriverConfig struct {
	Endpoint string
	Enabled  bool
}

type Catalog struct {
	Drivers map[string]DriverConfig
}

func DefaultCatalog() Catalog {
	return Catalog{Drivers: make(map[string]DriverConfig)}
}

func (c Catalog) Clone() Catalog {
	out := DefaultCatalog()
	for name, config := range c.Drivers {
		out.Drivers[name] = config
	}
	return out
}
