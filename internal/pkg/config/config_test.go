package config

import (
	"testing"
)

func TestLoadValidForwardfile(t *testing.T) {
	_, err := Load()
	if err != nil {
		t.Error(err)
	}
}

func TestLoadInvalidForwardfile(t *testing.T) {
	_, err := Load(WithPath("Forwardfile.invalid"))
	if err == nil {
		t.Errorf("no error on invalid Forwardfile")
	}
}
