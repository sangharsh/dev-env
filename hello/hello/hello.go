package hello

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

const name = "github.com/sangharsh/dev-env/hello"

var (
	tracer = otel.Tracer(name)
	logger = otelslog.NewLogger(name)
)

type Response struct {
	Msg              string      `json:"msg"`
	UpstreamResponse interface{} `json:"response,omitempty"`
}

type UpstreamResponseData struct {
	URL           string      `json:"url"`
	Data          interface{} `json:"data,omitempty"`
	UpstreamError string      `json:"error,omitempty"`
}

func processUpstreamCall(ctx context.Context, url string) *UpstreamResponseData {
	ctx, span := tracer.Start(ctx, "call-upstream")
	defer span.End()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	if err != nil {
		logger.Info("Error while creating new request", "err", err)
	} else {
		logger.Info("New Request", "headers", req.Header)
	}
	upstreamResp, err := http.DefaultClient.Do(req)
	var upstreamError string
	var upstreamData interface{}
	if err != nil {
		upstreamError = fmt.Sprintf("Error fetching upstream data: %v", err)
	} else {
		defer upstreamResp.Body.Close()
		upstreamBody, err := io.ReadAll(upstreamResp.Body)
		if err != nil {
			upstreamError = fmt.Sprintf("Error reading upstream response: %v", err)
		} else {
			err = json.Unmarshal(upstreamBody, &upstreamData)
			if err != nil {
				upstreamError = fmt.Sprintf("Error parsing upstream JSON: %v", err)
			}
		}
	}
	return &UpstreamResponseData{
		URL:           url,
		Data:          upstreamData,
		UpstreamError: upstreamError,
	}
}

func HandleHello(w http.ResponseWriter, r *http.Request) {
	logger.Info("handleHello", "X-Hello-1", r.Header.Get("X-Hello-1"))
	ctx := r.Context()
	propagator := otel.GetTextMapPropagator()
	ctx = propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))

	ctx, span := tracer.Start(ctx, "handle-hello")
	defer span.End()

	message := "hello"
	if val, found := os.LookupEnv("MESSAGE"); found {
		message = val
	}
	response := Response{
		Msg: message,
	}

	upstreamHost := os.Getenv("UPSTREAM_HOST")

	if upstreamHost != "" {
		upstreamURL := "http://" + upstreamHost + "/hello"
		upstreamResponse := processUpstreamCall(ctx, upstreamURL)
		if upstreamResponse != nil {
			response.UpstreamResponse = upstreamResponse
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
