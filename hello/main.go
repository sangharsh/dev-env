package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/sangharsh/dev-env/hello/hello"
	"github.com/sangharsh/dev-env/hello/otel_helper"
)

// Credits: https://opentelemetry.io/docs/languages/go/getting-started/#initialize-the-opentelemetry-sdk
func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() (err error) {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Set up OpenTelemetry.
	otel_helper.SetupOTelSDK()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// Start HTTP server.
	srv := &http.Server{
		Addr:         ":" + port,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      createHTTPHandler(),
	}
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	// Wait for interruption.
	select {
	case err = <-srvErr:
		// Error when starting HTTP server.
		return
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
	err = srv.Shutdown(context.Background())
	return
}

func createHTTPHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/statusz", handleStatusz)
	mux.HandleFunc("/hello", hello.HandleHello)
	return mux
}

func handleStatusz(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"message": "all ok",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
