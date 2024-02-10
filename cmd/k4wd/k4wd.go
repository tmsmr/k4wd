package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tmsmr/k4wd/internal/pkg/config"
	"github.com/tmsmr/k4wd/internal/pkg/envfile"
	"github.com/tmsmr/k4wd/internal/pkg/forwarder"
	"github.com/tmsmr/k4wd/internal/pkg/kubeclient"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func start(confPath string, kubeconfPath string) {
	conf, err := config.Load(config.WithPath(confPath))
	must(err)
	entries := make([]string, 0)
	for n := range conf.Forwards {
		entries = append(entries, n)
	}
	log.Infof("loaded %s (relaxed=%t) containing %d entries: (%s)", conf.Path, conf.Relaxed, len(conf.Forwards), strings.Join(entries, ", "))

	var kc *kubeclient.Kubeclient
	if kubeconfPath == "" {
		kc, err = kubeclient.New()
	} else {
		kc, err = kubeclient.New(kubeclient.WithKubeconfig(kubeconfPath))
	}
	must(err)
	log.Infof("created Kubeclient for %s", kc.Kubeconfig)

	ef, err := envfile.New(confPath)
	must(err)
	log.Infof("initialized Envfile for %s", ef.Path())
	defer func() {
		log.Infof("removing %s", ef.Path())
		must(ef.Remove())
	}()

	fwds := make(map[string]*forwarder.Forwarder)
	for name, spec := range conf.Forwards {
		fwd, err := forwarder.New(name, spec, io.Discard)
		must(err)
		fwds[name] = fwd
	}

	var fwdsActive sync.WaitGroup
	stopFwds := make(chan struct{}, 1)

	for name, fwd := range fwds {
		fwdFailed := make(chan struct{}, 1)

		go func() {
			fwdsActive.Add(1)
			defer fwdsActive.Done()

			err := fwd.Run(kc, stopFwds)
			if err != nil {
				fwd.Active = false
				must(ef.Update(fwds))
				log.Warnf("%s failed: %v", name, err)
				if !conf.Relaxed {
					log.Errorf("%s failed with global relaxed=false, stopping all forwards", name)
					// TODO: panics for multiple failed forwards...
					close(stopFwds)
				}
				close(fwdFailed)
			}
		}()

		select {
		case <-fwd.Ready:
			fwd.Active = true
			must(ef.Update(fwds))
			log.Infof("%s ready: %s:%d -> %s, %s, %d", name, fwd.BindAddr, fwd.BindPort, fwd.Namespace, fwd.TargetPod, fwd.TargetPort)
			break
		case <-fwdFailed:
			break
		}
	}

	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-term
		log.Warnf("received SIGTERM, stopping all forwards")
		close(stopFwds)
	}()

	fwdsActive.Wait()
	log.Info("no active forwards left, exiting")
}

func main() {
	opts := parseOpts()
	if opts.debug {
		log.SetLevel(log.DebugLevel)
	}
	if opts.cmdMode == envMode {
		ef, err := envfile.New(opts.conf)
		must(err)
		content, err := ef.Load(opts.format)
		must(err)
		fmt.Print(string(content))
		return
	}
	start(opts.conf, opts.kubeconf)
}
