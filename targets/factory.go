package targets

import (
	"errors"
)

type InitFunc func(options Options) (Target, error)

type Factory struct {
	targets map[string]InitFunc
}

var defaultFactory = Factory{
	targets: make(map[string]InitFunc),
}

func (f *Factory) RegisterProvider(id string, createFunc InitFunc) error {
	f.targets[id] = createFunc
	return nil
}

func (f *Factory) New(id string, options Options) (Target, error) {
	createFunc, ok := f.targets[id]
	if !ok {
		return nil, errors.New("bad target id type")
	}

	return createFunc(options)
}

func RegisterProvider(id string, createFunc InitFunc) error {
	defaultFactory.targets[id] = createFunc
	return nil
}

func New(id string, options Options) (Target, error) {
	return defaultFactory.New(id, options)
}
