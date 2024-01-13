package forwarder

import (
	"context"
	"errors"
	"fmt"
	"github.com/tmsmr/k4wd/internal/pkg/kubeclient"
	"github.com/tmsmr/k4wd/internal/pkg/model"
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
	podBySelectorTimeout = 10 * time.Second
)

type Forwarder struct {
	spec  *model.PortForwardSpec
	Io    genericiooptions.IOStreams
	Ready chan struct{}
}

func New(spec *model.PortForwardSpec) *Forwarder {
	return &Forwarder{
		spec: spec,
		Io: genericiooptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
		Ready: make(chan struct{}),
	}
}

func (fwd *Forwarder) resolveServiceTarget(kc *kubeclient.Kubeclient) (*v1.Pod, int32, error) {
	service, err := kc.Clientset.CoreV1().Services(fwd.spec.Namespace).Get(context.TODO(), fwd.spec.Service, metav1.GetOptions{})
	if err != nil {
		return nil, 0, err
	}
	var svcPort int32
	val, err := strconv.Atoi(fwd.spec.Remote)
	if err == nil {
		svcPort = int32(val)
	} else {
		svcPort, err = util.LookupServicePortNumberByName(*service, fwd.spec.Remote)
		if err != nil {
			return nil, 0, err
		}
	}
	selector := labels.SelectorFromSet(service.Spec.Selector)
	sorter := func(pods []*v1.Pod) sort.Interface { return sort.Reverse(podutils.ActivePods(pods)) }
	pod, _, err := polymorphichelpers.GetFirstPod(kc.Clientset.CoreV1(), fwd.spec.Namespace, selector.String(), podBySelectorTimeout, sorter)
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
	pod, err := kc.Clientset.CoreV1().Pods(fwd.spec.Namespace).Get(context.TODO(), fwd.spec.Pod, metav1.GetOptions{})
	if err != nil {
		return nil, 0, err
	}
	var podPort int32
	val, err := strconv.Atoi(fwd.spec.Remote)
	if err == nil {
		podPort = int32(val)
	} else {
		podPort, err = util.LookupContainerPortNumberByName(*pod, fwd.spec.Remote)
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
	switch fwd.spec.Type {
	case model.PortForwardTypePod:
		pod, port, err = fwd.resolvePodTarget(kc)
		if err != nil {
			return err
		}
		break
	case model.PortForwardTypeService:
		pod, port, err = fwd.resolveServiceTarget(kc)
		if err != nil {
			return err
		}
		fwd.spec.Pod = pod.Name
		break
	default:
		panic("forward type unsupported")
	}
	fwd.spec.TargetPort = port
	if fwd.spec.LocalPort == 0 {
		local, err := randomLocalPort()
		if err != nil {
			return err
		}
		fwd.spec.LocalPort = int32(local)
	}
	if pod.Status.Phase != v1.PodRunning {
		return errors.New("pod not running")
	}
	transport, upgrader, err := spdy.RoundTripperFor(kc.Config)
	if err != nil {
		return err
	}
	req := kc.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(fwd.spec.Namespace).
		Name(pod.Name).
		SubResource("portforward")
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, req.URL())
	forwarder, err := portforward.NewOnAddresses(
		dialer,
		[]string{fwd.spec.LocalAddr},
		[]string{fmt.Sprintf("%d:%d", fwd.spec.LocalPort, fwd.spec.TargetPort)},
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
