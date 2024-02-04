package envfile

import (
	"github.com/tmsmr/k4wd/internal/pkg/forwarder"
	"os"
	"path"
	"regexp"
	"testing"
)

func TestEnvfile_New_Update_Remove(t *testing.T) {
	ef, err := New("Forwardfile")
	if err != nil {
		t.Errorf("New() error = %v", err)
	}
	if ef.Path() == "" {
		t.Errorf("New() ef.Path() is empty")
	}
	fwdsmock := map[string]*forwarder.Forwarder{"test": {Name: "test", BindAddr: "test", BindPort: 8080}}
	if err := ef.Update(fwdsmock); err != nil {
		t.Errorf("Update() error = %v", err)
	}
	exists, err := ef.exists()
	if err != nil {
		t.Errorf("exists() error = %v", err)
	}
	if !exists {
		t.Errorf("envfile does not exist when it should")
	}
	if err := ef.Remove(); err != nil {
		t.Errorf("Remove() error = %v", err)
	}
	exists, err = ef.exists()
	if err != nil {
		t.Errorf("exists() error = %v", err)
	}
	if exists {
		t.Errorf("envfile does exist when it should not")
	}
}

func TestEnvfile_Load(t *testing.T) {
	dir, err := os.MkdirTemp(os.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}
	invalid := path.Join(dir, "invalid")
	if err := os.WriteFile(invalid, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	type fields struct {
		path string
	}
	type args struct {
		f EnvFormat
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"unavailable", fields{"doesnotexists"}, args{FormatDefault}, true},
		{"invalid", fields{invalid}, args{FormatDefault}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ef := &Envfile{
				path: tt.fields.path,
			}
			_, err := ef.Load(tt.args.f)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestEnvfile_Formats(t *testing.T) {
	ef, err := New("Forwardfile")
	if err != nil {
		t.Errorf("New() error = %v", err)
	}
	fwdsmock := map[string]*forwarder.Forwarder{"test": {Name: "test", BindAddr: "test", BindPort: 8080}}
	if err := ef.Update(fwdsmock); err != nil {
		t.Errorf("Update() error = %v", err)
	}
	defer func() {
		if err := ef.Remove(); err != nil {
			t.Errorf("Remove() error = %v", err)
		}
	}()
	whitespace := regexp.MustCompile(`\s`)
	type args struct {
		f EnvFormat
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"FormatJSON", args{FormatJSON}, []byte(`[{"addr": "TEST_ADDR", "value": "test:8080"}]`), false},
		{"FormatDefault", args{FormatDefault}, []byte(`export TEST_ADDR=test:8080`), false},
		{"FormatNoExport", args{FormatNoExport}, []byte(`TEST_ADDR=test:8080`), false},
		{"FormatPS", args{FormatPS}, []byte(`$Env:TEST_ADDR="test:8080"`), false},
		{"FormatCmd", args{FormatCmd}, []byte(`set TEST_ADDR=test:8080`), false},
		{"invalid format", args{-1}, []byte{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ef.Load(tt.args.f)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if whitespace.ReplaceAllString(string(got), "") != whitespace.ReplaceAllString(string(tt.want), "") {
				t.Errorf("Load() got = %s, want %s", whitespace.ReplaceAllString(string(got), ""), whitespace.ReplaceAllString(string(tt.want), ""))
			}
		})
	}
}
