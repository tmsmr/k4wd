package envfile

import (
	"github.com/tmsmr/k4wd/internal/pkg/model"
	"os"
	"testing"
)

func TestEnvfileRoundtrip(t *testing.T) {
	forwards := make(map[string]*model.PortForwardSpec)
	forwards["pod-a"] = &model.PortForwardSpec{
		Name:       "pod-a",
		Type:       0,
		Namespace:  "default",
		Pod:        "pod-a",
		Service:    "",
		Remote:     "http",
		Local:      "8080",
		LocalAddr:  "127.0.0.1",
		LocalPort:  8080,
		TargetPort: 80,
		Active:     true,
	}
	forwards["pod-b"] = &model.PortForwardSpec{
		Name:       "pod-b",
		Type:       0,
		Namespace:  "default",
		Pod:        "pod-b",
		Service:    "",
		Remote:     "http",
		Local:      "8081",
		LocalAddr:  "127.0.0.1",
		LocalPort:  8081,
		TargetPort: 80,
		Active:     false,
	}
	forwards["service-a"] = &model.PortForwardSpec{
		Name:       "service-b",
		Type:       1,
		Namespace:  "namespace",
		Pod:        "",
		Service:    "service-b",
		Remote:     "8080",
		Local:      "",
		LocalAddr:  "127.0.0.1",
		LocalPort:  51443,
		TargetPort: 8080,
		Active:     true,
	}
	ef, err := New()
	if err != nil {
		t.Fatal(err)
	}
	exists, err := ef.Exists()
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("env file exists already...")
	}
	err = ef.Update(forwards)
	if err != nil {
		t.Fatal(err)
	}
	defer func(ef *Envfile) {
		err := ef.Remove()
		if err != nil {
			t.Fatal(err)
		}
	}(ef)
	_, err = ef.Load()
	if err != nil {
		t.Fatal(err)
	}
	target, err := os.CreateTemp("", ".env")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(target.Name())
	err = ef.Copy(target.Name())
	if err != nil {
		t.Fatal(err)
	}
}
