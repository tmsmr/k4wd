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
	"sync"
	"syscall"
	"time"
)

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func run(opts cmdOpts) {
	conf, err := config.Load(config.WithPath(opts.conf))
	must(err)
	log.Debugf("loaded %s containing %d entries", conf.Path, len(conf.Forwards))

	var kc *kubeclient.Kubeclient
	if opts.kubeconf == "" {
		kc, err = kubeclient.New()
	} else {
		kc, err = kubeclient.New(kubeclient.WithKubeconfig(opts.kubeconf))
	}
	must(err)
	log.Debugf("created Kubeclient for %s", kc.Kubeconfig)

	ef, err := envfile.New(opts.conf)
	must(err)
	log.Debugf("initialized Envfile %s", ef.Path())
	defer func() {
		log.Debugf("removing %s", ef.Path())
		must(ef.Remove())
	}()

	fwds := make(map[string]*forwarder.Forwarder)
	for name, spec := range conf.Forwards {
		stdout := io.Discard
		if opts.debug {
			stdout = os.Stdout
		}
		fwd, err := forwarder.New(name, spec, stdout)
		must(err)
		fwds[name] = fwd
	}

	must(ef.Update(fwds))

	log.Infof("starting %d forwards", len(fwds))

	//TODO: do we really need several waitgroups/channels to control the flow?

	var active sync.WaitGroup
	// chan we use to signal the forwards to terminate
	stop := make(chan struct{}, 1)
	// chan we use to signal the main goroutine to initiate shutdown
	shutdown := make(chan bool, len(fwds))

	for name, fwd := range fwds {
		failed := make(chan struct{}, 1)

		go func() {
			active.Add(1)
			defer active.Done()

			err := fwd.Run(kc, stop)
			if err != nil {
				if !conf.Relaxed {
					log.Errorf("%s failed: %v", name, err)
					shutdown <- true
				} else {
					log.Warnf("%s failed: %v", name, err)
				}
				close(failed)
			}
		}()

		// wait for the forward to either be ready or have failed immediately to enforce sequential startup
		select {
		case <-fwd.Ready:
			log.Infof("%s ready (%s)", fwd.Name, fwd.String())
		case <-failed:
			break
		}
	}

	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// wait for either a forward to fail with relaxed mode disabled or SIGINT/SIGTERM coming in
		select {
		case <-shutdown:
			log.Errorf("at least one forward failed with relaxed mode disabled, shutting down...")
		case <-term:
			log.Warnf("received SIGTERM, shutting down...")
		}
		// signal all forwards to terminate
		close(stop)
	}()

	// wait for all forwards to have completed
	active.Wait()
	log.Info("no active forwards left, exiting")
}

func main() {
	opts := parseOpts()
	if opts.debug {
		log.SetLevel(log.DebugLevel)
	}
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.TimeOnly,
	})
	if opts.cmdMode == envMode {
		ef, err := envfile.New(opts.conf)
		must(err)
		content, err := ef.Load(opts.format)
		must(err)
		fmt.Print(string(content))
		return
	}
	run(opts)
}
