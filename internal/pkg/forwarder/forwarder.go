package forwarder

import (
	"context"
	"fmt"
	"github.com/tmsmr/k4wd/internal/pkg/config"
	"github.com/tmsmr/k4wd/internal/pkg/kubeclient"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/util"
	"k8s.io/kubectl/pkg/util/podutils"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

const (
	defaultNamespace     = "default"
	defaultBindAddr      = "127.0.0.1"
	podBySelectorTimeout = 5 * time.Second
)

type Forwarder struct {
	Name string
	config.Forward
	Clients *kubernetes.Clientset
	Io      genericiooptions.IOStreams
	Ready   chan struct{}
	Active  bool

	Namespace  string
	BindAddr   string
	BindPort   int32
	RandPort   bool
	TargetPod  string
	TargetPort int32
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
	fwd.BindAddr = addr
	if fwd.BindAddr == "" {
		fwd.BindAddr = defaultBindAddr
	}
	fwd.BindPort = port
	if fwd.BindPort == 0 {
		fwd.RandPort = true
	}

	return fwd, nil
}

// resolvePodTarget looks up a pod by name and resolves the target port.
func (fwd *Forwarder) resolvePodTarget(name string) (*v1.Pod, int32, error) {
	pod, err := fwd.Clients.CoreV1().Pods(fwd.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, 0, err
	}

	var podPort int32
	val, err := strconv.Atoi(fwd.Remote)
	if err == nil {
		podPort = int32(val)
	} else {
		podPort, err = util.LookupContainerPortNumberByName(*pod, fwd.Remote)
		if err != nil {
			return nil, 0, err
		}
	}

	return pod, podPort, nil
}

// resolveDeploymentTarget looks up the deployment, finds a matching pod and resolves the target port.
func (fwd *Forwarder) resolveDeploymentTarget() (*v1.Pod, int32, error) {
	deployment, err := fwd.Clients.AppsV1().Deployments(fwd.Namespace).Get(context.TODO(), fwd.Deployment, metav1.GetOptions{})
	if err != nil {
		return nil, 0, err
	}

	selector := labels.SelectorFromSet(deployment.Spec.Selector.MatchLabels)
	sorter := func(pods []*v1.Pod) sort.Interface { return sort.Reverse(podutils.ActivePods(pods)) }
	pod, _, err := polymorphichelpers.GetFirstPod(fwd.Clients.CoreV1(), fwd.Namespace, selector.String(), podBySelectorTimeout, sorter)
	if err != nil {
		return nil, 0, err
	}

	return fwd.resolvePodTarget(pod.Name)
}

// resolveServiceTarget looks up the service, finds a matching pod and resolves the target port.
func (fwd *Forwarder) resolveServiceTarget() (*v1.Pod, int32, error) {
	service, err := fwd.Clients.CoreV1().Services(fwd.Namespace).Get(context.TODO(), fwd.Service, metav1.GetOptions{})
	if err != nil {
		return nil, 0, err
	}

	var svcPort int32
	val, err := strconv.Atoi(fwd.Remote)
	if err == nil {
		svcPort = int32(val)
	} else {
		svcPort, err = util.LookupServicePortNumberByName(*service, fwd.Remote)
		if err != nil {
			return nil, 0, err
		}
	}

	selector := labels.SelectorFromSet(service.Spec.Selector)
	sorter := func(pods []*v1.Pod) sort.Interface { return sort.Reverse(podutils.ActivePods(pods)) }
	pod, _, err := polymorphichelpers.GetFirstPod(fwd.Clients.CoreV1(), fwd.Namespace, selector.String(), podBySelectorTimeout, sorter)
	if err != nil {
		return nil, 0, err
	}

	targetPort, err := util.LookupContainerPortNumberByServicePort(*service, *pod, svcPort)
	if err != nil {
		return nil, 0, err
	}
	return pod, targetPort, nil
}

// Run starts the port forwarding and blocks until it is stopped.
func (fwd *Forwarder) Run(kc *kubeclient.Kubeclient, stop chan struct{}) error {
	cs, err := kc.Clientset(fwd.Context, nil)
	if err != nil {
		return err
	}
	fwd.Clients = cs

	// TODO: make this more generic, compare portforward.go in kubectl
	// resolve target pod and port based on forward type
	var pod *v1.Pod
	var port int32
	switch fwd.Type() {
	case config.ForwardTypePod:
		pod, port, err = fwd.resolvePodTarget(fwd.Pod)
		break
	case config.ForwardTypeDeployment:
		pod, port, err = fwd.resolveDeploymentTarget()
		break
	case config.ForwardTypeService:
		pod, port, err = fwd.resolveServiceTarget()
		break
	default:
		panic("forward type unsupported")
	}
	if err != nil {
		return err
	}

	// try to find unsupported protocols
	for _, cont := range pod.Spec.Containers {
		for _, portSpec := range cont.Ports {
			if portSpec.ContainerPort != port {
				continue
			}
			if portSpec.Protocol != v1.ProtocolTCP {
				return fmt.Errorf("unsupported protocol: %s", portSpec.Protocol)
			}
		}
	}

	fwd.TargetPod = pod.Name
	fwd.TargetPort = port
	if fwd.RandPort {
		local, err := randomLocalPort()
		if err != nil {
			return err
		}
		fwd.BindPort = int32(local)
	}

	// ensure pod is running
	if pod.Status.Phase != v1.PodRunning {
		return fmt.Errorf("target pod not running: %s", fwd.TargetPod)
	}

	// start forwarding
	rc, err := kc.RESTConfig(fwd.Context)
	if err != nil {
		return err
	}
	transport, upgrader, err := spdy.RoundTripperFor(rc)
	if err != nil {
		return err
	}
	req := fwd.Clients.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(fwd.Namespace).
		Name(pod.Name).
		SubResource("portforward")
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, req.URL())
	forwarder, err := portforward.NewOnAddresses(
		dialer,
		[]string{fwd.BindAddr},
		[]string{fmt.Sprintf("%d:%d", fwd.BindPort, fwd.TargetPort)},
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
