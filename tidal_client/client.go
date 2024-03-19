package tidal_client

import (
	"encoding/base64"
	"encoding/json"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var bearerToken string

func init() {
	// Load the .env file in the init function to ensure it's done before the main function
	err := godotenv.Load("tidal.env") // This will load the .env file from the same directory as the go file
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	// Get initial token and refresh token 1 minute before expiration
	getAuthToken()
}

func getAuthToken() {
	// Get the client_id and client_secret from the environment
	clientId, clientSecret, url := os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"), os.Getenv("URL_AUTH")
	creds := base64.StdEncoding.EncodeToString([]byte(clientId + ":" + clientSecret))

	// Create a new request to the Tidal Authentication endpoint
	req, err := http.NewRequest(
		"POST", url,
		strings.NewReader("grant_type=client_credentials"))
	req.Header.Add("Authorization", "Basic "+creds)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error calling Tidal with default client: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalf("Error closing response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) // Ignoring error to simplify; handle appropriately in real code
		log.Fatalf("request failed with status code: %d and body: %s", resp.StatusCode, body)
	}

	var data struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Fatalf("Error decoding body of Tidal client response: %v", err)
	}

	if data.AccessToken == "" {
		log.Fatalf("Error Tidal client returned an empty bearer token: %v", err)
	}

	bearerToken = data.AccessToken

	// Wait until 1 minute before token expiration, then get new token
	go func(expiresIn int) {
		time.Sleep(time.Duration(expiresIn-60) * time.Second)
		getAuthToken()
	}(data.ExpiresIn)
}

// TRACK API
//func getTracksByArtist(artistId string, countryCode string, offset int, limit int)
