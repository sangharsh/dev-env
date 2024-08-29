package otel_helper

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
)

const customHeaderName = "X-Hello-1"

// CustomHeaderPropagator implements custom header propagation
type CustomHeaderPropagator struct{}

// Inject sets the custom header into the carrier
func (chp CustomHeaderPropagator) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	if customValue := ctx.Value(customHeaderName); customValue != nil {
		carrier.Set(customHeaderName, customValue.(string))
	}
}

// Extract reads the custom header from the carrier and adds it to the context
func (chp CustomHeaderPropagator) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	customValue := carrier.Get(customHeaderName)
	if customValue != "" {
		ctx = context.WithValue(ctx, customHeaderName, customValue)
	}
	return ctx
}

// Fields returns the keys whose values are set with Inject.
func (chp CustomHeaderPropagator) Fields() []string {
	return []string{customHeaderName}
}
