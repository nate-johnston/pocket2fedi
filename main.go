package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mattn/go-mastodon"
	"github.com/motemen/go-pocket/api"
	"github.com/motemen/go-pocket/auth"
	"golang.org/x/oauth2"
)

// Configuration struct to hold API keys and tokens
type Config struct {
	PocketConsumerKey string
	PocketAccessToken string
	MastodonServer    string
	MastodonToken     string
}

// PocketItem represents a simplified Pocket item structure
type PocketItem struct {
	Title string
	URL   string
}

// loadConfigFromEnv loads configuration from environment variables
func loadConfigFromEnv() (*Config, error) {
	config := &Config{
		PocketConsumerKey: os.Getenv("POCKET_CONSUMER_KEY"),
		PocketAccessToken: os.Getenv("POCKET_ACCESS_TOKEN"),
		MastodonServer:    os.Getenv("MASTODON_SERVER"),
		MastodonToken:     os.Getenv("MASTODON_TOKEN"),
	}

	if config.PocketConsumerKey == "" || config.PocketAccessToken == "" || config.MastodonServer == "" || config.MastodonToken == "" {
		return nil, fmt.Errorf("missing required environment variables")
	}

	return config, nil
}

// getRecentPocketSaves fetches recent Pocket saves
func getRecentPocketSaves(ctx context.Context, consumerKey, accessToken string) ([]*PocketItem, error) {
	client, err := api.NewClient(&oauth2.Config{}, &oauth2.Token{AccessToken: accessToken})
	if err != nil {
		return nil, fmt.Errorf("failed to create Pocket client: %w", err)
	}

	params := &api.RetrieveInput{
		Count:     10, // Fetch the 10 most recent items, adjust as needed
		Sort:      api.SortNewest,
		DetailType: api.DetailSimple,
	}

	output, err := client.Retrieve(ctx, consumerKey, params)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Pocket items: %w", err)
	}

	var recentSaves []*PocketItem
	for _, item := range output.List {
		if item.Status == 0 { // Unarchived items
			recentSaves = append(recentSaves, &PocketItem{
				Title: item.ResolvedTitle,
				URL:   item.ResolvedURL,
			})
		}
	}

	log.Printf("Successfully retrieved %d recent Pocket saves", len(recentSaves))
	return recentSaves, nil
}

// postToMastodon posts a status to Mastodon
func postToMastodon(ctx context.Context, server, accessToken, status string) error {
	client := mastodon.NewClient(&mastodon.Config{
		Server:       server,
		AccessToken:  accessToken,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	})

	_, err := client.PostStatus(ctx, &mastodon.Status{
		Status: status,
	})

	if err != nil {
		return fmt.Errorf("failed to post to Mastodon: %w", err)
	}

	log.Printf("Successfully posted to Mastodon: %s", status)
	return nil
}

func main() {
	config, err := loadConfigFromEnv()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	ctx := context.Background()

	recentSaves, err := getRecentPocketSaves(ctx, config.PocketConsumerKey, config.PocketAccessToken)
	if err != nil {
		log.Printf("Error fetching Pocket saves: %v", err)
		return
	}

	for _, save := range recentSaves {
		status := fmt.Sprintf("New Pocket save: %s - %s", save.Title, save.URL)
		err := postToMastodon(ctx, config.MastodonServer, config.MastodonToken, status)
		if err != nil {
			log.Printf("Error posting to Mastodon for '%s': %v", save.Title, err)
		}
		// Add a small delay to avoid rate limiting
		time.Sleep(2 * time.Second)
	}

	log.Println("Finished processing recent Pocket saves.")
}

// --- Unit Tests ---

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoadConfigFromEnv_Success(t *testing.T) {
	os.Setenv("POCKET_CONSUMER_KEY", "test_consumer_key")
	os.Setenv("POCKET_ACCESS_TOKEN", "test_access_token")
	os.Setenv("MASTODON_SERVER", "https://mastodon.example")
	os.Setenv("MASTODON_TOKEN", "test_mastodon_token")

	_, err := loadConfigFromEnv()
	if err != nil {
		t.Errorf("loadConfigFromEnv failed: %v", err)
	}

	os.Unsetenv("POCKET_CONSUMER_KEY")
	os.Unsetenv("POCKET_ACCESS_TOKEN")
	os.Unsetenv("MASTODON_SERVER")
	os.Unsetenv("MASTODON_TOKEN")
}

func TestLoadConfigFromEnv_MissingVariable(t *testing.T) {
	os.Setenv("POCKET_CONSUMER_KEY", "test_consumer_key")

	_, err := loadConfigFromEnv()
	if err == nil {
		t.Errorf("loadConfigFromEnv should have failed with missing variable")
	}

	os.Unsetenv("POCKET_CONSUMER_KEY")
}

func TestGetRecentPocketSaves_Success(t *testing.T) {
	// Mock Pocket API response
	mockPocketServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"list": {
				"123": {
					"resolved_title": "Test Article 1",
					"resolved_url": "https://example.com/article1",
					"status": "0"
				},
				"456": {
					"resolved_title": "Test Article 2",
					"resolved_url": "https://example.com/article2",
					"status": "2"
				}
			}
		}`))
	}))
	defer mockPocketServer.Close()

	// Temporarily patch the Pocket API endpoint for testing
	originalEndpoint := api.Endpoint
	api.Endpoint = mockPocketServer.URL
	defer func() { api.Endpoint = originalEndpoint }()

	ctx := context.Background()
	consumerKey := "test_consumer_key"
	accessToken := "test_access_token"

	saves, err := getRecentPocketSaves(ctx, consumerKey, accessToken)
	if err != nil {
		t.Fatalf("getRecentPocketSaves failed: %v", err)
	}

	if len(saves) != 1 {
		t.Errorf("Expected 1 save, got %d", len(saves))
	}

	if saves[0].Title != "Test Article 1" {
		t.Errorf("Expected title 'Test Article 1', got '%s'", saves[0].Title)
	}

	if saves[0].URL != "https://example.com/article1" {
		t.Errorf("Expected URL 'https://example.com/article1', got '%s'", saves[0].URL)
	}
}

func TestGetRecentPocketSaves_Failure(t *testing.T) {
	// Mock Pocket API returning an error
	mockPocketServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockPocketServer.Close()

	// Temporarily patch the Pocket API endpoint for testing
	originalEndpoint := api.Endpoint
	api.Endpoint = mockPocketServer.URL
	defer func() { api.Endpoint = originalEndpoint }()

	ctx := context.Background()
	consumerKey := "test_consumer_key"
	accessToken := "test_access_token"

	_, err := getRecentPocketSaves(ctx, consumerKey, accessToken)
	if err == nil {
		t.Errorf("getRecentPocketSaves should have failed")
	}
}

func TestPostToMastodon_Success(t *testing.T) {
	// Mock Mastodon API response
	mockMastodonServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Simulate a successful post response if needed
	}))
	defer mockMastodonServer.Close()

	ctx := context.Background()
	server := mockMastodonServer.URL
	accessToken := "test_mastodon_token"
	status := "Test Mastodon post"

	err := postToMastodon(ctx, server, accessToken, status)
	if err != nil {
		t.Errorf("postToMastodon failed: %v", err)
	}
}

func TestPostToMastodon_Failure(t *testing.T) {
	// Mock Mastodon API returning an error
	mockMastodonServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockMastodonServer.Close()

	ctx := context.Background()
	server := mockMastodonServer.URL
	accessToken := "test_mastodon_token"
	status := "Test Mastodon post"

	err := postToMastodon(ctx, server, accessToken, status)
	if err == nil {
		t.Errorf("postToMastodon should have failed")
	}
}
