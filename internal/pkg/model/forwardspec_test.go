package model

import "testing"

func TestPortForwardSpec_Complete(t *testing.T) {
	type fields struct {
		Namespace string
		Pod       string
		Service   string
		Remote    string
		Local     string
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"valid-pod-forward-random",
			fields{
				Namespace: "namespace",
				Pod:       "pod",
				Remote:    "http",
			},
			args{
				name: "name",
			},
			false,
		},
		{
			"valid-service-forward-random",
			fields{
				Namespace: "namespace",
				Service:   "service",
				Remote:    "http",
			},
			args{
				name: "name",
			},
			false,
		},
		{
			"valid-service-forward-port",
			fields{
				Namespace: "namespace",
				Service:   "service",
				Remote:    "http",
				Local:     "8080",
			},
			args{
				name: "name",
			},
			false,
		},
		{
			"valid-service-forward-addr",
			fields{
				Namespace: "namespace",
				Service:   "service",
				Remote:    "http",
				Local:     "0.0.0.0:8080",
			},
			args{
				name: "name",
			},
			false,
		},
		{
			"invalid-forward-missing-type",
			fields{
				Namespace: "namespace",
				Remote:    "http",
			},
			args{
				name: "name",
			},
			true,
		},
		{
			"invalid-forward-two-types",
			fields{
				Namespace: "namespace",
				Pod:       "pod",
				Service:   "service",
				Remote:    "http",
			},
			args{
				name: "name",
			},
			true,
		},
		{
			"invalid-pod-forward-missing-remote",
			fields{
				Namespace: "namespace",
				Pod:       "pod",
			},
			args{
				name: "name",
			},
			true,
		},
		{
			"invalid-pod-forward-local-format-1",
			fields{
				Namespace: "namespace",
				Pod:       "pod",
				Remote:    "http",
				Local:     "8080:",
			},
			args{
				name: "name",
			},
			true,
		},
		{
			"invalid-pod-forward-local-format-2",
			fields{
				Namespace: "namespace",
				Pod:       "pod",
				Remote:    "http",
				Local:     "0.0.0.0",
			},
			args{
				name: "name",
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := &PortForwardSpec{
				Namespace: tt.fields.Namespace,
				Pod:       tt.fields.Pod,
				Service:   tt.fields.Service,
				Remote:    tt.fields.Remote,
				Local:     tt.fields.Local,
			}
			if err := pf.Complete(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("Complete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
