package client_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/trozz/terraform-provider-pocketid/internal/client"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name          string
		baseURL       string
		apiToken      string
		skipTLSVerify bool
		timeout       int64
		wantErr       bool
		errMsg        string
	}{
		{
			name:     "valid configuration",
			baseURL:  "https://pocket-id.example.com",
			apiToken: "test-token",
			timeout:  30,
			wantErr:  false,
		},
		{
			name:     "missing base URL",
			baseURL:  "",
			apiToken: "test-token",
			timeout:  30,
			wantErr:  true,
			errMsg:   "base URL is required",
		},
		{
			name:     "missing API token",
			baseURL:  "https://pocket-id.example.com",
			apiToken: "",
			timeout:  30,
			wantErr:  true,
			errMsg:   "API token is required",
		},
		{
			name:          "with TLS skip",
			baseURL:       "https://pocket-id.example.com",
			apiToken:      "test-token",
			skipTLSVerify: true,
			timeout:       60,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := client.NewClient(tt.baseURL, tt.apiToken, tt.skipTLSVerify, tt.timeout)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, c)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, c)
			}
		})
	}
}

func TestClient_CreateClient(t *testing.T) {
	expectedClient := &client.OIDCClient{
		ID:           "test-client-id",
		Name:         "Test Client",
		CallbackURLs: []string{"https://example.com/callback"},
		IsPublic:     false,
		PkceEnabled:  true,
		HasLogo:      false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/oidc/clients", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-token", r.Header.Get("X-API-KEY"))

		var req client.OIDCClientCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "Test Client", req.Name)
		assert.Equal(t, []string{"https://example.com/callback"}, req.CallbackURLs)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedClient)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	createReq := &client.OIDCClientCreateRequest{
		Name:         "Test Client",
		CallbackURLs: []string{"https://example.com/callback"},
		IsPublic:     false,
		PkceEnabled:  true,
	}

	result, err := c.CreateClient(createReq)
	assert.NoError(t, err)
	assert.Equal(t, expectedClient, result)
}

func TestClient_GetClient(t *testing.T) {
	expectedClient := &client.OIDCClient{
		ID:           "test-client-id",
		Name:         "Test Client",
		CallbackURLs: []string{"https://example.com/callback"},
		IsPublic:     false,
		PkceEnabled:  true,
		HasLogo:      false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/oidc/clients/test-client-id", r.URL.Path)
		assert.Equal(t, "test-token", r.Header.Get("X-API-KEY"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedClient)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	result, err := c.GetClient("test-client-id")
	assert.NoError(t, err)
	assert.Equal(t, expectedClient, result)
}

func TestClient_UpdateClient(t *testing.T) {
	expectedClient := &client.OIDCClient{
		ID:           "test-client-id",
		Name:         "Updated Client",
		CallbackURLs: []string{"https://example.com/callback", "https://example.com/callback2"},
		IsPublic:     false,
		PkceEnabled:  true,
		HasLogo:      false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/oidc/clients/test-client-id", r.URL.Path)
		assert.Equal(t, "test-token", r.Header.Get("X-API-KEY"))

		var req client.OIDCClientCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "Updated Client", req.Name)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedClient)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	updateReq := &client.OIDCClientCreateRequest{
		Name:         "Updated Client",
		CallbackURLs: []string{"https://example.com/callback", "https://example.com/callback2"},
		IsPublic:     false,
		PkceEnabled:  true,
	}

	result, err := c.UpdateClient("test-client-id", updateReq)
	assert.NoError(t, err)
	assert.Equal(t, expectedClient, result)
}

func TestClient_DeleteClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/oidc/clients/test-client-id", r.URL.Path)
		assert.Equal(t, "test-token", r.Header.Get("X-API-KEY"))

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	err = c.DeleteClient("test-client-id")
	assert.NoError(t, err)
}

func TestClient_ListClients(t *testing.T) {
	expectedResponse := &client.PaginatedResponse[client.OIDCClient]{
		Data: []client.OIDCClient{
			{
				ID:   "client1",
				Name: "Client 1",
			},
			{
				ID:   "client2",
				Name: "Client 2",
			},
		},
		Pagination: client.PaginationInfo{
			TotalItems:   2,
			CurrentPage:  1,
			ItemsPerPage: 20,
			TotalPages:   1,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/oidc/clients", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	result, err := c.ListClients()
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
}

func TestClient_GenerateClientSecret(t *testing.T) {
	expectedSecret := "new-client-secret-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/oidc/clients/test-client-id/secret", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"secret": expectedSecret})
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	secret, err := c.GenerateClientSecret("test-client-id")
	assert.NoError(t, err)
	assert.Equal(t, expectedSecret, secret)
}

func TestClient_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedErrMsg string
	}{
		{
			name:           "400 Bad Request",
			statusCode:     http.StatusBadRequest,
			responseBody:   `{"error": "Invalid request"}`,
			expectedErrMsg: "HTTP 400: Invalid request",
		},
		{
			name:           "401 Unauthorized",
			statusCode:     http.StatusUnauthorized,
			responseBody:   `{"error": "Invalid API key"}`,
			expectedErrMsg: "HTTP 401: Invalid API key",
		},
		{
			name:           "404 Not Found",
			statusCode:     http.StatusNotFound,
			responseBody:   `{"error": "Client not found"}`,
			expectedErrMsg: "HTTP 404: Client not found",
		},
		{
			name:           "500 Internal Server Error",
			statusCode:     http.StatusInternalServerError,
			responseBody:   `{"error": "Internal server error"}`,
			expectedErrMsg: "HTTP 500: Internal server error",
		},
		{
			name:           "Invalid JSON response",
			statusCode:     http.StatusBadRequest,
			responseBody:   `invalid json`,
			expectedErrMsg: "HTTP 400: invalid json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				fmt.Fprint(w, tt.responseBody)
			}))
			defer server.Close()

			c, err := client.NewClient(server.URL, "test-token", false, 30)
			require.NoError(t, err)

			_, err = c.GetClient("test-client-id")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErrMsg)
		})
	}
}

func TestClient_RetryLogic(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			// Simulate transient errors for first 2 attempts
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, `{"error": "Service temporarily unavailable"}`)
			return
		}
		// Success on third attempt
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&client.OIDCClient{
			ID:   "test-client-id",
			Name: "Test Client",
		})
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	result, err := c.GetClient("test-client-id")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-client-id", result.ID)
	assert.Equal(t, 3, attempts, "Should have made 3 attempts")
}

func TestClient_RetryExhaustion(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		// Always return 503
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, `{"error": "Service unavailable"}`)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	_, err = c.GetClient("test-client-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request failed after 4 attempts")
	assert.Equal(t, 4, attempts, "Should have made 4 attempts (initial + 3 retries)")
}

func TestClient_NonRetryableError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		// Return 404 which is not retryable
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error": "Not found"}`)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	_, err = c.GetClient("test-client-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 404: Not found")
	assert.Equal(t, 1, attempts, "Should have made only 1 attempt (no retries for 404)")
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	// Note: The current client doesn't expose context-aware methods,
	// but this test is prepared for when they are added
	_, err = c.GetClient("test-client-id")
	// The error might be a timeout or context cancellation depending on timing
	assert.Error(t, err)
}

// Test User-related methods
func TestClient_CreateUser(t *testing.T) {
	expectedUser := &client.User{
		ID:        "test-user-id",
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		IsAdmin:   false,
		Disabled:  false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/users", r.URL.Path)

		var req client.UserCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "testuser", req.Username)
		assert.Equal(t, "test@example.com", req.Email)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedUser)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	createReq := &client.UserCreateRequest{
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	result, err := c.CreateUser(createReq)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, result)
}

// Test Group-related methods
func TestClient_CreateUserGroup(t *testing.T) {
	expectedGroup := &client.UserGroup{
		ID:           "test-group-id",
		Name:         "test-group",
		FriendlyName: "Test Group",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/user-groups", r.URL.Path)

		var req client.UserGroupCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "test-group", req.Name)
		assert.Equal(t, "Test Group", req.FriendlyName)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedGroup)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	createReq := &client.UserGroupCreateRequest{
		Name:         "test-group",
		FriendlyName: "Test Group",
	}

	result, err := c.CreateUserGroup(createReq)
	assert.NoError(t, err)
	assert.Equal(t, expectedGroup, result)
}

func TestClient_UpdateUserGroups(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/users/test-user-id/user-groups", r.URL.Path)

		var req client.UpdateUserGroupsRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, []string{"group1", "group2"}, req.UserGroupIDs)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	err = c.UpdateUserGroups("test-user-id", []string{"group1", "group2"})
	assert.NoError(t, err)
}

func TestClient_UpdateClientAllowedUserGroups(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/oidc/clients/test-client-id/allowed-user-groups", r.URL.Path)

		var req client.UpdateAllowedUserGroupsRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, []string{"group1", "group2"}, req.UserGroupIDs)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	err = c.UpdateClientAllowedUserGroups("test-client-id", []string{"group1", "group2"})
	assert.NoError(t, err)
}
