package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"strconv"
	"strings"
)

type ForwardType int

const (
	ForwardTypePod ForwardType = iota
	ForwardTypeDeployment
	ForwardTypeService
)

type Forward struct {
	Context    *string
	Namespace  *string
	Pod        string
	Deployment string
	Service    string
	Remote     string
	Local      string
}

func (f *Forward) Type() ForwardType {
	if f.Deployment != "" {
		return ForwardTypeDeployment
	}
	if f.Service != "" {
		return ForwardTypeService
	}
	return ForwardTypePod
}

func (f *Forward) LocalAddr() (string, int32, error) {
	var port = 0
	var addr = ""
	if f.Local == "" {
		return addr, int32(port), nil
	}
	parts := strings.Split(f.Local, ":")
	var err error
	switch len(parts) {
	case 1:
		port, err = strconv.Atoi(parts[0])
		break
	case 2:
		addr = parts[0]
		port, err = strconv.Atoi(parts[1])
		break
	default:
		return "", 0, fmt.Errorf("invalid local port format")
	}
	if err != nil {
		return "", 0, err
	}
	return addr, int32(port), nil
}

func (f *Forward) Validate() error {
	resources := 0
	for _, s := range []string{f.Pod, f.Deployment, f.Service} {
		if s != "" {
			resources++
		}
	}
	if resources != 1 {
		return fmt.Errorf("exactly one of pod, deployment or service must be specified")
	}
	if f.Remote == "" {
		return fmt.Errorf("remote (named) port must be specified")
	}
	_, _, err := f.LocalAddr()
	if err != nil {
		return err
	}
	return nil
}

type Forwardfile struct {
	Path     string `toml:"-"`
	Relaxed  bool
	Forwards map[string]Forward
}

func (ff *Forwardfile) Validate() error {
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
