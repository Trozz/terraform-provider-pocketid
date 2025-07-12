package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Client represents a Pocket-ID API client
type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
}

// RateLimitError represents a 429 rate limit error with optional Retry-After information
type RateLimitError struct {
	StatusCode int
	Message    string
	RetryAfter string // Can be seconds or HTTP-date
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter != "" {
		return fmt.Sprintf("HTTP %d: %s (Retry-After: %s)", e.StatusCode, e.Message, e.RetryAfter)
	}
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// NewClient creates a new Pocket-ID API client
func NewClient(baseURL, apiToken string, skipTLSVerify bool, timeout int64) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	if apiToken == "" {
		return nil, fmt.Errorf("API token is required")
	}

	// Configure HTTP client with TLS settings
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			// Allow users to skip TLS verification for development environments
			// This is controlled by provider configuration and defaults to false
			InsecureSkipVerify: skipTLSVerify, // #nosec G402 - Legitimate use case for development
		},
	}

	return &Client{
		baseURL:  baseURL,
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout:   time.Duration(timeout) * time.Second,
			Transport: transport,
		},
	}, nil
}

// doRequest performs an HTTP request to the Pocket-ID API
func (c *Client) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
	return c.doRequestWithContext(context.Background(), method, endpoint, body)
}

// doRequestWithContext performs an HTTP request to the Pocket-ID API with context support
func (c *Client) doRequestWithContext(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error) {
	const maxRetries = 3
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		var backoff time.Duration

		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			backoff = time.Duration(1<<(attempt-1)) * time.Second

			// Special handling for rate limit errors to respect Retry-After header
			if rateLimitErr, ok := lastErr.(*RateLimitError); ok && rateLimitErr.RetryAfter != "" {
				retryAfterSeconds := parseRetryAfter(rateLimitErr.RetryAfter)
				if retryAfterSeconds > 0 {
					backoff = time.Duration(retryAfterSeconds) * time.Second
					tflog.Info(ctx, "Rate limited, using Retry-After header", map[string]interface{}{
						"retry_after":     rateLimitErr.RetryAfter,
						"backoff_seconds": retryAfterSeconds,
					})
				}
			}

			tflog.Debug(ctx, "Retrying request after backoff", map[string]interface{}{
				"attempt": attempt,
				"backoff": backoff.String(),
			})

			select {
			case <-time.After(backoff):
				// Continue with retry
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry backoff: %w", ctx.Err())
			}
		}

		respBody, err := c.doSingleRequest(ctx, method, endpoint, body)
		if err == nil {
			return respBody, nil
		}

		lastErr = err

		// Determine if error is retryable
		if !isRetryableError(err) {
			return nil, err
		}

		tflog.Warn(ctx, "Request failed with retryable error", map[string]interface{}{
			"error":        err.Error(),
			"attempt":      attempt + 1,
			"max_attempts": maxRetries + 1,
		})
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries+1, lastErr)
}

// isRetryableError determines if an error is retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Network errors are retryable
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "timeout") {
		return true
	}

	// 5xx errors are retryable
	if strings.Contains(errStr, "HTTP 502") ||
		strings.Contains(errStr, "HTTP 503") ||
		strings.Contains(errStr, "HTTP 504") ||
		strings.Contains(errStr, "HTTP 500") {
		return true
	}

	// 429 Too Many Requests is retryable (rate limiting)
	if strings.Contains(errStr, "HTTP 429") {
		return true
	}

	// Check if it's a RateLimitError
	if _, ok := err.(*RateLimitError); ok {
		return true
	}

	return false
}

// doSingleRequest performs a single HTTP request without retries
func (c *Client) doSingleRequest(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	var reqBodyLog []byte
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
		reqBodyLog = jsonBody
	}

	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-KEY", c.apiToken) // Note: Using X-API-KEY header, not Authorization Bearer

	// Log request details (excluding sensitive headers)
	tflog.Debug(ctx, "Pocket-ID API Request", map[string]interface{}{
		"method":   method,
		"url":      url,
		"endpoint": endpoint,
		"headers": map[string]string{
			"Content-Type": req.Header.Get("Content-Type"),
			"Accept":       req.Header.Get("Accept"),
			"X-API-KEY":    "[REDACTED]",
		},
	})

	if reqBodyLog != nil {
		tflog.Trace(ctx, "Request Body", map[string]interface{}{
			"body": string(reqBodyLog),
		})
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		tflog.Error(ctx, "HTTP Request Failed", map[string]interface{}{
			"error": err.Error(),
			"url":   url,
		})
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			tflog.Warn(ctx, "Failed to close response body", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Log response details
	tflog.Debug(ctx, "Pocket-ID API Response", map[string]interface{}{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
		"url":         url,
	})

	tflog.Trace(ctx, "Response Body", map[string]interface{}{
		"body": string(respBody),
	})

	// Check for errors
	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			tflog.Error(ctx, "API Error Response", map[string]interface{}{
				"status_code": resp.StatusCode,
				"raw_body":    string(respBody),
			})
			// Handle rate limit errors with Retry-After header
			if resp.StatusCode == 429 {
				retryAfter := resp.Header.Get("Retry-After")
				return nil, &RateLimitError{
					StatusCode: resp.StatusCode,
					Message:    string(respBody),
					RetryAfter: retryAfter,
				}
			}
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
		}
		tflog.Error(ctx, "API Error", map[string]interface{}{
			"status_code": resp.StatusCode,
			"error":       errResp.Error,
		})
		// Handle rate limit errors with Retry-After header
		if resp.StatusCode == 429 {
			retryAfter := resp.Header.Get("Retry-After")
			return nil, &RateLimitError{
				StatusCode: resp.StatusCode,
				Message:    errResp.Error,
				RetryAfter: retryAfter,
			}
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, errResp.Error)
	}

	return respBody, nil
}

// parseRetryAfter parses the Retry-After header value
// It can be either a delay in seconds or an HTTP-date
func parseRetryAfter(retryAfter string) int {
	// First try to parse as integer seconds
	if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds > 0 {
		return seconds
	}

	// Try to parse as HTTP-date
	if t, err := http.ParseTime(retryAfter); err == nil {
		delay := time.Until(t).Seconds()
		if delay > 0 {
			return int(delay)
		}
	}

	// Default to 60 seconds if we can't parse
	return 60
}

// OIDC Client methods

// CreateClient creates a new OIDC client
func (c *Client) CreateClient(createReq *OIDCClientCreateRequest) (*OIDCClient, error) {
	body, err := c.doRequest("POST", "/api/oidc/clients", createReq)
	if err != nil {
		return nil, err
	}

	var result OIDCClient
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &result, nil
}

// GetClient retrieves an OIDC client by ID
func (c *Client) GetClient(clientID string) (*OIDCClient, error) {
	body, err := c.doRequest("GET", fmt.Sprintf("/api/oidc/clients/%s", clientID), nil)
	if err != nil {
		return nil, err
	}

	var result OIDCClient
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &result, nil
}

// UpdateClient updates an existing OIDC client
func (c *Client) UpdateClient(clientID string, updateReq *OIDCClientCreateRequest) (*OIDCClient, error) {
	body, err := c.doRequest("PUT", fmt.Sprintf("/api/oidc/clients/%s", clientID), updateReq)
	if err != nil {
		return nil, err
	}

	var result OIDCClient
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &result, nil
}

// DeleteClient deletes an OIDC client
func (c *Client) DeleteClient(clientID string) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("/api/oidc/clients/%s", clientID), nil)
	return err
}

// ListClients retrieves all OIDC clients
func (c *Client) ListClients() (*PaginatedResponse[OIDCClient], error) {
	body, err := c.doRequest("GET", "/api/oidc/clients", nil)
	if err != nil {
		return nil, err
	}

	var result PaginatedResponse[OIDCClient]
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &result, nil
}

// UpdateClientAllowedUserGroups updates the allowed user groups for an OIDC client
func (c *Client) UpdateClientAllowedUserGroups(clientID string, groupIDs []string) error {
	req := UpdateAllowedUserGroupsRequest{UserGroupIDs: groupIDs}
	_, err := c.doRequest("PUT", fmt.Sprintf("/api/oidc/clients/%s/allowed-user-groups", clientID), req)
	return err
}

// GenerateClientSecret generates a new client secret for an OIDC client
func (c *Client) GenerateClientSecret(clientID string) (string, error) {
	body, err := c.doRequest("POST", fmt.Sprintf("/api/oidc/clients/%s/secret", clientID), nil)
	if err != nil {
		return "", err
	}

	var result struct {
		Secret string `json:"secret"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error unmarshaling response: %w", err)
	}

	return result.Secret, nil
}

// User methods

// CreateUser creates a new user
func (c *Client) CreateUser(user *UserCreateRequest) (*User, error) {
	body, err := c.doRequest("POST", "/api/users", user)
	if err != nil {
		return nil, err
	}

	var result User
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &result, nil
}

// GetUser retrieves a user by ID
func (c *Client) GetUser(userID string) (*User, error) {
	body, err := c.doRequest("GET", fmt.Sprintf("/api/users/%s", userID), nil)
	if err != nil {
		return nil, err
	}

	var result User
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &result, nil
}

// UpdateUser updates an existing user
func (c *Client) UpdateUser(userID string, user *UserCreateRequest) (*User, error) {
	body, err := c.doRequest("PUT", fmt.Sprintf("/api/users/%s", userID), user)
	if err != nil {
		return nil, err
	}

	var result User
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &result, nil
}

// DeleteUser deletes a user
func (c *Client) DeleteUser(userID string) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("/api/users/%s", userID), nil)
	return err
}

// ListUsers retrieves all users
func (c *Client) ListUsers() (*PaginatedResponse[User], error) {
	body, err := c.doRequest("GET", "/api/users", nil)
	if err != nil {
		return nil, err
	}

	var result PaginatedResponse[User]
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &result, nil
}

// UpdateUserGroups updates the groups a user belongs to
func (c *Client) UpdateUserGroups(userID string, groupIDs []string) error {
	// Ensure groupIDs is never nil to serialize as empty array instead of null
	if groupIDs == nil {
		groupIDs = []string{}
	}
	req := UpdateUserGroupsRequest{UserGroupIDs: groupIDs}
	_, err := c.doRequest("PUT", fmt.Sprintf("/api/users/%s/user-groups", userID), req)
	return err
}

// User Group methods

// CreateUserGroup creates a new user group
func (c *Client) CreateUserGroup(group *UserGroupCreateRequest) (*UserGroup, error) {
	body, err := c.doRequest("POST", "/api/user-groups", group)
	if err != nil {
		return nil, err
	}

	var result UserGroup
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &result, nil
}

// GetUserGroup retrieves a user group by ID
func (c *Client) GetUserGroup(groupID string) (*UserGroup, error) {
	body, err := c.doRequest("GET", fmt.Sprintf("/api/user-groups/%s", groupID), nil)
	if err != nil {
		return nil, err
	}

	var result UserGroup
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &result, nil
}

// UpdateUserGroup updates an existing user group
func (c *Client) UpdateUserGroup(groupID string, group *UserGroupCreateRequest) (*UserGroup, error) {
	body, err := c.doRequest("PUT", fmt.Sprintf("/api/user-groups/%s", groupID), group)
	if err != nil {
		return nil, err
	}

	var result UserGroup
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &result, nil
}

// DeleteUserGroup deletes a user group
func (c *Client) DeleteUserGroup(groupID string) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("/api/user-groups/%s", groupID), nil)
	return err
}

// ListUserGroups retrieves all user groups
func (c *Client) ListUserGroups() (*PaginatedResponse[UserGroup], error) {
	body, err := c.doRequest("GET", "/api/user-groups", nil)
	if err != nil {
		return nil, err
	}

	var result PaginatedResponse[UserGroup]
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &result, nil
}

// OneTimeAccessToken represents a one-time access token
type OneTimeAccessToken struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// OneTimeAccessTokenRequest represents a request to create a one-time access token
type OneTimeAccessTokenRequest struct {
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// GetOneTimeAccessToken retrieves the one-time access token for a user
func (c *Client) GetOneTimeAccessToken(userID string) (*OneTimeAccessToken, error) {
	body, err := c.doRequest("GET", fmt.Sprintf("/api/users/%s/one-time-access-token", userID), nil)
	if err != nil {
		return nil, err
	}

	var token OneTimeAccessToken
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &token, nil
}

// CreateOneTimeAccessToken creates a new one-time access token for a user
func (c *Client) CreateOneTimeAccessToken(userID string, req *OneTimeAccessTokenRequest) (*OneTimeAccessToken, error) {
	body, err := c.doRequest("POST", fmt.Sprintf("/api/users/%s/one-time-access-token", userID), req)
	if err != nil {
		return nil, err
	}

	var token OneTimeAccessToken
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &token, nil
}

// DeleteOneTimeAccessToken deletes the one-time access token for a user
func (c *Client) DeleteOneTimeAccessToken(userID string) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("/api/users/%s/one-time-access-token", userID), nil)
	return err
}
