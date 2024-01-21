package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
)

type Forwardfile struct {
	Path     string `toml:"-"`
	Relaxed  bool
	Forwards map[string]Forward
}

func (ff *Forwardfile) Validate() error {
	if ff.Forwards == nil || len(ff.Forwards) == 0 {
		return fmt.Errorf("no forwards defined")
	}
	for _, forward := range ff.Forwards {
		if err := forward.Validate(); err != nil {
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
	if err := ff.Validate(); err != nil {
		return nil, err
	}
	return ff, nil
}
