package clients

import (
	"context"
	"log"

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
