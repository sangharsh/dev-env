package main

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/sangharsh/dev-env/admctl/internal/handlers"
)

const (
	tlsDir      = "/etc/webhook/certs"
	tlsCertFile = "tls.crt"
	tlsKeyFile  = "tls.key"
)

func main() {
	ac := &handlers.AdmissionController{}

	certPath := filepath.Join(tlsDir, tlsCertFile)
	keyPath := filepath.Join(tlsDir, tlsKeyFile)

	mux := http.NewServeMux()
	mux.HandleFunc("/validate", ac.Serve)

	server := &http.Server{
		Addr:    ":8443",
		Handler: mux,
	}

	log.Printf("Starting server on port 8443")
	log.Printf("Using TLS certificate: %s", certPath)
	log.Printf("Using TLS key: %s", keyPath)

	err := server.ListenAndServeTLS(certPath, keyPath)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
