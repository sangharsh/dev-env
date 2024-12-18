package context_propagation

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func SetupOTelSDK() {
	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
		// CustomHeaderPropagator{},
	)
}
