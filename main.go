package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/schollz/progressbar/v3"
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

func downloadModelsFromFile(filepath string, basepath string) {
	// Open the file
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Printf("Error opening file. %v\n", err)
		return
	}
	defer file.Close()

	// load .env file
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	civitai_api_key := os.Getenv("CIVITAI_API_KEY")

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		parts := strings.Split(line, ":")

		// Extract base model
		baseModel := parts[2]
	
		// Extract model ID and version ID using a regular expression
		re := regexp.MustCompile(`(\d+)@(\d+)`)
		matches := re.FindStringSubmatch(parts[5])

		modelID := matches[1]
		versionID := matches[2]

		// The URL of the RESTful API
		url := "https://civitai.com/api/v1/models/" + modelID

		// Call the function to get the response from the API
		responses, err := getAPIResponse(url, civitai_api_key)
		if err != nil {
			fmt.Printf("Error getting response from API. %v\n", err)
			continue
		}

		var modelVersion ModelVersion
		versionIDInt, err := strconv.Atoi(versionID)
		if err != nil {
			fmt.Printf("Error converting versionID to integer. %v\n", err)
			continue
		}

		for _, v := range responses.ModelVersions {
			if v.ID == versionIDInt {
				modelVersion = v
				break
			}
		}

		filename := modelVersion.Files[0].Name
		baseFolder := basepath
		// if baseFolder does not exist, create it
		if _, err := os.Stat(baseFolder); os.IsNotExist(err) {
			err = os.Mkdir(baseFolder, 0755)
			if err != nil {
				fmt.Printf("Error creating base folder. %v\n", err)
				continue
			}
		}

		filetype := responses.Type
		subfolder := "others"
		// check filetype is = "lora", "checkpoints", or others, put in corresponding folder
		if filetype == "LORA" {
			subfolder = "lora/"
		} else if filetype == "Checkpoint" {
			subfolder = "checkpoints/"
		} else if filetype == "TextualInversion" {
			subfolder = "embeddings/"
		}

		save_path := baseFolder + subfolder + baseModel
		// if subfolder does not exist, create it
		if _, err := os.Stat(baseFolder + subfolder); os.IsNotExist(err) {
			err = os.Mkdir(baseFolder+subfolder, 0755)
			if err != nil {
				fmt.Printf("Error creating subfolder. %v\n", err)
				continue
			}
		}

		// if save path does not exist, create it
		if _, err := os.Stat(save_path); os.IsNotExist(err) {
			err = os.Mkdir(save_path, 0755)
			if err != nil {
				fmt.Printf("Error creating subfolder. %v\n", err)
				continue
			}
		}

		filepath := save_path + "/" + filename

		req, err := http.NewRequest("GET", modelVersion.DownloadURL, nil)
		if err != nil {
			fmt.Printf("Error creating HTTP request. %v\n", err)
			continue
		}
		req.Header.Set("Authorization", "Bearer "+civitai_api_key)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("Error making HTTP request. %v\n", err)
			continue
		}
		defer resp.Body.Close()

		f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Error opening file. %v\n", err)
			continue
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

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file. %v\n", err)
		return
	}
}
func main() {
	filepath := "download.txt"
	basepath := "./models/"
	downloadModelsFromFile(filepath, basepath)
}
