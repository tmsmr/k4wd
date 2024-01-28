package forwarder

/*
Tests in this file are integration tests, they require some prerequisites to run:
- a running Kubernetes cluster and the kubeconfig file at the default location
- deployed manifests from the https://github.com/tmsmr/k4wd/tree/main/test/integration directory

The tests will be skipped if K4WD_INTEGRATION_TESTS is not set.

To validate the forward targets, the workloads run https://github.com/tmsmr/context with the K4WD_TYPE environment variable set.
After a Forwarder is started, a request is sent to the forwarded port and the response is checked for the expected value of K4WD_TYPE.
*/

import (
	"encoding/json"
	"fmt"
	"github.com/tmsmr/k4wd/internal/pkg/config"
	"github.com/tmsmr/k4wd/internal/pkg/kubeclient"
	"net/http"
	"os"
	"strconv"
	"testing"
)

func setupIntegrationTests() (bool, error, *kubeclient.Kubeclient) {
	if os.Getenv("K4WD_INTEGRATION_TESTS") == "" {
		return true, nil, nil
	}
	kc, err := kubeclient.New()
	if err != nil {
		return false, err, kc
	}
	return false, nil, kc
}

func checkForward(fwd *Forwarder, kc *kubeclient.Kubeclient) (string, error) {
	stop := make(chan struct{}, 1)
	defer close(stop)
	done := make(chan error, 1)
	go func() {
		err := fwd.Run(kc, stop)
		if err != nil {
			done <- err
		}
		close(done)
	}()
	select {
	case <-fwd.Ready:
		url := "http://" + fwd.BindAddr + ":" + strconv.Itoa(int(fwd.BindPort))
		resp, err := http.Get(url)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		data := struct {
			Env map[string]string `json:"env"`
		}{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return "", err
		}
		if _, ok := data.Env["K4WD_TYPE"]; !ok {
			return "", fmt.Errorf("missing K4WD_TYPE env variable")
		}
		return data.Env["K4WD_TYPE"], nil
	case err := <-done:
		return "", err
	}
}

func TestForwarder_Integration_Run(t *testing.T) {
	skip, err, kc := setupIntegrationTests()
	if skip {
		t.Skip("skipping integration test")
	}
	if err != nil {
		t.Fatal(err)
	}
	type fields struct {
		Name    string
		Forward config.Forward
	}
	type args struct {
		kc       *kubeclient.Kubeclient
		k4wdType string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"missing/invalid forward type", fields{"int-test-invalid-type", config.Forward{
			Namespace: func() *string { s := "k4wd"; return &s }(),
			Remote:    "http-alt",
		}}, args{kc, "pod"}, true},

		{"valid pod forward", fields{"int-test-po", config.Forward{
			Pod:       "int-test-po",
			Namespace: func() *string { s := "k4wd"; return &s }(),
			Remote:    "http-alt",
		}}, args{kc, "pod"}, false},

		{"unknown pod forward", fields{"int-test-po-unknown", config.Forward{
			Pod:       "int-test-po-unknown",
			Namespace: func() *string { s := "k4wd"; return &s }(),
			Remote:    "http-alt",
		}}, args{kc, ""}, true},

		{"valid numerical port pod forward", fields{"int-test-po-numerical", config.Forward{
			Pod:       "int-test-po",
			Namespace: func() *string { s := "k4wd"; return &s }(),
			Remote:    "8080",
		}}, args{kc, "pod"}, false},

		{"invalid named port pod forward", fields{"int-test-po-invalid-named", config.Forward{
			Pod:       "int-test-po",
			Namespace: func() *string { s := "k4wd"; return &s }(),
			Remote:    "mysql",
		}}, args{kc, ""}, true},

		{"udp port pod forward", fields{"int-test-po-udp", config.Forward{
			Pod:       "int-test-po-multiple-udp",
			Namespace: func() *string { s := "k4wd"; return &s }(),
			Remote:    "8081",
		}}, args{kc, ""}, true},

		{"pod forward connection refused", fields{"int-test-po-refused", config.Forward{
			Pod:       "int-test-po-multiple-udp",
			Namespace: func() *string { s := "k4wd"; return &s }(),
			Remote:    "8082",
		}}, args{kc, ""}, true},

		{"valid deployment forward", fields{"int-test-de", config.Forward{
			Deployment: "int-test-de",
			Namespace:  func() *string { s := "k4wd"; return &s }(),
			Remote:     "http-alt",
		}}, args{kc, "deployment"}, false},

		{"unknown deployment forward", fields{"int-test-de-unknown", config.Forward{
			Deployment: "int-test-de-unknown",
			Namespace:  func() *string { s := "k4wd"; return &s }(),
			Remote:     "http-alt",
		}}, args{kc, ""}, true},

		{"zero pods deployment forward", fields{"int-test-de-zero", config.Forward{
			Deployment: "int-test-de-zero",
			Namespace:  func() *string { s := "k4wd"; return &s }(),
			Remote:     "http-alt",
		}}, args{kc, ""}, true},

		{"valid service forward", fields{"int-test-svc", config.Forward{
			Service:   "int-test-svc",
			Namespace: func() *string { s := "k4wd"; return &s }(),
			Remote:    "http-alt",
		}}, args{kc, "deployment"}, false},

		{"valid numerical port service forward", fields{"int-test-svc-numerical", config.Forward{
			Service:   "int-test-svc",
			Namespace: func() *string { s := "k4wd"; return &s }(),
			Remote:    "8080",
		}}, args{kc, "deployment"}, false},

		{"invalid named port service forward", fields{"int-test-svc-invalid-named", config.Forward{
			Service:   "int-test-svc",
			Namespace: func() *string { s := "k4wd"; return &s }(),
			Remote:    "mysql",
		}}, args{kc, ""}, true},

		{"invalid named target port service forward", fields{"int-test-svc-invalid-named-target", config.Forward{
			Service:   "int-test-svc-invalid-target",
			Namespace: func() *string { s := "k4wd"; return &s }(),
			Remote:    "http-alt",
		}}, args{kc, ""}, true},

		{"unknown service target", fields{"int-test-svc-unknown", config.Forward{
			Service:   "int-test-svc-unknown",
			Namespace: func() *string { s := "k4wd"; return &s }(),
			Remote:    "http-alt",
		}}, args{kc, ""}, true},

		{"zero pods service target", fields{"int-test-svc-zero", config.Forward{
			Service:   "int-test-svc-zero",
			Namespace: func() *string { s := "k4wd"; return &s }(),
			Remote:    "http-alt",
		}}, args{kc, ""}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fwd, err := New(tt.fields.Name, tt.fields.Forward)
			if err != nil {
				t.Fatal(err)
			}
			val, err := checkForward(fwd, tt.args.kc)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if val != tt.args.k4wdType {
				t.Errorf("Run() = %v, want %v", val, tt.args.k4wdType)
			}
		})
	}
}
