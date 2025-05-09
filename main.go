package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Configuration struct to hold API keys and server details
type Config struct {
	PocketConsumerKey string
	PocketAccessToken string
	MastodonServer    string
	MastodonToken     string
}

// PocketItem represents a simplified Pocket save
type PocketItem struct {
	ItemID    string `json:"item_id"`
	ResolvedURL string `json:"resolved_url"`
	GivenTitle  string `json:"given_title"`
}

// PocketResponse represents the structure of the Pocket API response
type PocketResponse struct {
	List map[string]PocketItem `json:"list"`
}

// MastodonStatus represents the data for a Mastodon post
type MastodonStatus struct {
	Status string `json:"status"`
}

// loadConfig loads configuration from environment variables
func loadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, attempting to use environment variables")
	}

	consumerKey := os.Getenv("POCKET_CONSUMER_KEY")
	accessToken := os.Getenv("POCKET_ACCESS_TOKEN")
	mastodonServer := os.Getenv("MASTODON_SERVER")
	mastodonToken := os.Getenv("MASTODON_TOKEN")

	if consumerKey == "" || accessToken == "" || mastodonServer == "" || mastodonToken == "" {
		return nil, fmt.Errorf("missing required configuration (POCKET_CONSUMER_KEY, POCKET_ACCESS_TOKEN, MASTODON_SERVER, MASTODON_TOKEN)")
	}

	return &Config{
		PocketConsumerKey: consumerKey,
		PocketAccessToken: accessToken,
		MastodonServer:    mastodonServer,
		MastodonToken:     mastodonToken,
	}, nil
}

// fetchRecentPocketSaves fetches recent saves from Pocket
func fetchRecentPocketSaves(config *Config) ([]PocketItem, error) {
	endpoint := "https://getpocket.com/v3/get"
	values := url.Values{}
	values.Set("consumer_key", config.PocketConsumerKey)
	values.Set("access_token", config.PocketAccessToken)
	values.Set("sort", "newest")
	values.Set("count", "10") // Adjust the number of items as needed
	values.Set("detailType", "simple")

	resp, err := http.PostForm(endpoint, values)
	if err != nil {
		return nil, fmt.Errorf("error making Pocket API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Pocket API request failed with status code: %d", resp.StatusCode)
	}

	var pocketResponse PocketResponse
	err = json.NewDecoder(resp.Body).Decode(&pocketResponse)
	if err != nil {
		return nil, fmt.Errorf("error decoding Pocket API response: %w", err)
	}

	var items []PocketItem
	for _, item := range pocketResponse.List {
		items = append(items, item)
	}

	log.Printf("Successfully fetched %d recent Pocket saves", len(items))
	return items, nil
}

// postToMastodon posts a status to Mastodon
func postToMastodon(config *Config, status string) error {
	endpoint := fmt.Sprintf("%s/api/v1/statuses", config.MastodonServer)
	payload := url.Values{}
	payload.Set("status", status)

	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return fmt.Errorf("error creating Mastodon API request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.MastodonToken))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Body = http.NewUrlEncodedReader(payload)
	defer req.Body.Close()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making Mastodon API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Attempt to read the error body for more details
		var errorResponse map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			return fmt.Errorf("Mastodon API request failed with status code: %d, error: %v", resp.StatusCode, errorResponse)
		}
		return fmt.Errorf("Mastodon API request failed with status code: %d", resp.StatusCode)
	}

	log.Printf("Successfully posted to Mastodon: %s", status)
	return nil
}

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	recentSaves, err := fetchRecentPocketSaves(config)
	if err != nil {
		log.Printf("Error fetching Pocket saves: %v", err)
		return
	}

	// For simplicity, we'll just post the title and URL of each new save.
	// In a more advanced scenario, you might want to track which items have already been posted.
	for _, save := range recentSaves {
		status := fmt.Sprintf("New Pocket save: %s - %s #Pocket", save.GivenTitle, save.ResolvedURL)
		err := postToMastodon(config, status)
		if err != nil {
			log.Printf("Error posting to Mastodon for item %s: %v", save.ItemID, err)
		}
		// Add a small delay to avoid rate limiting
		time.Sleep(2 * time.Second)
	}

	log.Println("Finished processing recent Pocket saves.")
}
