package main

import (
	"flag"
	"github.com/tmsmr/k4wd/internal/pkg/envfile"
)

type cmdMode int

const (
	runMode cmdMode = iota
	envMode
)

type cmdOpts struct {
	cmdMode
	debug    bool
	conf     string
	kubeconf string
	format   envfile.EnvFormat
}

func parseOpts() cmdOpts {
	opts := cmdOpts{}
	e := flag.Bool("e", false, "print environment instead of running k4wd")
	o := flag.String("o", "env", "output format for environment (env, no-export, json, ps, cmd)")
	flag.BoolVar(&opts.debug, "d", false, "enable debug logging")
	flag.StringVar(&opts.conf, "f", "Forwardfile", "path to Forwardfile (context)")
	flag.StringVar(&opts.kubeconf, "k", "", "alternative path to kubeconfig")
	flag.Parse()
	if *e {
		opts.cmdMode = envMode
		switch *o {
		case "no-export":
			opts.format = envfile.FormatNoExport
			break
		case "json":
			opts.format = envfile.FormatJSON
			break
		case "ps":
			opts.format = envfile.FormatPS
			break
		case "cmd":
			opts.format = envfile.FormatCmd
			break
		default:
			opts.format = envfile.FormatDefault
		}
	} else {
		opts.cmdMode = runMode
	}
	return opts
}
