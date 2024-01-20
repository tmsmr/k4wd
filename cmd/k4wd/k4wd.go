package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tmsmr/k4wd/internal/pkg/config"
	"github.com/tmsmr/k4wd/internal/pkg/envfile"
	"github.com/tmsmr/k4wd/internal/pkg/forwarder"
	"github.com/tmsmr/k4wd/internal/pkg/kubeclient"
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

	ef, err := envfile.New()
	must(err)
	log.Infof("initialized Envfile for %s", ef.Path())
	defer func() {
		log.Infof("removing %s", ef.Path())
		must(ef.Remove())
	}()

	fwds := make(map[string]*forwarder.Forwarder)
	for name, spec := range conf.Forwards {
		fwd, err := forwarder.New(name, spec)
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
					close(stopFwds)
				}
				close(fwdFailed)
			}
		}()

		select {
		case <-fwd.Ready:
			fwd.Active = true
			must(ef.Update(fwds))
			log.Infof("%s ready: %s:%d -> %s, %s, %d", name, fwd.LocalAddr, fwd.LocalPort, fwd.Namespace, fwd.TargetPod, fwd.TargetPort)
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
	envGet := flag.Bool("e", false, "get environment instead of running k4wd")
	envPath := flag.String("p", "", "copy environment to the specified path instead of printing to STDOUT")
	confPath := flag.String("f", "Forwardfile", "path to Forwardfile")
	kubePath := flag.String("k", "", "alternative path to kubeconfig")
	flag.Parse()
	if *envGet {
		ef, err := envfile.New()
		must(err)
		if *envPath == "" {
			content, err := ef.Load()
			must(err)
			fmt.Print(string(content))
		} else {
			must(ef.Copy(*envPath))
		}
	} else {
		start(*confPath, *kubePath)
	}
}
