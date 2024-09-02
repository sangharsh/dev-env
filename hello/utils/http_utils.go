package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func prepareRequest(inRequest *http.Request, url string) (*http.Request, error) {
	ctx := inRequest.Context()
	outRequest, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	propagator := otel.GetTextMapPropagator()
	ctx = propagator.Extract(ctx, propagation.HeaderCarrier(inRequest.Header))

	propagator.Inject(ctx, propagation.HeaderCarrier(outRequest.Header))
	return outRequest, nil
}

func parseJSONResponse(response *http.Response) (interface{}, error) {
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	var jsonData interface{}
	err = json.Unmarshal(responseBody, &jsonData)
	if err != nil {
		return nil, fmt.Errorf("error parsing response body to JSON: %v", err)
	}

	return jsonData, nil
}

func FetchJSONResponse(inRequest *http.Request, url string) (interface{}, error) {
	outRequest, err := prepareRequest(inRequest, url)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(outRequest)
	if err != nil {
		return nil, fmt.Errorf("error while making HTTP request: %v", err)
	}
	jsonData, err := parseJSONResponse(response)
	return jsonData, err
}
