package main_test

import (
	"net/http"
	"net/http/httptest"
	"os"
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
