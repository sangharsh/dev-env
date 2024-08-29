module github.com/sangharsh/dev-env/hello

go 1.23

require (
	go.opentelemetry.io/otel v1.29.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutlog v0.5.0
	go.opentelemetry.io/otel/log v0.5.0
	go.opentelemetry.io/otel/sdk/log v0.5.0
)

require (
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
)

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/contrib/bridges/otelslog v0.4.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.54.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.29.0
	go.opentelemetry.io/otel/metric v1.29.0 // indirect
	go.opentelemetry.io/otel/sdk v1.29.0
	go.opentelemetry.io/otel/trace v1.29.0 // indirect
)
