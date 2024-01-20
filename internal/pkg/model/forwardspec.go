package model

import (
	"fmt"
	"strconv"
	"strings"
)

type PortForwardType int

const (
	PortForwardTypePod PortForwardType = iota
	PortForwardTypeService
)

type PortForwardSpec struct {
	Name       string          `toml:"-"`
	Type       PortForwardType `toml:"-"`
	Namespace  string
	Pod        string
	Service    string
	Remote     string
	Local      string
	LocalAddr  string `toml:"-"`
	LocalPort  int32  `toml:"-"`
	TargetPort int32  `toml:"-"`
	Active     bool   `toml:"-"`

	Context *string
}

func (pf *PortForwardSpec) Complete(name string) error {
	pf.Name = name
	if pf.Namespace == "" {
		pf.Namespace = "default"
	}
	if (pf.Pod == "" && pf.Service == "") || (pf.Pod != "" && pf.Service != "") {
		return fmt.Errorf("either pod (x)or service must be specified")
	}
	if pf.Pod != "" {
		pf.Type = PortForwardTypePod
	} else {
		pf.Type = PortForwardTypeService
	}
	if pf.Remote == "" {
		return fmt.Errorf("remote (named) port must be specified")
	}
	if pf.Local == "" {
		pf.LocalAddr = "127.0.0.1"
	} else {
		comps := strings.Split(pf.Local, ":")
		var port string
		switch len(comps) {
		case 1:
			port = comps[0]
			pf.LocalAddr = "127.0.0.1"
			break
		case 2:
			pf.LocalAddr = comps[0]
			port = comps[1]
			break
		default:
			return fmt.Errorf("invalid local port format")
		}
		val, err := strconv.Atoi(port)
		if err != nil {
			return err
		}
		pf.LocalPort = int32(val)
	}
	return nil
}

func (pf *PortForwardSpec) String() string {
	return fmt.Sprintf("Name: '%s', Type: %d, Namespace: '%s', Pod: '%s', Service: '%s', Remote: '%s', Local: '%s', LocalAddr: '%s', LocalPort: %d, TargetPort: %d, Active: %t",
		pf.Name, pf.Type, pf.Namespace, pf.Pod, pf.Service, pf.Remote, pf.Local, pf.LocalAddr, pf.LocalPort, pf.TargetPort, pf.Active)
}
