package main

import (
	//"encoding/json"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Response structure to match the JSON returned by httpbin.org
type HTTPBinResponse struct {
	URL string `json:"url"`
}

type CivitaiResponse struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	ModelVersions []struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		BaseModel   string `json:"baseModel"`
		DownloadURL string `json:"downloadUrl"`
		Files       []struct {
			Name string `json:"name"`
		} `json:"files"`
	} `json:"modelVersions"`
}

func main() {
	// The URL of the RESTful API
	url := "https://civitai.com/api/v1/models?limit=1"

	// Send a GET request to the URL
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error sending request to API endpoint. %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body. %v\n", err)
		return
	}

	var responses CivitaiResponse
	err = json.Unmarshal(body, &responses)
	if err != nil {
		fmt.Printf("Error decoding JSON response. %v\n", err)
		return
	}

	fmt.Printf("Response: %+v\n", responses)
}
