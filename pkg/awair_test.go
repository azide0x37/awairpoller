package awairpoller

import "testing"

func TestAwairPoller_Validate(t *testing.T) {
	y := New()
	y.deployment = nil
	err := y.Validate()
	if err == nil {
		t.Errorf("Expecting error for nil deployment")
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
