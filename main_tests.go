package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestLoadConfig_Success(t *testing.T) {
	os.Setenv("POCKET_CONSUMER_KEY", "test_consumer_key")
	os.Setenv("POCKET_ACCESS_TOKEN", "test_access_token")
	os.Setenv("MASTODON_SERVER", "https://example.com")
	os.Setenv("MASTODON_TOKEN", "test_mastodon_token")

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
	}

	expectedConfig := &Config{
		PocketConsumerKey: "test_consumer_key",
		PocketAccessToken: "test_access_token",
		MastodonServer:    "https://example.com",
		MastodonToken:     "test_mastodon_token",
	}

	if !reflect.DeepEqual(config, expectedConfig) {
		t.Errorf("loadConfig returned incorrect config. Got: %+v, Expected: %+v", config, expectedConfig)
	}

	// Clean up environment variables
	os.Unsetenv("POCKET_CONSUMER_KEY")
	os.Unsetenv("POCKET_ACCESS_TOKEN")
	os.Unsetenv("MASTODON_SERVER")
	os.Unsetenv("MASTODON_TOKEN")
}

func TestLoadConfig_MissingEnvVars(t *testing.T) {
	_, err := loadConfig()
	if err == nil {
		t.Errorf("loadConfig should have failed due to missing environment variables")
	}
}

func TestFetchRecentPocketSaves_Success(t *testing.T) {
	// Mock the Pocket API response
	mockPocketServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got: %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		response := `{"list": {"123": {"item_id": "123", "resolved_url": "http://example.com/article1", "given_title": "Article One"}, "456": {"item_id": "456", "resolved_url": "http://example.com/article2", "given_title": "Article Two"}}}`
		w.Write([]byte(response))
	}))
	defer mockPocketServer.Close()

	config := &Config{
		PocketConsumerKey: "test_consumer_key",
		PocketAccessToken: "test_access_token",
	}

	items, err := fetchRecentPocketSaves(config)
	if err != nil {
		t.Fatalf("fetchRecentPocketSaves failed: %v", err)
	}

	expectedItems := []PocketItem{
		{ItemID: "123", ResolvedURL: "http://example.com/article1", GivenTitle: "Article One"},
		{ItemID: "456", ResolvedURL: "http://example.com/article2", GivenTitle: "Article Two"},
	}

	if len(items) != len(expectedItems) {
		t.Fatalf("Expected %d items, got %d", len(expectedItems), len(items))
	}

	// Pocket API response is a map, so the order might not be guaranteed.
	// We'll check if all expected items are present.
	found1 := false
	found2 := false
	for _, item := range items {
		if item.ItemID == "123" && item.ResolvedURL == "http://example.com/article1" && item.GivenTitle == "Article One" {
			found1 = true
		}
		if item.ItemID == "456" && item.ResolvedURL == "http://example.com/article2" && item.GivenTitle == "Article Two" {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Errorf("fetchRecentPocketSaves returned incorrect items. Got: %+v, Expected: %+v", items, expectedItems)
	}
}

func TestFetchRecentPocketSaves_Failure(t *testing.T) {
	// Mock the Pocket API returning an error
	mockPocketServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockPocketServer.Close()

	config := &Config{
		PocketConsumerKey: "test_consumer_key",
		PocketAccessToken: "test_access_token",
	}

	_, err := fetchRecentPocketSaves(config)
	if err == nil {
		t.Errorf("fetchRecentPocketSaves should have failed")
	}
}

func TestPostToMastodon_Success(t *testing.T) {
	// Mock the Mastodon API response
	mockMastodonServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got: %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test_mastodon_token" {
			t.Errorf("Expected Authorization header, got: %s", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("Expected Content-Type application/x-www-form-urlencoded, got: %s", r.Header.Get("Content-Type"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		response := `{"id": "12345", "content": "<p>Test Status</p>"}`
		w.Write([]byte(response))
	}))
	defer mockMastodonServer.Close()

	config := &Config{
		MastodonServer: mockMastodonServer.URL,
		MastodonToken:  "test_mastodon_token",
	}

	err := postToMastodon(config, "Test Status")
	if err != nil {
		t.Fatalf("postToMastodon failed: %v", err)
	}
}

func TestPostToMastodon_Failure(t *testing.T) {
	// Mock the Mastodon API returning an error
	mockMastodonServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockMastodonServer.Close()

	config := &Config{
		MastodonServer: mockMastodonServer.URL,
		MastodonToken:  "test_mastodon_token",
	}

	err := postToMastodon(config, "Test Status")
	if err == nil {
		t.Errorf("postToMastodon should have failed")
	}
}
