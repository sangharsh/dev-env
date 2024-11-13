package main

import (
	"log"
	"net/http"
	"path/filepath"
)

const (
	tlsDir      = "/etc/webhook/certs"
	tlsCertFile = "tls.crt"
	tlsKeyFile  = "tls.key"
)

func main() {
	ac := &admissionController{}

	certPath := filepath.Join(tlsDir, tlsCertFile)
	keyPath := filepath.Join(tlsDir, tlsKeyFile)

	mux := http.NewServeMux()
	mux.HandleFunc("/validate", ac.serve)

	server := &http.Server{
		Addr:    ":8443",
		Handler: mux,
	}

	log.Printf("Starting server on port 8443")
	log.Printf("Using TLS certificate: %s", certPath)
	log.Printf("Using TLS key: %s", keyPath)

	if err := server.ListenAndServeTLS(certPath, keyPath); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
