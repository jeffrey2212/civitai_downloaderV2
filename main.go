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

	"github.com/cavaliergopher/grab/v3"
	"github.com/cheggaaa/pb/v3"
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

	// create grab client
	client := grab.NewClient()
	req, _ := grab.NewRequest(filepath, modelVersion.DownloadURL)
	req.NoResume = false
	req.HTTPRequest.Header.Set("Authorization", "Bearer "+civitai_api_key)
	// start download
	fmt.Printf("Downloading %v...\n", filename)
	resp := client.Do(req)
	// create and start new bar
	bar := pb.Full.Start64(resp.Size())

	// create a buffer to control how many bytes to read at a time
	buffer := make([]byte, 1024*1024) // 1MB

	// read from resp.Body into buffer until io.EOF
	for {
		bytesRead, err := resp.HTTPResponse.Body.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			break
		}
		bar.Add(bytesRead)
	}

	// finish progress bar
	bar.Finish()
	// check for errors
	if err := resp.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Download saved to ./%v \n", resp.Filename)
}
