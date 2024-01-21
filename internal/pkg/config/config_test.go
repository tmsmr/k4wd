package config

import (
	"os"
	"path"
	"reflect"
	"testing"
)

func TestForwardfile_Validate(t *testing.T) {
	type fields struct {
		Forwards map[string]Forward
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"valid", fields{Forwards: map[string]Forward{"test": {Pod: "test", Remote: "http"}}}, false},
		{"no forwards", fields{Forwards: nil}, true},
		{"empty forwards", fields{Forwards: map[string]Forward{}}, true},
		{"invalid forward", fields{Forwards: map[string]Forward{"test": {Pod: "test"}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ff := &Forwardfile{
				Forwards: tt.fields.Forwards,
			}
			if err := ff.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	dir, err := os.MkdirTemp(os.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path.Join(dir, "Forwardfile"), []byte(`
[forwards.test]
pod = "test"
remote = "http"
`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path.Join(dir, "Forwardfile-invalid"), []byte("relaxed = true"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path.Join(dir, "Forwardfile-invalid-fmt"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	type args struct {
		opts []ForwardfileOption
	}
	tests := []struct {
		name    string
		args    args
		want    *Forwardfile
		wantErr bool
	}{
		{"valid", args{}, &Forwardfile{Path: "Forwardfile", Relaxed: false, Forwards: map[string]Forward{"test": {Pod: "test", Remote: "http"}}}, false},
		{"invalid", args{[]ForwardfileOption{WithPath("Forwardfile-invalid")}}, nil, true},
		{"invalid fmt", args{[]ForwardfileOption{WithPath("Forwardfile-invalid-fmt")}}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Load(tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() got = %v, want %v", got, tt.want)
			}
		})
	}
}
