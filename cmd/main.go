package main

import (
	"fmt"
	"os"

	awair "github.com/azide0x37/awairpoller/pkg"

	"github.com/kris-nova/logger"
	"k8s.io/client-go/util/homedir"
)

func main() {
	logger.BitwiseLevel = logger.LogEverything
	logger.Always("AwairPoller")
	logger.Always("Precheck: Uninstalling previous AwairPoller if it exists.")

	y := awair.New()
	y.ContainerImage = "thenetworkchuck/nccoffee:vacpot"
	y.ContainerPort = 80
	err := y.KubernetesClient(fmt.Sprintf("%s/.kube/kubeconfig.yaml", homedir.HomeDir()))
	if err != nil {
		// Oh no!
		logger.Critical("Unable to load the Kubernetes client")
		logger.Critical("Oof: %v", err)
		os.Exit(1) // <--- Kill the program
	}
	logger.Success("Success! Created Kubernetes Client!")
	err = y.UninstallKubernetes()
	if err != nil {
		// Oh no!
		logger.Warning("Looks like UninstallKubernetes() failed: %v", err)
	}
	logger.Success("Success! Cleaned up any existing deployments!")
	err = y.InstallKubernetes()
	if err != nil {
		// Oh no!
		logger.Critical("Unable to install in Kubernetes!")
		logger.Critical("Something went wrong: %v", err)
	}
	logger.Success("Success! Installed AwairPoller!")

}
