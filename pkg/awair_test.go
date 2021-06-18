package awairpoller

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestAwairPoller_InstallKubernetes(t *testing.T) {
	y := New()
	y.client = nil
	err := y.InstallKubernetes()
	if err == nil {
		t.Errorf("Expecting error for nil client")
	}
}
