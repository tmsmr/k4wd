package config

import "testing"

func TestForward_Type(t *testing.T) {
	type fields struct {
		Pod        string
		Deployment string
		Service    string
	}
	tests := []struct {
		name   string
		fields fields
		want   ForwardType
	}{
		{"pod", fields{Pod: "pod"}, ForwardTypePod},
		{"deployment", fields{Deployment: "deployment"}, ForwardTypeDeployment},
		{"service", fields{Service: "service"}, ForwardTypeService},
		{"invalid", fields{}, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Forward{
				Pod:        tt.fields.Pod,
				Deployment: tt.fields.Deployment,
				Service:    tt.fields.Service,
			}
			if got := f.Type(); got != tt.want {
				t.Errorf("Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestForward_LocalAddr(t *testing.T) {
	type fields struct {
		Local string
	}
	tests := []struct {
		name     string
		fields   fields
		wantAddr string
		wantPort int32
		wantErr  bool
	}{
		{"empty", fields{""}, "", 0, false},
		{"port", fields{"1234"}, "", 1234, false},
		{"addr:port", fields{"0.0.0.0:1234"}, "0.0.0.0", 1234, false},
		{"addr", fields{"0.0.0.0"}, "", 0, true},
		{"invalid", fields{"1234:0.0.0.0:1234"}, "", 0, true},
		{"invalid port", fields{"65536"}, "", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Forward{
				Local: tt.fields.Local,
			}
			addr, port, err := f.LocalAddr()
			if (err != nil) != tt.wantErr {
				t.Errorf("LocalAddr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if addr != tt.wantAddr {
				t.Errorf("LocalAddr() got = %v, want %v", addr, tt.wantAddr)
			}
			if port != tt.wantPort {
				t.Errorf("LocalAddr() got1 = %v, want %v", port, tt.wantPort)
			}
		})
	}
}

func TestForward_Validate(t *testing.T) {
	type fields struct {
		Pod        string
		Deployment string
		Service    string
		Remote     string
		Local      string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"valid", fields{Pod: "pod", Remote: "http", Local: ""}, false},
		{"too many res", fields{Pod: "pod", Deployment: "deployment", Remote: "http", Local: ""}, true},
		{"missing res", fields{Remote: "http", Local: ""}, true},
		{"missing remote", fields{Pod: "pod", Local: ""}, true},
		{"invalid local", fields{Pod: "pod", Remote: "http", Local: "NaN"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Forward{
				Pod:        tt.fields.Pod,
				Deployment: tt.fields.Deployment,
				Service:    tt.fields.Service,
				Remote:     tt.fields.Remote,
				Local:      tt.fields.Local,
			}
			if err := f.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
