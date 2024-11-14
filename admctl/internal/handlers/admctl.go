package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/sangharsh/dev-env/admctl/internal/clients"
	"github.com/sangharsh/dev-env/admctl/pkg/api"
	admission "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AdmissionController struct {
}

func (ac *AdmissionController) Serve(w http.ResponseWriter, r *http.Request) {
	parser := api.NewParser()
	responder := api.NewResponder()
	ar, err := parser.ParseRequest(r)
	if err != nil {
		responder.Error(w, ar, err)
		return
	}

	if *ar.Request.DryRun {
		admissionResponse := &admission.AdmissionResponse{
			Allowed: true,
		}
		responder.PrepareAndSendResponse(w, ar, admissionResponse)
		return
	}

	admissionResponse := ac.handle(ar)
	responder.PrepareAndSendResponse(w, ar, admissionResponse)
}

func (ac *AdmissionController) handle(ar *admission.AdmissionReview) *admission.AdmissionResponse {
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

	istioClient, err := clients.NewIstioClient()
	// getDeployments()
	err = istioClient.GetVirtualServices("default")
	if err != nil {
		log.Printf("error get VS: %v", err)
	}
	err = istioClient.GetDestinationRules("default")
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