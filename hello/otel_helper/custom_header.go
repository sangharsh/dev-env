package otel_helper

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
)

type CustomHeaderKey string

var customHeaders = [...]string{"X-Hello-1", "X-Hello-2"}

// CustomHeaderPropagator implements custom header propagation
type CustomHeaderPropagator struct{}

// Inject sets the custom header into the carrier
func (chp CustomHeaderPropagator) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	for _, header := range customHeaders {
		if customValue := ctx.Value(CustomHeaderKey(header)); customValue != nil {
			carrier.Set(header, customValue.(string))

		}
	}
}

// Extract reads the custom header from the carrier and adds it to the context
func (chp CustomHeaderPropagator) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	for _, header := range customHeaders {
		headerValue := carrier.Get(header)
		if headerValue != "" {
			ctx = context.WithValue(ctx, CustomHeaderKey(header), headerValue)
		}
	}
	return ctx
}

// Fields returns the keys whose values are set with Inject.
func (chp CustomHeaderPropagator) Fields() []string {
	return customHeaders[:]
}
