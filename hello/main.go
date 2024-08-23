package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
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

func main() {
	fmt.Println("Hello, World!")
	r := mux.NewRouter()

	r.HandleFunc("/statusz", handleStatusz)
	r.HandleFunc("/hello", handleHello)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting http server at port %v", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func processUpstreamCall(url string) *UpstreamResponseData {
	var upstreamError string
	var upstreamData interface{}
	upstreamResp, err := http.Get(url)
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

func handleHello(w http.ResponseWriter, r *http.Request) {
	message := "hello"
	if val, found := os.LookupEnv("MESSAGE"); found {
		message = val
	}
	response := Response{
		Msg: message,
	}

	upstreamHost := os.Getenv("UPSTREAM_HOST")

	var upstreamURL string
	if upstreamHost != "" {
		upstreamURL = upstreamHost + "/hello"
	}
	upstreamResponse := processUpstreamCall(upstreamURL)
	if upstreamResponse != nil {
		response.UpstreamResponse = upstreamResponse
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleStatusz(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request at %s", r.Method, r.RequestURI)
	response := map[string]string{
		"message": "all ok",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
