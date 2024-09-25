package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	admission "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
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
		admissionResponse = ac.validate(&ar)
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
	}
	log.Printf("Ready to write response ...")
	log.Printf("admissionReview: %v", admissionReview)
	if _, err := w.Write(resp); err != nil {
		log.Printf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}

func (ac *admissionController) validate(ar *admission.AdmissionReview) *admission.AdmissionResponse {
	req := ar.Request

	var pod corev1.Pod
	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		log.Printf("Could not unmarshal raw object: %v", err)
		return &admission.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	log.Printf("AdmissionReview for Kind=%v, Namespace=%v Name=%v (%v) UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, pod.Name, req.UID, req.Operation, req.UserInfo)

	if _, ok := pod.Labels[requiredLabel]; !ok {
		return &admission.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: fmt.Sprintf("Pod is missing required label: %s", requiredLabel),
			},
		}
	}

	return &admission.AdmissionResponse{
		Allowed: true,
	}
}

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
