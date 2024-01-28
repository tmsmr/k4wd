package forwarder

import "testing"

func Test_randomLocalPort(t *testing.T) {
	_, err := randomLocalPort()
	if err != nil {
		t.Fatalf("randomLocalPort() error = %v", err)
	}
}
