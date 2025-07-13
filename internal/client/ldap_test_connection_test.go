package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
)

func TestTestLDAPConnection(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse interface{}
		serverStatus   int
		expectedResult *client.LDAPTestResult
		expectedError  bool
	}{
		{
			name: "successful connection test",
			serverResponse: client.LDAPTestResult{
				ConnectionSuccess: true,
				BindSuccess:       true,
				UserCount:         50,
				GroupCount:        5,
				SampleUsers:       []string{"user1", "user2", "user3"},
				SampleGroups:      []string{"group1", "group2"},
			},
			serverStatus: http.StatusOK,
			expectedResult: &client.LDAPTestResult{
				ConnectionSuccess: true,
				BindSuccess:       true,
				UserCount:         50,
				GroupCount:        5,
				SampleUsers:       []string{"user1", "user2", "user3"},
				SampleGroups:      []string{"group1", "group2"},
			},
			expectedError: false,
		},
		{
			name: "connection failed",
			serverResponse: client.LDAPTestResult{
				ConnectionSuccess: false,
				BindSuccess:       false,
				Error:             "Connection refused",
			},
			serverStatus: http.StatusOK,
			expectedResult: &client.LDAPTestResult{
				ConnectionSuccess: false,
				BindSuccess:       false,
				Error:             "Connection refused",
			},
			expectedError: false,
		},
		{
			name: "bind failed",
			serverResponse: client.LDAPTestResult{
				ConnectionSuccess: true,
				BindSuccess:       false,
				Error:             "Invalid credentials",
			},
			serverStatus: http.StatusOK,
			expectedResult: &client.LDAPTestResult{
				ConnectionSuccess: true,
				BindSuccess:       false,
				Error:             "Invalid credentials",
			},
			expectedError: false,
		},
		{
			name:          "server error",
			serverStatus:  http.StatusInternalServerError,
			expectedError: true,
		},
		{
			name:          "unauthorized",
			serverStatus:  http.StatusUnauthorized,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/application-configuration/test-ldap", r.URL.Path)
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "test-api-key", r.Header.Get("X-API-KEY"))

				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					err := json.NewEncoder(w).Encode(tt.serverResponse)
					require.NoError(t, err)
				}
			}))
			defer server.Close()

			client, err := client.NewClient(server.URL, "test-api-key", false, 30)
			require.NoError(t, err)

			result, err := client.TestLDAPConnection()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestTestLDAPConnectionWithContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := client.LDAPTestResult{
			ConnectionSuccess: true,
			BindSuccess:       true,
			UserCount:         10,
			GroupCount:        2,
		}
		err := json.NewEncoder(w).Encode(result)
		require.NoError(t, err)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-api-key", false, 30)
	require.NoError(t, err)

	ctx := context.Background()
	result, err := c.TestLDAPConnectionWithContext(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.ConnectionSuccess)
	assert.True(t, result.BindSuccess)
	assert.Equal(t, 10, result.UserCount)
	assert.Equal(t, 2, result.GroupCount)
}

func TestTestLDAPConnection_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow response
		<-r.Context().Done()
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-api-key", false, 30)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = c.TestLDAPConnectionWithContext(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}
