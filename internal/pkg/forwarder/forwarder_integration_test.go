package forwarder

/*
Tests in this file are integration tests, they require some prerequisites to run:
- a running Kubernetes cluster and the kubeconfig file at the default location
- deployed manifests from the https://github.com/tmsmr/k4wd/tree/main/test directory

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

var ff = config.Forwardfile{
	Forwards: map[string]config.Forward{
		"int-test-po": {
			Pod:       "int-test-po",
			Namespace: func() *string { s := "k4wd"; return &s }(),
			Remote:    "http-alt",
		},
		"int-test-de": {
			Deployment: "int-test-de",
			Namespace:  func() *string { s := "k4wd"; return &s }(),
			Remote:     "http-alt",
		},
		"int-test-svc": {
			Service:   "int-test-svc",
			Namespace: func() *string { s := "k4wd"; return &s }(),
			Remote:    "http-alt",
		},
	},
}

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
	failed := make(chan error, 1)
	defer close(failed)
	go func() {
		err := fwd.Run(kc, stop)
		if err != nil {
			failed <- err
		}
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
	case err := <-failed:
		return "", err
	}
}

func TestForwarder_Run(t *testing.T) {
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
		{"valid pod tagret", fields{"int-test-po", ff.Forwards["int-test-po"]}, args{kc, "pod"}, false},
		{"valid deployment target", fields{"int-test-de", ff.Forwards["int-test-de"]}, args{kc, "deployment"}, false},
		{"valid service target", fields{"int-test-svc", ff.Forwards["int-test-svc"]}, args{kc, "deployment"}, false},
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
			if val != tt.args.k4wdType {
				t.Errorf("Run() = %v, want %v", val, tt.args.k4wdType)
			}
		})
	}
}
