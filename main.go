package main

import (
    "encoding/json"
    "net/http"
)

// ModelDetails represents the structure of the API response
type ModelDetails struct {
    ID     int    `json:"id"`
    Name   string `json:"name"`
    Creator struct {
        Username string `json:"username"`
    } `json:"creator"`
    // Add other fields as needed
}

// This function will be the one making the API call in your application
func FetchModelDetails(modelID string) (*ModelDetails, error) {
    // Implementation will go here
	apiURL := "https://civitai.com/api/v1/models/" + modelID
	resp, err := http.Get(apiURL)
	if err != nil {	
		return nil, err
	}
	defer resp.Body.Close()

	var details ModelDetails
	err = json.NewDecoder(resp.Body).Decode(&details)
	if err != nil {
		return nil, err
	}

	return &details, nil
}

func main(){
	FetchModelDetails("175522")
}