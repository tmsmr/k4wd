package forwarder

import (
	"github.com/tmsmr/k4wd/internal/pkg/config"
	"github.com/tmsmr/k4wd/internal/pkg/kubeclient"
	"k8s.io/client-go/tools/clientcmd/api"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		name string
		spec config.Forward
	}
	tests := []struct {
		name    string
		args    args
		want    *Forwarder
		wantErr bool
	}{
		{"minimal pod forward", args{"name", config.Forward{Pod: "pod", Remote: "http-alt"}}, &Forwarder{Namespace: defaultNamespace, BindAddr: defaultBindAddr, BindPort: 0, RandPort: true}, false},
		{"pod forward with namespace", args{"name", config.Forward{Namespace: func() *string { s := "namespace"; return &s }(), Pod: "pod", Remote: "http-alt"}}, &Forwarder{Namespace: "namespace", BindAddr: defaultBindAddr, BindPort: 0, RandPort: true}, false},
		{"pod forward with invalid local", args{"name", config.Forward{Local: defaultBindAddr}}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.name, tt.args.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got.Namespace != tt.want.Namespace {
				t.Errorf("New() Namespace = %v, want %v", got.Namespace, tt.want.Namespace)
			}
			if got.BindAddr != tt.want.BindAddr {
				t.Errorf("New() BindAddr = %v, want %v", got.BindAddr, tt.want.BindAddr)
			}
			if got.BindPort != tt.want.BindPort {
				t.Errorf("New() BindPort = %v, want %v", got.BindPort, tt.want.BindPort)
			}
			if got.RandPort != tt.want.RandPort {
				t.Errorf("New() RandPort = %v, want %v", got.RandPort, tt.want.RandPort)
			}
		})
	}
}

func TestForwarder_Run(t *testing.T) {
	type fields struct {
		Name    string
		Forward config.Forward
	}
	type args struct {
		kc *kubeclient.Kubeclient
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"invalid kubeclient", fields{"invalid-kubeclient", config.Forward{}}, args{&kubeclient.Kubeclient{APIConfig: &api.Config{}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fwd := &Forwarder{
				Name:    tt.fields.Name,
				Forward: tt.fields.Forward,
			}
			if err := fwd.Run(tt.args.kc, make(chan struct{})); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
