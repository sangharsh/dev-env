package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sangharsh/dev-env/samples/hello/app/utils"
)

type Response struct {
	Msg              string      `json:"msg"`
	UpstreamResponse interface{} `json:"response,omitempty"`
}

type UpstreamResponseData struct {
	Data          interface{} `json:"data,omitempty"`
	UpstreamError string      `json:"error,omitempty"`
}

func helloUpstream(inRequest *http.Request, host string) *UpstreamResponseData {
	upstreamResponse := &UpstreamResponseData{}
	url := fmt.Sprintf("http://%s/hello", host)
	responseJSON, err := utils.FetchJSONResponse(inRequest, url)
	if err != nil {
		upstreamResponse.UpstreamError = err.Error()
	} else {
		upstreamResponse.Data = responseJSON
	}
	return upstreamResponse
}

func HandleHello(w http.ResponseWriter, r *http.Request) {
	log.Printf("handleHello headers: %v", r.Header)

	var response Response

	message := os.Getenv("MESSAGE")
	if message == "" {
		message = "hello"
	}
	response.Msg = message

	upstreamHost := os.Getenv("UPSTREAM_HOST")
	if upstreamHost != "" {
		response.UpstreamResponse = helloUpstream(r, upstreamHost)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
