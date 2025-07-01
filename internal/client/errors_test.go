package client_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
)

func TestClient_CreateClient_UnmarshalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"invalid json":}`)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	createReq := &client.OIDCClientCreateRequest{
		Name:         "test-client",
		CallbackURLs: []string{"https://example.com/callback"},
		IsPublic:     false,
		PkceEnabled:  true,
	}

	result, err := c.CreateClient(createReq)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error unmarshaling response")
}

func TestClient_UpdateClient_UnmarshalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"invalid json":}`)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	updateReq := &client.OIDCClientCreateRequest{
		Name:         "updated-client",
		CallbackURLs: []string{"https://example.com/callback"},
		IsPublic:     false,
		PkceEnabled:  true,
	}

	result, err := c.UpdateClient("test-id", updateReq)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error unmarshaling response")
}

func TestClient_ListClients_UnmarshalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"data": "should be array"}`)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	result, err := c.ListClients()
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error unmarshaling response")
}

func TestClient_GenerateClientSecret_UnmarshalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"secret": 123}`) // secret should be string
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	_, err = c.GenerateClientSecret("test-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error unmarshaling response")
}

func TestClient_CreateUser_UnmarshalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"username": 123}`) // username should be string
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	createReq := &client.UserCreateRequest{
		Username: "testuser",
		Email:    "test@example.com",
	}

	result, err := c.CreateUser(createReq)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error unmarshaling response")
}

func TestClient_UpdateUser_UnmarshalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"email": []}`) // email should be string
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	updateReq := &client.UserCreateRequest{
		Username: "testuser",
		Email:    "test@example.com",
	}

	result, err := c.UpdateUser("test-id", updateReq)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error unmarshaling response")
}

func TestClient_ListUsers_UnmarshalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"data": "not an array", "pagination": {}}`)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	result, err := c.ListUsers()
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error unmarshaling response")
}

func TestClient_CreateUserGroup_UnmarshalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"name": true}`) // name should be string
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	createReq := &client.UserGroupCreateRequest{
		Name:         "test-group",
		FriendlyName: "Test Group",
	}

	result, err := c.CreateUserGroup(createReq)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error unmarshaling response")
}

func TestClient_UpdateUserGroup_UnmarshalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"friendlyName": 123}`) // friendlyName should be string
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	updateReq := &client.UserGroupCreateRequest{
		Name:         "test-group",
		FriendlyName: "Test Group",
	}

	result, err := c.UpdateUserGroup("test-id", updateReq)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error unmarshaling response")
}

func TestClient_GetUser_UnmarshalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id": [], "username": "test"}`) // id should be string
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	result, err := c.GetUser("test-id")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error unmarshaling response")
}

func TestClient_GetUserGroup_UnmarshalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id": {}, "name": "test"}`) // id should be string
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	result, err := c.GetUserGroup("test-id")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error unmarshaling response")
}

func TestClient_ListUserGroups_UnmarshalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"data": {}, "pagination": "invalid"}`)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	result, err := c.ListUserGroups()
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error unmarshaling response")
}

// Test rate limiting with Retry-After header (numeric seconds)
func TestClient_RateLimitWithRetryAfterSeconds(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.Header().Set("Retry-After", "2") // 2 seconds
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, `{"error": "Rate limit exceeded"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id": "test-id", "name": "Test Client"}`)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	start := time.Now()
	result, err := c.GetClient("test-id")
	elapsed := time.Since(start)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, attempts)
	// Should wait approximately 2 seconds
	assert.True(t, elapsed >= 1900*time.Millisecond && elapsed <= 2500*time.Millisecond,
		"Expected wait time around 2 seconds, got %v", elapsed)
}

// Test timeout handling
func TestClient_RequestTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep longer than client timeout
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id": "test-id"}`)
	}))
	defer server.Close()

	// Create client with 1 second timeout
	c, err := client.NewClient(server.URL, "test-token", false, 1)
	require.NoError(t, err)

	start := time.Now()
	result, err := c.GetClient("test-id")
	elapsed := time.Since(start)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "deadline exceeded")
	// Should timeout after approximately 1 second
	assert.True(t, elapsed >= 900*time.Millisecond && elapsed <= 1500*time.Millisecond,
		"Expected timeout around 1 second, got %v", elapsed)
}

// Test error response with empty body
func TestClient_ErrorResponseEmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		// No body
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	result, err := c.GetClient("test-id")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "HTTP 500")
}

// Test with non-retryable errors
func TestClient_NonRetryableErrors(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
		attempts   int
	}{
		{
			name:       "400 Bad Request",
			statusCode: http.StatusBadRequest,
			attempts:   1,
		},
		{
			name:       "401 Unauthorized",
			statusCode: http.StatusUnauthorized,
			attempts:   1,
		},
		{
			name:       "403 Forbidden",
			statusCode: http.StatusForbidden,
			attempts:   1,
		},
		{
			name:       "404 Not Found",
			statusCode: http.StatusNotFound,
			attempts:   1,
		},
		{
			name:       "409 Conflict",
			statusCode: http.StatusConflict,
			attempts:   1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			attempts := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attempts++
				w.WriteHeader(tc.statusCode)
				fmt.Fprintf(w, `{"error": "Error %d"}`, tc.statusCode)
			}))
			defer server.Close()

			c, err := client.NewClient(server.URL, "test-token", false, 30)
			require.NoError(t, err)

			result, err := c.GetClient("test-id")
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), fmt.Sprintf("HTTP %d", tc.statusCode))
			assert.Equal(t, tc.attempts, attempts, "Should not retry for %d errors", tc.statusCode)
		})
	}
}
