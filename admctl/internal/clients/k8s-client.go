package clients

import (
	"context"
	"fmt"
	"log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type K8sClient struct {
	client *kubernetes.Clientset
}

func NewK8Client() (*K8sClient, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("Error: %v", err.Error())
		return nil, err
	}
	// creates the clientset
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("Error: %v", err.Error())
		return nil, err
	}
	return &K8sClient{client}, nil
}

func (client *K8sClient) GetDeployments() {
	deployList, err := client.client.AppsV1().Deployments("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, deployment := range deployList.Items {
		log.Printf("Deployment: %v", deployment.Name)
	}
}

func (client *K8sClient) FindServiceForDeployment(deployment *appsv1.Deployment) (*corev1.Service, error) {
	// Get all services in the deployment's namespace
	services, err := client.client.CoreV1().Services(deployment.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %v", err)
	}

	// Look for matching service
	for _, service := range services.Items {
		// Check if service selector matches deployment labels
		matches := true
		for key, value := range service.Spec.Selector {
			if deploymentValue, ok := deployment.Spec.Template.Labels[key]; !ok || deploymentValue != value {
				matches = false
				break
			}
		}

		if matches {
			return &service, nil
		}
	}

	return nil, fmt.Errorf("no matching service found for deployment %s/%s", deployment.Namespace, deployment.Name)
}
