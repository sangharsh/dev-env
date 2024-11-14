package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	admission "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Responder struct {
	// Could have fields like logger if needed
}

func NewResponder() *Responder {
	return &Responder{}
}

func (responder *Responder) Error(w http.ResponseWriter, ar *admission.AdmissionReview, err error) {
	log.Printf("err: %v", err)
	admissionResponse := &admission.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
	responder.PrepareAndSendResponse(w, ar, admissionResponse)
}

func (responder *Responder) PrepareAndSendResponse(w http.ResponseWriter, ar *admission.AdmissionReview, admissionResponse *admission.AdmissionResponse) {
	admissionReview := responder.prepareResponse(ar, admissionResponse)
	responder.sendResponse(w, admissionReview)
}

func (responder *Responder) prepareResponse(ar *admission.AdmissionReview, admissionResponse *admission.AdmissionResponse) admission.AdmissionReview {
	admissionReview := admission.AdmissionReview{}
	if admissionResponse == nil {
		return admissionReview
	}
	admissionReview.Response = admissionResponse
	if ar.Request != nil {
		admissionReview.Response.UID = ar.Request.UID
	}
	admissionReview.APIVersion = ar.APIVersion
	admissionReview.Kind = ar.Kind
	return admissionReview
}

func (responder *Responder) sendResponse(w http.ResponseWriter, admissionReview admission.AdmissionReview) {
	resp, err := json.Marshal(admissionReview)
	if err != nil {
		log.Printf("Can't encode response: %v", err)
		errString := fmt.Sprintf("could not encode response: %v", err)
		http.Error(w, errString, http.StatusInternalServerError)
		return
	}
	_, err = w.Write(resp)

	if err != nil {
		log.Printf("Can't write response: %v", err)
		errString := fmt.Sprintf("could not write response: %v", err)
		http.Error(w, errString, http.StatusInternalServerError)
		return
	}

	log.Printf("Sent response: admissionReview: %v", admissionReview)
}
