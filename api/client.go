package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type IClient interface {
	Login(context.Context) error
	GetGateway(context.Context) (GatewayResponse, error)
}

// ClientConfig holds configuration for the API client
type ClientConfig struct {
	BaseURL  string
	Username string
	Password string
	Logger   *log.Logger
}

// Client handles communication with the gateway API
type Client struct {
	config     ClientConfig
	httpClient *http.Client
	auth       *authToken
}

// NewClient creates a new API client with the provided base URL
func NewClient(baseURL string, logger *log.Logger) *Client {
	username := os.Getenv("GATEWAY_USERNAME")
	password := os.Getenv("GATEWAY_PASSWORD")
	if username == "" || password == "" {
		log.Fatal("GATEWAY_USERNAME and GATEWAY_PASSWORD must be set")
	}

	config := ClientConfig{
		BaseURL:  baseURL,
		Username: username,
		Password: password,
		Logger:   logger,
	}

	return NewClientWithConfig(config, nil)
}

// NewClientWithConfig creates a new API client with the provided configuration and HTTP client
func NewClientWithConfig(config ClientConfig, httpClient *http.Client) *Client {
	if config.Logger == nil {
		config.Logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	if httpClient == nil {
		httpClient = &http.Client{}
	}

	config.Logger.Printf("Gateway URL: %s", config.BaseURL)

	return &Client{
		config:     config,
		httpClient: httpClient,
	}
}

// Login authenticates with the API and stores the auth token
func (c *Client) Login(ctx context.Context) error {
	loginBody, err := json.Marshal(struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: c.config.Username,
		Password: c.config.Password,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal login body: %w", err)
	}

	respBody, err := c.post(ctx, "auth/login", bytes.NewBuffer(loginBody), nil)
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}

	var authResponse authResponse
	err = json.Unmarshal(respBody, &authResponse)
	if err != nil {
		return fmt.Errorf("unexpected login response: %w", err)
	}

	c.auth = &authResponse.Auth
	return nil
}

// GetGateway retrieves gateway information from the API
func (c *Client) GetGateway(ctx context.Context) (GatewayResponse, error) {
	var gateway GatewayResponse

	body, err := c.get(ctx, "gateway/?get=all")
	if err != nil {
		return gateway, fmt.Errorf("failed to get gateway info: %w", err)
	}

	err = json.Unmarshal(body, &gateway)
	if err != nil {
		return gateway, fmt.Errorf("failed to unmarshal gateway response: %w", err)
	}

	return gateway, nil
}

// ensureAuthenticated makes sure the client has a valid auth token
func (c *Client) ensureAuthenticated(ctx context.Context) error {
	if c.auth == nil || isTokenExpired(c.auth.Expiration) {
		return c.Login(ctx)
	}
	return nil
}

// get performs an HTTP GET request to the API
func (c *Client) get(ctx context.Context, endpoint string) ([]byte, error) {
	if err := c.ensureAuthenticated(ctx); err != nil {
		return nil, err
	}

	c.config.Logger.Printf("GET %s", endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(endpoint), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/text")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.auth.Token))

	return c.doRequest(req)
}

// post performs an HTTP POST request to the API
func (c *Client) post(ctx context.Context, endpoint string, body io.Reader, headers map[string]string) ([]byte, error) {
	c.config.Logger.Printf("POST %s", endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url(endpoint), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Apply any custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.doRequest(req)
}

// doRequest executes an HTTP request and processes the response
func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return readResponse(resp)
}

// url constructs the full URL for an API endpoint
func (c *Client) url(endpoint string) string {
	return fmt.Sprintf("%s/%s", c.config.BaseURL, endpoint)
}

// readResponse reads and validates the HTTP response
func readResponse(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, body)
	}

	return body, nil
}

// isTokenExpired checks if an auth token has expired
func isTokenExpired(expiration int64) bool {
	// Add a small buffer to account for clock skew and network latency
	const expirationBuffer = 30 * time.Second
	return time.Unix(expiration, 0).Add(-expirationBuffer).Before(time.Now())
}
