package awairpoller

import (
	"context"
	"fmt"
	"strings"

	v1alpha3 "istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"

	versionedclient "istio.io/client-go/pkg/clientset/versioned"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// AwairPoller is our very important business application
type AwairPoller struct {
	ContainerPort  int32
	ContainerImage string
	client         *kubernetes.Clientset
	istioclient    *versionedclient.Clientset
	deployment     *v1.Deployment
	service        *apiv1.Service
	virtualservice *networkingv1alpha3.VirtualService
	gateway        *networkingv1alpha3.Gateway
}

// New returns an empty *AwairPoller{}
func New() *AwairPoller {
	return &AwairPoller{}
}

// InstallKubernetes will try to install AwairPoller in Kubernetes
func (y *AwairPoller) InstallKubernetes() error {
	// Notice how we can do amazing things like add custom error messages
	if y.client == nil {
		return fmt.Errorf("missing kube client: use KubernetesClient()")
	}

	deployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "awair-poller",
			Labels: map[string]string{
				"app":     "static-site",
				"version": "v1",
			},
		},
		Spec: v1.DeploymentSpec{
			Replicas: int32Ptr(3),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "static-site",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":     "static-site",
						"version": "v1",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            "awair-poller",
							Image:           y.ContainerImage,
							ImagePullPolicy: "Always",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: y.ContainerPort,
								},
							},
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									"cpu":    resource.MustParse("64Mi"),
									"memory": resource.MustParse("25m"),
								},
								Limits: apiv1.ResourceList{
									"cpu":    resource.MustParse("128Mi"),
									"memory": resource.MustParse("50m"),
								},
							},
						},
					},
				},
			},
		},
	}

	service := &apiv1.Service{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: "static-site-service",
			Labels: map[string]string{
				"app": "static-site",
			},
			Annotations: map[string]string{
				"service.beta.kubernetes.io/linode-loadbalancer-throttle": "4",
			},
		},
		Spec:   apiv1.ServiceSpec{},
		Status: apiv1.ServiceStatus{},
	}

	gateway := &networkingv1alpha3.Gateway{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: v1alpha3.Gateway{
			Servers: []*v1alpha3.Server{
				{
					Port:  &v1alpha3.Port{Number: 80},
					Hosts: []string{"*"},
					Name:  "HTTP",
				},
			},
			Selector: map[string]string{
				"istio": "ingressgateway",
			},
		},
	}

	virtualservice := &networkingv1alpha3.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name: "static-site",
		},
		Spec: v1alpha3.VirtualService{
			Hosts:    []string{"*"},
			Gateways: []string{"static-site-gateway"},
			Http: []*v1alpha3.HTTPRoute{
				{
					Route: []*v1alpha3.HTTPRouteDestination{
						{
							Destination: &v1alpha3.Destination{
								Host: "static-site-service",
								//Subset:"",
							},
							Weight: 100,
						},
					},
				},
			},
		},
	}

	y.deployment = deployment
	err := y.Validate()
	if err != nil {
		// See how cool errors are? Especially when the errors are things that keep you from
		// getting paged.
		return fmt.Errorf("invalid deployment: %v", err)
	}
	// We could even make a copy of this and send it to MongoDB you know - just for record keeping.
	//
	// But why would you do that?
	//
	// Oh you know SO THAT WE HAVE A RECORD OF EVERYTHING OUR TEAM DOES SO WE CAN QUERY THE DATA LATER IF WE WANT
	//
	// The point is that you can build anything here because you are writing Go.
	//
	// You can introduce all kinds of cool features that make your team look awesome.
	err = y.Archive()
	if err != nil {
		return fmt.Errorf("unable to archive in mongodb: %s", err)
	}

	// And here we go folks
	//
	// The deployment is crafted.
	// The client is authenticated.
	// The validation checks are passed.
	// We can finally install in Kubernetes.

	//NOTE: Execute our Deployment
	deployResult, err := y.client.AppsV1().Deployments("default").Create(context.TODO(), deployment, metav1.CreateOptions{})

	if err != nil {
		return fmt.Errorf("oh no! something went wrong deploying to kubernetes: %v", err)
	}

	//NOTE: Execute our Service
	serviceResult, err := y.client.CoreV1().Services("default").Create(context.TODO(), service, metav1.CreateOptions{})

	if err != nil {
		return fmt.Errorf("oh no! something went wrong deploying DestinationRule to kubernetes: %v", err)
	}

	//NOTE: Execute our VirtualService
	virtualServiceResult, err := y.istioclient.NetworkingV1alpha3().VirtualServices("default").Create(context.TODO(), virtualservice, metav1.CreateOptions{})

	if err != nil {
		return fmt.Errorf("oh no! something went wrong deploying Gateway to kubernetes: %v", err)
	}

	//NOTE: Execute our Gateway
	gatewayResult, err := y.istioclient.NetworkingV1alpha3().Gateways("default").Create(context.TODO(), gateway, metav1.CreateOptions{})

	if err != nil {
		return fmt.Errorf("oh no! something went wrong deploying VirtualService to kubernetes: %v", err)
	}

	// Update the deployment in memory with the object from Kubernetes
	y.deployment = deployResult
	y.service = serviceResult
	y.virtualservice = virtualServiceResult
	y.gateway = gatewayResult

	return nil
}

// KubernetesClient will try to configure client for Kubernetes
func (y *AwairPoller) KubernetesClient(kubeconfigPath string) error {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return fmt.Errorf("unable to authenticate with Kubernetes with kube config %s: %v", kubeconfigPath, err)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("unable to authenticate with Kubernetes with kube config %s: %v", kubeconfigPath, err)
	}
	y.client = client
	return nil
}

// KubernetesClient will try to configure client for Kubernetes
func (y *AwairPoller) IstioClient(kubeconfigPath string) error {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return fmt.Errorf("unable to authenticate with Kubernetes with kube config %s: %v", kubeconfigPath, err)
	}
	ic, err := versionedclient.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("Failed to create istio client: %s", err)
	}

	y.istioclient = ic
	return nil
}

// Archive will try to back up this config
func (y *AwairPoller) Archive() error {
	// See you can even build out scaffolding for other features
	// you hope to build later. Maybe we will get to this next sprint.
	//
	// Because you are now a software engineer you do things like this.
	//
	// TODO @kris-nova save to mongo
	if y.deployment == nil {
		return fmt.Errorf("unable validate AwairPoller deployment: missing deployment")
	}
	return nil
}

// Validate is all the handsome checks you get to design and talk about as a team.
// I wonder what would be important for you and your org to check for here?
// Hrmm...
func (y *AwairPoller) Validate() error {
	// Let's just make sure we have one...
	if y.deployment == nil {
		return fmt.Errorf("unable validate AwairPoller deployment: missing deployment")
	}
	// It wouldn't make much sense to try to deploy a deployment without any containers
	// Let's make sure we have at least 1 defined
	if len(y.deployment.Spec.Template.Spec.Containers) < 1 {
		return fmt.Errorf("unable to validate AwairPoller deployment. less than 1 container")
	}

	// So basically can do anything we want here because we have an entire programming language
	// at our disposal.
	//
	// Oh but you could use OPA and admissions controllers for things like this...
	//
	// Or we could just keep everything in one place and build a library of validation tools
	// like we would build a library of unit tests.
	//
	// You know... Fail quick.. Fail fast.. thats... a thing... right?
	//
	// Anyway I just dreamt up something we want to check for. Check that the name contains
	// the string "yam". Why? I don't know. Just seemed like a good example.
	//
	if !strings.Contains(y.deployment.Name, "yam") {
		return fmt.Errorf("unable validate AwairPoller deployment. invalid name %s", y.deployment.Name)
	}
	return nil
}

// UninstallKubernetes is a good method that is implemented poorly (on purpose).
// Feel free to come clean this up.
func (y *AwairPoller) UninstallKubernetes() error {
	if y.client == nil {
		return fmt.Errorf("missing kube client: use KubernetesClient()")
	}
	// This is actually really bad practice.
	// Here we "hard code" both the name of the namespace, and the name of the deployment to delete.
	// But you know what? A TODO is still 100% better than YAML files interpolated at runtime.
	// TODO @kris-nova introduce dynamic namespaces and names
	return y.client.AppsV1().Deployments("default").Delete(context.TODO(), "awair-poller", metav1.DeleteOptions{})

}

func int32Ptr(i int32) *int32 {
	return &i
}
