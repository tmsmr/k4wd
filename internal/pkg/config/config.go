package config

import (
	"github.com/BurntSushi/toml"
	"github.com/tmsmr/k4wd/internal/pkg/model"
)

type Forwardfile struct {
	Path     string `toml:"-"`
	Relaxed  bool
	Forwards map[string]*model.PortForwardSpec
}

func (ff *Forwardfile) Complete() error {
	for name, spec := range ff.Forwards {
		if err := spec.Complete(name); err != nil {
			return err
		}
	}
	return nil
}

type ForwardfileOption func(ff *Forwardfile)

func WithPath(path string) ForwardfileOption {
	return func(ff *Forwardfile) {
		ff.Path = path
	}
}

func Load(opts ...ForwardfileOption) (*Forwardfile, error) {
	ff := &Forwardfile{}
	for _, opt := range opts {
		opt(ff)
	}
	if ff.Path == "" {
		ff.Path = "Forwardfile"
	}
	_, err := toml.DecodeFile(ff.Path, &ff)
	if err != nil {
		return nil, err
	}
	if err := ff.Complete(); err != nil {
		return nil, err
	}
	return ff, nil
}
