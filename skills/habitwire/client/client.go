// Package client provides HTTP client functionality for HabitWire API
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Config holds HabitWire API configuration
type Config struct {
	BaseURL string
	APIKey  string
}

// Client wraps HTTP client for HabitWire API
type Client struct {
	config     Config
	httpClient *http.Client
}

// New creates a new HabitWire API client from environment
func New() (*Client, error) {
	baseURL := os.Getenv("HABITWIRE_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("HABITWIRE_URL environment variable is required")
	}
	// Remove trailing slash if present
	baseURL = strings.TrimSuffix(baseURL, "/")

	apiKey := os.Getenv("HABITWIRE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("HABITWIRE_API_KEY environment variable is required")
	}

	return &Client{
		config: Config{
			BaseURL: baseURL,
			APIKey:  apiKey,
		},
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// NewWithConfig creates a client with explicit config
func NewWithConfig(config Config) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Request performs an HTTP request to the HabitWire API
func (c *Client) Request(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	// Prepend /api/v1 to endpoint
	url := c.config.BaseURL + "/api/v1" + endpoint
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "ApiKey "+c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Get performs a GET request
func (c *Client) Get(endpoint string) ([]byte, error) {
	return c.Request(http.MethodGet, endpoint, nil)
}

// Post performs a POST request
func (c *Client) Post(endpoint string, body interface{}) ([]byte, error) {
	return c.Request(http.MethodPost, endpoint, body)
}

// Put performs a PUT request
func (c *Client) Put(endpoint string, body interface{}) ([]byte, error) {
	return c.Request(http.MethodPut, endpoint, body)
}

// Delete performs a DELETE request
func (c *Client) Delete(endpoint string) ([]byte, error) {
	return c.Request(http.MethodDelete, endpoint, nil)
}
