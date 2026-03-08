package registry

type RegisteredRegistrar struct {
	Registrar *EventRegistrar
}

func ProvideRegisteredRegistrar(registrar *EventRegistrar) (*RegisteredRegistrar, func(), error) {
	if err := registrar.Register(); err != nil {
		return nil, nil, err
	}
	return &RegisteredRegistrar{Registrar: registrar}, func() {
		_ = registrar.Stop()
	}, nil
}
