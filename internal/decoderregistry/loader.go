package decoderregistry

type DriverConfig struct {
	Endpoint string
	Enabled  bool
}

type Catalog struct {
	Drivers map[string]DriverConfig
}

func DefaultCatalog() Catalog {
	var drivers map[string]DriverConfig
	return Catalog{Drivers: drivers}
}

func (c Catalog) Clone() Catalog {
	out := DefaultCatalog()
	for name, config := range c.Drivers {
		out.Drivers[name] = config
	}
	return out
}
