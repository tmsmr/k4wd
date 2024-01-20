package forwarder

import (
	"context"
	"errors"
	"fmt"
	"github.com/tmsmr/k4wd/internal/pkg/config"
	"github.com/tmsmr/k4wd/internal/pkg/kubeclient"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/util"
	"k8s.io/kubectl/pkg/util/podutils"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

const (
	defaultNamespace     = "default"
	defaultLocalAddr     = "127.0.0.1"
	podBySelectorTimeout = 10 * time.Second
)

type Forwarder struct {
	Name            string
	Forward         config.Forward
	Namespace       string
	LocalAddr       string
	RandomLocalPort bool
	LocalPort       int32
	TargetPod       string
	TargetPort      int32
	Active          bool
	Io              genericiooptions.IOStreams
	Ready           chan struct{}
}

func New(name string, spec config.Forward) (*Forwarder, error) {
	fwd := &Forwarder{
		Name:    name,
		Forward: spec,
		Io: genericiooptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
		Ready: make(chan struct{}),
	}
	if spec.Namespace != nil {
		fwd.Namespace = *spec.Namespace
	} else {
		fwd.Namespace = defaultNamespace
	}
	addr, port, err := spec.LocalAddr()
	if err != nil {
		return nil, err
	}
	fwd.LocalAddr = addr
	if fwd.LocalAddr == "" {
		fwd.LocalAddr = defaultLocalAddr
	}
	fwd.LocalPort = port
	if fwd.LocalPort == 0 {
		fwd.RandomLocalPort = true
	}
	return fwd, nil
}

func (fwd *Forwarder) resolveServiceTarget(kc *kubeclient.Kubeclient) (*v1.Pod, int32, error) {
	cs, err := kc.Clientset(fwd.Forward.Context)
	if err != nil {
		return nil, 0, err
	}
	service, err := cs.CoreV1().Services(fwd.Namespace).Get(context.TODO(), fwd.Forward.Service, metav1.GetOptions{})
	if err != nil {
		return nil, 0, err
	}
	var svcPort int32
	val, err := strconv.Atoi(fwd.Forward.Remote)
	if err == nil {
		svcPort = int32(val)
	} else {
		svcPort, err = util.LookupServicePortNumberByName(*service, fwd.Forward.Remote)
		if err != nil {
			return nil, 0, err
		}
	}
	selector := labels.SelectorFromSet(service.Spec.Selector)
	sorter := func(pods []*v1.Pod) sort.Interface { return sort.Reverse(podutils.ActivePods(pods)) }
	pod, _, err := polymorphichelpers.GetFirstPod(cs.CoreV1(), fwd.Namespace, selector.String(), podBySelectorTimeout, sorter)
	if err != nil {
		return nil, 0, err
	}
	targetPort, err := util.LookupContainerPortNumberByServicePort(*service, *pod, svcPort)
	if err != nil {
		return nil, 0, err
	}
	return pod, targetPort, nil
}

func (fwd *Forwarder) resolvePodTarget(kc *kubeclient.Kubeclient) (*v1.Pod, int32, error) {
	cs, err := kc.Clientset(fwd.Forward.Context)
	if err != nil {
		return nil, 0, err
	}
	pod, err := cs.CoreV1().Pods(fwd.Namespace).Get(context.TODO(), fwd.Forward.Pod, metav1.GetOptions{})
	if err != nil {
		return nil, 0, err
	}
	var podPort int32
	val, err := strconv.Atoi(fwd.Forward.Remote)
	if err == nil {
		podPort = int32(val)
	} else {
		podPort, err = util.LookupContainerPortNumberByName(*pod, fwd.Forward.Remote)
		if err != nil {
			return nil, 0, err
		}
	}
	return pod, podPort, nil
}

func (fwd *Forwarder) Run(kc *kubeclient.Kubeclient, stop chan struct{}) error {
	var pod *v1.Pod
	var port int32
	var err error
	switch fwd.Forward.Type() {
	case config.ForwardTypePod:
		pod, port, err = fwd.resolvePodTarget(kc)
		if err != nil {
			return err
		}
		break
	case config.ForwardTypeService:
		pod, port, err = fwd.resolveServiceTarget(kc)
		if err != nil {
			return err
		}
		fwd.TargetPod = pod.Name
		break
	default:
		panic("forward type unsupported")
	}
	fwd.TargetPort = port
	if fwd.RandomLocalPort {
		local, err := randomLocalPort()
		if err != nil {
			return err
		}
		fwd.LocalPort = int32(local)
	}
	if pod.Status.Phase != v1.PodRunning {
		return errors.New("pod not running")
	}
	rc, err := kc.RESTConfig(fwd.Forward.Context)
	if err != nil {
		return err
	}
	transport, upgrader, err := spdy.RoundTripperFor(rc)
	if err != nil {
		return err
	}
	cs, err := kc.Clientset(fwd.Forward.Context)
	if err != nil {
		return err
	}
	req := cs.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(fwd.Namespace).
		Name(pod.Name).
		SubResource("portforward")
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, req.URL())
	forwarder, err := portforward.NewOnAddresses(
		dialer,
		[]string{fwd.LocalAddr},
		[]string{fmt.Sprintf("%d:%d", fwd.LocalPort, fwd.TargetPort)},
		stop,
		fwd.Ready,
		fwd.Io.Out,
		fwd.Io.ErrOut,
	)
	if err != nil {
		return err
	}
	return forwarder.ForwardPorts()
}

func randomLocalPort() (port int, err error) {
	a, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", a)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
