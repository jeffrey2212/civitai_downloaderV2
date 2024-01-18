package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/cheggaaa/pb/v3"
)


type CivitaiResponse struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	ModelVersions []ModelVersion`json:"modelVersions"`
}

type ModelVersion struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	BaseModel   string `json:"baseModel"`
	DownloadURL string `json:"downloadUrl"`
	Files       []struct {
		Name string `json:"name"`
	} `json:"files"`
} 

func getAPIResponse(url string) (*CivitaiResponse, error) {
	// Send a GET request to the URL
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error sending request to API endpoint. %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body. %v\n", err)
		return nil, err
	}

	var responses CivitaiResponse
	err = json.Unmarshal(body, &responses)
	if err != nil {
		fmt.Printf("Error decoding JSON response. %v\n", err)
		return nil, err
	}
	return &responses, nil
}

func downloadFile(url string, filepath string) error {
	// Send a GET request to the URL
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Create a progress bar
	bar := pb.Full.Start64(resp.ContentLength)
	barReader := bar.NewProxyReader(resp.Body)

	// Write the body to file
	_, err = io.Copy(out, barReader)
	return err
}

func main() {

	// handle command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Please provide the model ID and version ID in the format modelID@versionID")
		return
	}

	arg := os.Args[1]
	ids := strings.Split(arg, "@")
	if len(ids) != 2 {
		fmt.Println("Invalid argument. Please provide the model ID and version ID in the format modelID@versionID")
		return
	}

	modelID := ids[0]
    versionID := ids[1]

	// The URL of the RESTful API
	url := "https://civitai.com/api/v1/models/" + modelID

	// Call the function to get the response from the API
	responses, err := getAPIResponse(url)
	if err != nil {
		fmt.Printf("Error getting response from API. %v\n", err)
		return
	}

	var modelVersion ModelVersion 
	versionIDInt, err := strconv.Atoi(versionID)
	if err != nil {
		fmt.Printf("Error converting versionID to integer. %v\n", err)
		return
	}

	for _, v := range responses.ModelVersions {
		if v.ID == versionIDInt {
			modelVersion = v
			break
		}
	}

	filepath := "./" + modelVersion.Files[0].Name
	err = downloadFile(modelVersion.DownloadURL, filepath)
	if err != nil {
		fmt.Printf("Error downloading file. %v\n", err)
		return
	}

	//fmt.Printf("Response: %+v\n", responses)
	//fmt.Printf("Response: %+v\n", responses.ModelVersions[0].DownloadURL)
}
