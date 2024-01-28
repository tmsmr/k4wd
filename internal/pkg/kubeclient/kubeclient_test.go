package kubeclient

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
	"os"
	"path"
	"reflect"
	"testing"
)

const kubeconfig = `
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUI5RENDQVYyZ0F3SUJBZ0lSQU5kM281bEVJWXNyeW9SNzRydStMU013RFFZSktvWklodmNOQVFFTEJRQXcKRmpFVU1CSUdBMVVFQXhNTFpYaGhiWEJzWlM1amIyMHdIaGNOTWpRd01USXhNRFkwTnpFNVdoY05NelF3TVRFNApNRFkwTnpFNVdqQVdNUlF3RWdZRFZRUURFd3RsZUdGdGNHeGxMbU52YlRDQm56QU5CZ2txaGtpRzl3MEJBUUVGCkFBT0JqUUF3Z1lrQ2dZRUFuczlaWnExa0E1eUtOd09pN0VGVHJKWDlmeUZWcDlwQmNkSDJkaXR5cGpmNmFoNkwKRFcrZTl3M0lRZ1FQMnFqQnVBV0srRnd1eHVFVndJK0xKL21mMVkvQzdTc0V2bjZnZEZmMWJ5R2h2MW55Q3NNbgpzaXlxLzBzNGdyQUxUK3J6VnRNUlFKdUphQUprdktqdnlTdTNQR0wwNExrSDM3U3FtVVVMU0MxbVgra0NBd0VBCkFhTkNNRUF3RGdZRFZSMFBBUUgvQkFRREFnSUVNQThHQTFVZEV3RUIvd1FGTUFNQkFmOHdIUVlEVlIwT0JCWUUKRlBuZTJjWlVnSlB1blMwNHp0L25pREc4UFpiS01BMEdDU3FHU0liM0RRRUJDd1VBQTRHQkFDSStPQ1ljYy81SQpaM0szaG9janJyTWlrR3V3QnVQNzEva3hVb1BBTTNzaTA2TWlWTEdGTGhmQ2tjTkdXUTFLaUo1dHZpNEs2QStWCi9DZjlYTkZ1bFhVVDRCdVZJbzdQWkFQS2JabFZsVTdBUGo3MjVYYXB6VGhzenNMU2JpOUpRRjNvSjJiTmkxZ0wKaHdkQWNSSnZMUUo2Q0Vlb2piQXVuSXdiOHVtVGVPbk8KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    server: https://127.0.0.1:6443
  name: local
contexts:
- context:
    cluster: local
    user: user
  name: local
current-context: local
kind: Config
preferences: {}
users:
- name: user
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNBRENDQVdtZ0F3SUJBZ0lRRWR0bGJlUWxYSzQvTzM1dmV6WG9zREFOQmdrcWhraUc5dzBCQVFzRkFEQVcKTVJRd0VnWURWUVFERXd0bGVHRnRjR3hsTG1OdmJUQWVGdzB5TkRBeE1qRXdOalEzTVRsYUZ3MHlOVEF4TWpBdwpOalEzTVRsYU1COHhIVEFiQmdOVkJBTVRGR05zYVdWdWRDMWhMbVY0WVcxd2JHVXVZMjl0TUlHZk1BMEdDU3FHClNJYjNEUUVCQVFVQUE0R05BRENCaVFLQmdRREFIQUNTVUFmT1I3c21pcFV1cTRiOG5vZlZ1SENJOG15MEZ5eUUKelRIUmEzL0poYmtpbXdZZzZZSkNWb1ordGo5MWdiYUZsSkJKSm5IN0t4ZlVTWDIwei9ENUM4ME1mcVQrUDhadgozbDk3ZjJvc1RyRmkwd0gvVlBELzc0VzJJamoycHVIMmZTNEYvclMxcFhtdHdQVSt0K1BzWEZvN3NVRS8xS1dSCldaSGVDd0lEQVFBQm8wWXdSREFUQmdOVkhTVUVEREFLQmdnckJnRUZCUWNEQWpBTUJnTlZIUk1CQWY4RUFqQUEKTUI4R0ExVWRJd1FZTUJhQUZQbmUyY1pVZ0pQdW5TMDR6dC9uaURHOFBaYktNQTBHQ1NxR1NJYjNEUUVCQ3dVQQpBNEdCQUI2UG4rQ3V5OXdoRGRPU3FHSFN4R09HTDNwUEFwMC9jUVdRYjVsa1pIR09XSUp6azM4YnJDWUtCdUdFCjNpL0JPVVpRaCtXS3pDSGdFZ0NpY0FIZUZKL3RxRitDcGx0QVNDT1U2OUg0R1p2NE9vanRRLysyK0xZb2NZM08KODFpQ0V4eDVybTlKYUF6ZVNRSnFpRCtiZGtNWlE5eGMyUEVTTm5kd0lrRHI1UXNlCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlDWEFJQkFBS0JnUURBSEFDU1VBZk9SN3NtaXBVdXE0Yjhub2ZWdUhDSThteTBGeXlFelRIUmEzL0poYmtpCm13WWc2WUpDVm9aK3RqOTFnYmFGbEpCSkpuSDdLeGZVU1gyMHovRDVDODBNZnFUK1A4WnYzbDk3ZjJvc1RyRmkKMHdIL1ZQRC83NFcySWpqMnB1SDJmUzRGL3JTMXBYbXR3UFUrdCtQc1hGbzdzVUUvMUtXUldaSGVDd0lEQVFBQgpBb0dBS2l0N2JLS29zLzFHOWJUaC9uYWtrUHR6c2RSa3o0cjNsYWlvbXhZMzQxS0hvSUw4R3I2UTM5U2lSVkdkCkFGK2RHbnc0eHFYUDdsN0VFbkJwTUs1WksxL1BZOUFIYmh5eWdiWUhGeTJaaTVMc2VqVUViek1iQVdqd2hHeDEKL0RCNTFZTWNXSDlVdUhSUy9qczhyT1FwRTN6SFNGb2RRM2JoZEdObUxIYWNtL0VDUVFESk1oRlNhS1EyelNZcwp0MncvYnM5K3l1SzdZU3E3WFd4WmlHbHo3L3pKR1l3YkhnVmtoTCtVa21HVGFYeURvWDZzK3Vqd0x5NkoydVFwCndQNnFBZmlQQWtFQTlIQlVMWUF6WkRLT1c1TytkUGRwd3NQL0R3TnBaSlhMaVNwYUdyeWN5KytuaDlwZTNMU3EKOEMwc2EwV0tLaHZxWEtXb0crSTBzU2NwdUNoeUpIVG94UUpBWGU1dklkOVMwM2NCM0p5aUFCZDI3a1pBaHFUOQpzMDRSbU5kVURGbTkxaEdFVk9DMk9kQzBOT1FHaERFYWZjWDNBMEY5WVMxVjkreG0yNHVNR2Nranh3SkJBSVVVCklzQWk2OWZCTG4vdEQrUGVUMVlhSHVLdG1OT0tPaVdUU1RzRk5OaFN3WUxWQUpCb1RDZzJiOWgzSTZHSlVTN2YKZ1lhc3dNTXg3eVN6NEhDNHRZRUNRRVk5VzhVQTk2Y0tkM3d2SE5mNGZNRlVSeWI2Z3YrZFRCUCtyT2ZTN25adApqTy9TSVZtb2RXNytlS2Z1NnR1OTFSWmc0ME0vTzJTL3hrTW9qSHRFM3BFPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=
`

func TestNew(t *testing.T) {
	dir, err := os.MkdirTemp(os.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path.Join(dir, "config"), []byte(kubeconfig), 0644); err != nil {
		t.Fatal(err)
	}
	type args struct {
		opts []ClientOption
	}
	tests := []struct {
		name    string
		args    args
		want    *Kubeclient
		wantErr bool
	}{
		{"valid", args{[]ClientOption{WithKubeconfig(path.Join(dir, "config"))}}, &Kubeclient{Kubeconfig: path.Join(dir, "config"), APIConfig: &api.Config{}}, false},
		{"nonexistent kubeconfig", args{[]ClientOption{WithKubeconfig(path.Join(dir, "nonexistent"))}}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("New() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func mockUserKubeconfig(envkey string, unsetkey string) error {
	base, err := os.MkdirTemp(os.TempDir(), "")
	if err != nil {
		return err
	}
	if err := os.Setenv(envkey, base); err != nil {
		return err
	}
	if err := os.Unsetenv(unsetkey); err != nil {
		return err
	}
	err = os.Mkdir(path.Join(base, ".kube"), 0755)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path.Join(base, ".kube", "config"), []byte(kubeconfig), 0644); err != nil {
		return err
	}
	return nil
}

func TestNewDefaultLocationNix(t *testing.T) {
	if err := mockUserKubeconfig("HOME", "USERPROFILE"); err != nil {
		t.Fatal(err)
	}
	_, err := New()
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewDefaultLocationWin(t *testing.T) {
	if err := mockUserKubeconfig("USERPROFILE", "HOME"); err != nil {
		t.Fatal(err)
	}
	_, err := New()
	if err != nil {
		t.Fatal(err)
	}
}

func TestKubeclient_Clientset_RESTConfig(t *testing.T) {
	dir, err := os.MkdirTemp(os.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path.Join(dir, "config"), []byte(kubeconfig), 0644); err != nil {
		t.Fatal(err)
	}
	kc, err := New(WithKubeconfig(path.Join(dir, "config")))
	if err != nil {
		t.Fatal(err)
	}
	type fields struct {
		Kubeconfig  string
		Kubecontext string
		APIConfig   *api.Config
	}
	type args struct {
		context *string
		config  *rest.Config
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *kubernetes.Clientset
		wantErr bool
	}{
		{"valid", fields{APIConfig: kc.APIConfig}, args{nil, nil}, &kubernetes.Clientset{}, false},
		{"with context", fields{APIConfig: kc.APIConfig}, args{context: func() *string { s := "local"; return &s }()}, &kubernetes.Clientset{}, false},
		{"invalid context", fields{APIConfig: kc.APIConfig}, args{context: func() *string { s := "invalid"; return &s }()}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kc := Kubeclient{
				Kubeconfig:  tt.fields.Kubeconfig,
				Kubecontext: tt.fields.Kubecontext,
				APIConfig:   tt.fields.APIConfig,
			}
			got, err := kc.Clientset(tt.args.context, tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Clientset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("Clientset() got = %v, want %v", got, tt.want)
			}
		})
	}
}
