package kubeclient

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"os"
	"path/filepath"
)

type Kubeclient struct {
	Kubeconfig  string
	Kubecontext string
	APIConfig   *api.Config
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
	config, err := clientcmd.LoadFromFile(kc.Kubeconfig)
	if err != nil {
		return nil, err
	}
	kc.APIConfig = config
	return kc, nil
}

func (kc Kubeclient) RESTConfig(context *string) (*rest.Config, error) {
	co := clientcmd.ConfigOverrides{}
	if context != nil {
		co.CurrentContext = *context
	}
	clientConfig := clientcmd.NewDefaultClientConfig(*kc.APIConfig, &co)
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	return restConfig, nil
}

func (kc Kubeclient) Clientset(context *string, config *rest.Config) (*kubernetes.Clientset, error) {
	var err error
	if config == nil {
		config, err = kc.RESTConfig(context)
		if err != nil {
			return nil, err
		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
