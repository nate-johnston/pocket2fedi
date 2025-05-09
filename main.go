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
		Count:      10, // Fetch the 10 most recent items, adjust as needed
		Sort:       api.SortNewest,
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
		Server:      server,
		AccessToken: accessToken,
		HTTPClient:  &http.Client{Timeout: 10 * time.Second},
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
