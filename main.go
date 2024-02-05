package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/joho/godotenv"
)

type CivitaiResponse struct {
	ID            int            `json:"id"`
	Name          string         `json:"name"`
	Type          string         `json:"type"`
	ModelVersions []ModelVersion `json:"modelVersions"`
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

func getAPIResponse(url string, api_key string) (*CivitaiResponse, error) {
	// Send a GET request to the URL
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request. %v\n", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+api_key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error sending request to API endpoint. %v\n", err)
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Error closing response body. %v\n", err)
		}
	}(resp.Body)

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

	// load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	civitai_api_key := os.Getenv("CIVITAI_API_KEY")

	// The URL of the RESTful API
	url := "https://civitai.com/api/v1/models/" + modelID

	// Call the function to get the response from the API
	responses, err := getAPIResponse(url, civitai_api_key)
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

	filename := modelVersion.Files[0].Name
	baseFolder := "./models/"
	// if baseFolder does not exist, create it
	if _, err := os.Stat(baseFolder); os.IsNotExist(err) {
		err = os.Mkdir(baseFolder, 0755)
		if err != nil {
			fmt.Printf("Error creating base folder. %v\n", err)
			return
		}
	}

	filetype := responses.Type
	subfolder := "others"
	// check filetype is = "lora", "checkpoints", or others, put in coresponding folder
	if filetype == "LORA" {
		subfolder = "lora"
	} else if filetype == "Checkpoint" {
		subfolder = "checkpoints"
	} else if filetype == "TextualInversion" {
		subfolder = "embeddings"
	}

	// if subfolder does not exist, create it
	if _, err := os.Stat(baseFolder + subfolder); os.IsNotExist(err) {
		err = os.Mkdir(baseFolder+subfolder, 0755)
		if err != nil {
			fmt.Printf("Error creating subfolder. %v\n", err)
			return
		}
	}

	filepath := baseFolder + subfolder + "/" + filename

	req, err := http.NewRequest("GET", modelVersion.DownloadURL, nil)
	if err != nil {
		fmt.Printf("Error creating HTTP request. %v\n", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+civitai_api_key)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error making HTTP request. %v\n", err)
		return
	}
	defer resp.Body.Close()

	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file. %v\n", err)
		return
	}
	defer f.Close()

	fmt.Printf("Downloading %s to %s\n", modelVersion.DownloadURL, filepath)
	// create a progress bar
	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		filename,
	)
	io.Copy(io.MultiWriter(f, bar), resp.Body)
}
