package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	admission "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	istioclient "istio.io/client-go/pkg/clientset/versioned"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
	clientset     *kubernetes.Clientset
	istioClient   *istioclient.Clientset
)

const (
	requiredLabel = "app"
	tlsDir        = "/etc/webhook/certs"
	tlsCertFile   = "tls.crt"
	tlsKeyFile    = "tls.key"
)

type admissionController struct{}

func (ac *admissionController) serve(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := io.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.Printf("contentType=%s, expect application/json", contentType)
		return
	}

	var admissionResponse *admission.AdmissionResponse
	ar := admission.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		log.Printf("Can't decode body: %v", err)
		admissionResponse = &admission.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		admissionResponse = ac.handle(&ar)
	}

	admissionReview := admission.AdmissionReview{}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
		admissionReview.APIVersion = ar.APIVersion
		admissionReview.Kind = ar.Kind
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		log.Printf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
		return
	} else {
		log.Printf("Response: admissionReview: %v", admissionReview)
	}
	if _, err := w.Write(resp); err != nil {
		log.Printf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}

func (ac *admissionController) handle(ar *admission.AdmissionReview) *admission.AdmissionResponse {
	req := ar.Request

	log.Printf("AdmissionReview for Kind=%v, Namespace=%v Name=%v UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, req.UID, req.Operation, req.UserInfo)

	switch req.Operation {
	case admission.Create:
		return handleCreate(ar)
	case admission.Delete:
		return handleDelete(ar)
	case admission.Update:
		return handleUpdate(ar)
	}

	return &admission.AdmissionResponse{
		Allowed: true,
	}
}

func handleCreate(ar *admission.AdmissionReview) *admission.AdmissionResponse {
	req := ar.Request
	var deployment appsv1.Deployment
	if err := json.Unmarshal(req.Object.Raw, &deployment); err != nil {
		log.Printf("Could not unmarshal raw object: %v", err)
		return &admission.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	log.Printf("Deployment labels: %v", deployment.Labels)
	deploymentType := deployment.GetLabels()["devenv/type"]
	log.Printf("deploymentType: %v", deploymentType)

	getDeployments()
	err := getVirtualServices("default")
	if err != nil {
		log.Printf("error get VS: %v", err)
	}
	err = getDestinationRules("default")
	if err != nil {
		log.Printf("error get DR: %v", err)
	}
	// if deploymentType == "feature" {
	// 	featureName := deployment.GetLabels()["devenv/feature"]
	// }
	return &admission.AdmissionResponse{
		Allowed: true,
	}
}

func handleDelete(ar *admission.AdmissionReview) *admission.AdmissionResponse {
	req := ar.Request
	var deployment appsv1.Deployment
	if err := json.Unmarshal(req.OldObject.Raw, &deployment); err != nil {
		log.Printf("Could not unmarshal raw object: %v", err)
		return &admission.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	log.Printf("Deployment labels: %v", deployment.Labels)
	return &admission.AdmissionResponse{
		Allowed: true,
	}
}

func handleUpdate(ar *admission.AdmissionReview) *admission.AdmissionResponse {
	req := ar.Request
	log.Printf("req.Kind: %v", req.Kind)
	return &admission.AdmissionResponse{
		Allowed: true,
	}
}

func initK8Client() {
	if clientset != nil {
		return
	}
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
}

func getDeployments() {
	initK8Client()
	deployList, err := clientset.AppsV1().Deployments("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, deployment := range deployList.Items {
		log.Printf("Deployment: %v", deployment.Name)
	}
}

func initIstioClient() error {
	if istioClient != nil {
		return nil
	}

	// Get Istio client config from the admission controller's config
	config, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to get cluster config: %v", err)
	}

	// Create Istio client
	istioClient, err = istioclient.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create istio client: %v", err)
	}
	return nil
}

func getDestinationRules(namespace string) error {
	err := initIstioClient()
	if err != nil {
		return err
	}

	destinationRules, err := istioClient.NetworkingV1().DestinationRules(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list destination rules: %v", err)
	}

	for _, dr := range destinationRules.Items {
		log.Printf("DestinationRule: %s in namespace %s", dr.Name, dr.Namespace)
	}
	return nil
}
func getVirtualServices(namespace string) error {
	// Get Istio client config from the admission controller's config
	err := initIstioClient()
	if err != nil {
		return err
	}

	// List VirtualServices in the specified namespace
	// Use "" for namespace to list across all namespaces
	virtualServices, err := istioClient.NetworkingV1beta1().VirtualServices(namespace).List(context.TODO(), metav1.ListOptions{})
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

// func getPods() {
// 	initK8Client()
// 	// get pods in all the namespaces by omitting namespace
// 	// Or specify namespace to get pods in particular namespace
// 	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

// 	// Examples for error handling:
// 	// - Use helper functions e.g. errors.IsNotFound()
// 	// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
// 	_, err = clientset.CoreV1().Pods("default").Get(context.TODO(), "example-xxxxx", metav1.GetOptions{})
// 	if errors.IsNotFound(err) {
// 		fmt.Printf("Pod example-xxxxx not found in default namespace\n")
// 	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
// 		fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
// 	} else if err != nil {
// 		panic(err.Error())
// 	} else {
// 		fmt.Printf("Found example-xxxxx pod in default namespace\n")
// 	}
// }

func main() {
	ac := &admissionController{}

	certPath := filepath.Join(tlsDir, tlsCertFile)
	keyPath := filepath.Join(tlsDir, tlsKeyFile)

	mux := http.NewServeMux()
	mux.HandleFunc("/validate", ac.serve)

	server := &http.Server{
		Addr:    ":8443",
		Handler: mux,
	}

	log.Printf("Starting server on port 8443")
	log.Printf("Using TLS certificate: %s", certPath)
	log.Printf("Using TLS key: %s", keyPath)

	if err := server.ListenAndServeTLS(certPath, keyPath); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Function to check if a file exists
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func init() {
	// Check if TLS certificate and key files exist
	certPath := filepath.Join(tlsDir, tlsCertFile)
	keyPath := filepath.Join(tlsDir, tlsKeyFile)

	if !fileExists(certPath) {
		log.Fatalf("TLS certificate file not found: %s", certPath)
	}

	if !fileExists(keyPath) {
		log.Fatalf("TLS key file not found: %s", keyPath)
	}
}
