package utils

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func SetupOTelSDK() {
	// Set up to propagate `baggage` header
	otel.SetTextMapPropagator(propagation.Baggage{})
}
