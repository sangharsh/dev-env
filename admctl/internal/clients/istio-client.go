package clients

import (
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

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
		log.Printf("DestinationRule: %s in namespace %s", dr.Name, dr.Namespace)
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
		log.Printf("VirtualService: %s in namespace %s\n", vs.Name, vs.Namespace)
		// Access hosts
		for _, host := range vs.Spec.Hosts {
			log.Printf("  Host: %s\n", host)
		}
		// Access gateways
		for _, gateway := range vs.Spec.Gateways {
			log.Printf("  Gateway: %s\n", gateway)
		}
		// Access HTTP routes
		for _, http := range vs.Spec.Http {
			log.Printf("  HTTP Route: %+v\n", http)
		}
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
