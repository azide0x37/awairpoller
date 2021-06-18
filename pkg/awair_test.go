package awairpoller

import (
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func TestAwairPoller_Validate_Nil_Deployment(t *testing.T) {
	y := New()
	y.deployment = nil
	err := y.Validate()
	if err == nil {
		t.Errorf("Expecting error for nil deployment")
	}
}

func TestAwairPoller_Validate_Zero_Containers(t *testing.T) {
	y := New()
	y.deployment = &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "awair-poller",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"beeps": "boops",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"beeps": "boops",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{},
				},
			},
		},
	}
	err := y.Validate()
	if err == nil {
		t.Errorf("Expecting error for no containers")
	}
}

func TestAwairPoller_InstallKubernetesNilClient(t *testing.T) {
	y := New()
	y.client = nil
	err := y.InstallKubernetes()
	if err == nil {
		t.Errorf("Expecting error for nil client")
	}
}

func TestAwairPoller_InstallKubernetesNilDeployment(t *testing.T) {
	y := New()
	y.client = nil
	err := y.InstallKubernetes()
	if err == nil {
		t.Errorf("Expecting error for nil deployment")
	}
}

func TestAwairPoller_UninstallKubernetesNilClient(t *testing.T) {
	y := New()
	y.client = nil
	err := y.UninstallKubernetes()
	if err == nil {
		t.Errorf("Expecting error for nil client")
	}
}

func TestAwairPoller_ArchiveNilClient(t *testing.T) {
	y := New()
	y.client = nil
	err := y.Archive()
	if err == nil {
		t.Errorf("Expecting error for nil client")
	}
}

func TestAwairPoller_ArchiveNilDeployment(t *testing.T) {
	y := New()
	y.deployment = nil
	err := y.Archive()
	if err == nil {
		t.Errorf("Expecting error for nil deployment")
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want *AwairPoller
	}{
		{
			want: &AwairPoller{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAwairPoller_KubernetesClient(t *testing.T) {
	type fields struct {
		ContainerPort  int32
		ContainerImage string
		client         *kubernetes.Clientset
		deployment     *v1.Deployment
	}
	type args struct {
		kubeconfigPath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			args:    args{kubeconfigPath: "./kubeconfig"},
			wantErr: true,
		},
		{
			args:    args{kubeconfigPath: ""},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			y := &AwairPoller{
				ContainerPort:  tt.fields.ContainerPort,
				ContainerImage: tt.fields.ContainerImage,
				client:         tt.fields.client,
				deployment:     tt.fields.deployment,
			}
			if err := y.KubernetesClient(tt.args.kubeconfigPath); (err != nil) != tt.wantErr {
				t.Errorf("AwairPoller.KubernetesClient() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAwairPoller_UninstallKubernetes(t *testing.T) {
	type fields struct {
		ContainerPort  int32
		ContainerImage string
		client         *kubernetes.Clientset
		deployment     *v1.Deployment
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			y := &AwairPoller{
				ContainerPort:  tt.fields.ContainerPort,
				ContainerImage: tt.fields.ContainerImage,
				client:         tt.fields.client,
				deployment:     tt.fields.deployment,
			}
			if err := y.UninstallKubernetes(); (err != nil) != tt.wantErr {
				t.Errorf("AwairPoller.UninstallKubernetes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
