package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	admission "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
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
	} else if *ar.Request.DryRun {
		admissionResponse = &admission.AdmissionResponse{
			Allowed: true,
		}
	} else {
		admissionResponse = ac.handle(&ar)
	}

	admissionReview := prepareResponse(ar, admissionResponse)
	sendResponse(w, admissionReview)
}

func prepareResponse(ar admission.AdmissionReview, admissionResponse *admission.AdmissionResponse) admission.AdmissionReview {
	admissionReview := admission.AdmissionReview{}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
		admissionReview.APIVersion = ar.APIVersion
		admissionReview.Kind = ar.Kind
	}
	return admissionReview
}

func sendResponse(w http.ResponseWriter, admissionReview admission.AdmissionReview) {
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
