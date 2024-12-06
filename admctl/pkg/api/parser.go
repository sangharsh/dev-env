package api

import (
	"errors"
	"io"
	"net/http"

	admission "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

type Parser struct {
	// Could have fields like maxBodySize if needed
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseRequest(r *http.Request) (*admission.AdmissionReview, error) {

	// Verify Content-Type
	if r.Header.Get("Content-Type") != "application/json" {
		return nil, errors.New("content type is not application/json")
	}
	if r.Body == nil {
		return nil, errors.New("request body is nil")
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	ar := admission.AdmissionReview{}
	_, _, err = deserializer.Decode(body, nil, &ar)
	if err != nil {
		return nil, err
	}
	return &ar, nil
}
