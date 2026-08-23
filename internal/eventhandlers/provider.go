package eventhandlers

type Provider struct{ Registry *Registry }

func (p *Provider) Ready() bool { return p != nil && p.Registry != nil }
func (p *Provider) Install(name string, h Handler) error {
	if p == nil || p.Registry == nil {
		return ErrHandlerUnavailable
	}
	return p.Registry.Register(name, h)
}
