package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"

	apiv1alpha3 "istio.io/api/networking/v1alpha3"
	istionetworking "istio.io/client-go/pkg/apis/networking/v1beta1"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
)

type IstioClient struct {
	client *istioclient.Clientset
}

func NewIstioClient() (*IstioClient, error) {

	// Get Istio client config from the admission controller's config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster config: %v", err)
	}

	// Create Istio client
	client, err := istioclient.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create istio client: %v", err)
	}
	return &IstioClient{client}, nil
}

func (client *IstioClient) GetDestinationRules(namespace string) error {
	destinationRules, err := client.client.NetworkingV1().DestinationRules(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list destination rules: %v", err)
	}

	for _, dr := range destinationRules.Items {
		log.Printf("DestinationRule: %v", dr)
	}
	return nil
}
func (client *IstioClient) GetVirtualServices(namespace string) error {
	// List VirtualServices in the specified namespace
	// Use "" for namespace to list across all namespaces
	virtualServices, err := client.client.NetworkingV1beta1().VirtualServices(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list virtual services: %v", err)
	}

	// Process the virtual services
	for _, vs := range virtualServices.Items {
		// Access virtual service properties
		log.Printf("VirtualService: %v", vs)
	}

	return nil
}

func (client *IstioClient) FindDestinationRuleForService(
	service *corev1.Service,
) (*istionetworking.DestinationRule, error) {
	// Get all destination rules in the service's namespace
	destRules, err := client.client.NetworkingV1beta1().DestinationRules(service.Namespace).List(
		context.TODO(),
		metav1.ListOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list destination rules: %v", err)
	}

	possibleHosts := generatePossibleHosts(service)

	// Look for matching destination rule
	for i := range destRules.Items {
		dr := destRules.Items[i]
		for _, host := range possibleHosts {
			if dr.Spec.Host == host {
				return dr, nil
			}
		}
	}

	return nil, fmt.Errorf("no destination rule found for service %s/%s", service.Namespace, service.Name)
}

func (client *IstioClient) FindVirtualServiceForService(
	service *corev1.Service,
) (*istionetworking.VirtualService, error) {
	// Get all virtual services in the service's namespace
	virtualServices, err := client.client.NetworkingV1beta1().VirtualServices(service.Namespace).List(
		context.TODO(),
		metav1.ListOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list virtual services: %v", err)
	}

	possibleHosts := generatePossibleHosts(service)

	// Look for matching virtual service
	for i := range virtualServices.Items {
		vs := virtualServices.Items[i]
		for _, vsHost := range vs.Spec.Hosts {
			for _, host := range possibleHosts {
				if vsHost == host {
					return vs, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no virtual service found for service %s/%s", service.Namespace, service.Name)
}

// generatePossibleHosts generates all possible host variations for a service
func generatePossibleHosts(service *corev1.Service) []string {
	hosts := []string{
		service.Name, // Short name
		fmt.Sprintf("%s.%s", service.Name, service.Namespace),                   // name.namespace
		fmt.Sprintf("%s.%s.svc", service.Name, service.Namespace),               // name.namespace.svc
		fmt.Sprintf("%s.%s.svc.cluster.local", service.Name, service.Namespace), // FQDN
	}
	return hosts
}

func (c *IstioClient) AddSubsetToDestinationRule(
	dr *istionetworking.DestinationRule,
	version string,
) (*istionetworking.DestinationRule, error) {
	// Create the new subset
	newSubset := &apiv1alpha3.Subset{
		Name: version,
		Labels: map[string]string{
			"version": version,
		},
	}

	patchObj := []map[string]interface{}{
		{
			"op":    "add",
			"path":  "/spec/subsets/0",
			"value": newSubset,
		},
	}

	// Convert patch to JSON
	patchBytes, err := json.Marshal(patchObj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal patch: %v", err)
	}

	// Apply the patch
	updatedDR, err := c.client.NetworkingV1beta1().DestinationRules(dr.Namespace).Patch(
		context.TODO(),
		dr.Name,
		types.JSONPatchType,
		patchBytes,
		metav1.PatchOptions{},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to patch destination rule: %v", err)
	}
	log.Printf("updated dr: %v", updatedDR)
	return updatedDR, nil
}
