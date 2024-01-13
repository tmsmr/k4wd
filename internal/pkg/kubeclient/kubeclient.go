package kubeclient

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

type Kubeclient struct {
	Kubeconfig string
	Config     *rest.Config
	Clientset  *kubernetes.Clientset
}

type ClientOption func(ff *Kubeclient)

func WithKubeconfig(path string) ClientOption {
	return func(kc *Kubeclient) {
		kc.Kubeconfig = path
	}
}

func New(opts ...ClientOption) (*Kubeclient, error) {
	kc := &Kubeclient{}
	for _, opt := range opts {
		opt(kc)
	}
	if kc.Kubeconfig == "" {
		home := os.Getenv("HOME")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		kc.Kubeconfig = filepath.Join(home, ".kube", "config")
	}
	config, err := clientcmd.BuildConfigFromFlags("", kc.Kubeconfig)
	if err != nil {
		return nil, err
	}
	kc.Config = config
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	kc.Clientset = clientset
	return kc, nil
}
