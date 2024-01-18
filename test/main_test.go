package main

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"encoding/json"

)

func TestFetchModelDetails(t *testing.T) {
    // Create a mock server
    mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        response := ModelDetails{
            ID:   175522,
            Name: "Sample Model",
            Creator: struct {
                Username string `json:"username"`
            }{
                Username: "XRYCJ",
            },
        }
        json.NewEncoder(w).Encode(response)
    }))
    defer mockServer.Close()

    // Replace the API URL with the mock server URL in your FetchModelDetails function
    // For example, if you have a variable or constant for the API URL, you would set it here

    // Call the function
    details, err := FetchModelDetails("175522")
    if err != nil {
        t.Fatalf("Failed to fetch model details: %s", err)
    }

    // Check if the response matches the expected output
    if details.ID != 175522 || details.Name != "Sample Model" || details.Creator.Username != "XRYCJ" {
        t.Errorf("Unexpected response: got %v", details)
    }
}
