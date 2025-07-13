package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetApplicationConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/application-configuration/all", r.URL.Path)
		assert.Equal(t, "test-key", r.Header.Get("X-API-KEY"))

		response := ApplicationConfiguration{
			LDAP: &LDAPConfiguration{
				Enabled:               "true",
				URL:                   "ldaps://ldap.example.com:636",
				BindDN:                "cn=admin,dc=example,dc=com",
				BindPassword:          "secret",
				BaseDN:                "dc=example,dc=com",
				SkipCertVerify:        "false",
				UserSearchFilter:      "(objectClass=person)",
				UserGroupSearchFilter: "(objectClass=groupOfNames)",
				UserAttributes: &LDAPUserAttributes{
					UniqueIdentifier: "objectGUID",
					Username:         "sAMAccountName",
					Email:            "mail",
					FirstName:        "givenName",
					LastName:         "sn",
					ProfilePicture:   "thumbnailPhoto",
				},
				GroupAttributes: &LDAPGroupAttributes{
					Member:           "member",
					UniqueIdentifier: "objectGUID",
					Name:             "cn",
					AdminGroupName:   "PocketID-Admins",
				},
				SoftDeleteUsers: "true",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test-key", false, 30)
	require.NoError(t, err)

	config, err := client.GetApplicationConfiguration()
	require.NoError(t, err)
	require.NotNil(t, config)
	require.NotNil(t, config.LDAP)

	assert.Equal(t, "true", config.LDAP.Enabled)
	assert.Equal(t, "ldaps://ldap.example.com:636", config.LDAP.URL)
	assert.Equal(t, "cn=admin,dc=example,dc=com", config.LDAP.BindDN)
	assert.Equal(t, "secret", config.LDAP.BindPassword)
	assert.Equal(t, "dc=example,dc=com", config.LDAP.BaseDN)
	assert.Equal(t, "false", config.LDAP.SkipCertVerify)
	assert.Equal(t, "(objectClass=person)", config.LDAP.UserSearchFilter)
	assert.Equal(t, "(objectClass=groupOfNames)", config.LDAP.UserGroupSearchFilter)
	assert.Equal(t, "true", config.LDAP.SoftDeleteUsers)

	assert.NotNil(t, config.LDAP.UserAttributes)
	assert.Equal(t, "objectGUID", config.LDAP.UserAttributes.UniqueIdentifier)
	assert.Equal(t, "sAMAccountName", config.LDAP.UserAttributes.Username)
	assert.Equal(t, "mail", config.LDAP.UserAttributes.Email)
	assert.Equal(t, "givenName", config.LDAP.UserAttributes.FirstName)
	assert.Equal(t, "sn", config.LDAP.UserAttributes.LastName)
	assert.Equal(t, "thumbnailPhoto", config.LDAP.UserAttributes.ProfilePicture)

	assert.NotNil(t, config.LDAP.GroupAttributes)
	assert.Equal(t, "member", config.LDAP.GroupAttributes.Member)
	assert.Equal(t, "objectGUID", config.LDAP.GroupAttributes.UniqueIdentifier)
	assert.Equal(t, "cn", config.LDAP.GroupAttributes.Name)
	assert.Equal(t, "PocketID-Admins", config.LDAP.GroupAttributes.AdminGroupName)
}

func TestGetApplicationConfiguration_NoLDAP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/application-configuration/all", r.URL.Path)

		response := ApplicationConfiguration{
			LDAP: nil,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test-key", false, 30)
	require.NoError(t, err)

	config, err := client.GetApplicationConfiguration()
	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Nil(t, config.LDAP)
}

func TestUpdateApplicationConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/api/application-configuration", r.URL.Path)
		assert.Equal(t, "test-key", r.Header.Get("X-API-KEY"))

		var req ApplicationConfigurationUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.NotNil(t, req.LDAP)
		assert.Equal(t, "true", req.LDAP.Enabled)
		assert.Equal(t, "ldaps://ldap.example.com:636", req.LDAP.URL)

		// Return the updated configuration
		response := ApplicationConfiguration(req)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		jsonErr := json.NewEncoder(w).Encode(response)
		require.NoError(t, jsonErr)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test-key", false, 30)
	require.NoError(t, err)

	updateReq := &ApplicationConfigurationUpdateRequest{
		LDAP: &LDAPConfiguration{
			Enabled:        "true",
			URL:            "ldaps://ldap.example.com:636",
			BindDN:         "cn=admin,dc=example,dc=com",
			BindPassword:   "secret",
			BaseDN:         "dc=example,dc=com",
			SkipCertVerify: "false",
		},
	}

	config, err := client.UpdateApplicationConfiguration(updateReq)
	require.NoError(t, err)
	require.NotNil(t, config)
	require.NotNil(t, config.LDAP)

	assert.Equal(t, "true", config.LDAP.Enabled)
	assert.Equal(t, "ldaps://ldap.example.com:636", config.LDAP.URL)
}

func TestTriggerLDAPSync(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/application-configuration/sync-ldap", r.URL.Path)
		assert.Equal(t, "test-key", r.Header.Get("X-API-KEY"))

		response := LDAPSyncResponse{
			Status:       "completed",
			StartTime:    "2024-01-15T10:29:00Z",
			EndTime:      "2024-01-15T10:30:00Z",
			UsersAdded:   10,
			UsersUpdated: 5,
			GroupsAdded:  2,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test-key", false, 30)
	require.NoError(t, err)

	syncResp, err := client.TriggerLDAPSync()
	require.NoError(t, err)
	require.NotNil(t, syncResp)

	assert.Equal(t, "completed", syncResp.Status)
	assert.Equal(t, "2024-01-15T10:29:00Z", syncResp.StartTime)
	assert.Equal(t, "2024-01-15T10:30:00Z", syncResp.EndTime)
	assert.Equal(t, 10, syncResp.UsersAdded)
	assert.Equal(t, 5, syncResp.UsersUpdated)
	assert.Equal(t, 2, syncResp.GroupsAdded)
}

func TestTriggerLDAPSync_Failed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/application-configuration/sync-ldap", r.URL.Path)

		response := LDAPSyncResponse{
			Status:    "failed",
			Error:     "Failed to connect to LDAP server",
			StartTime: "2024-01-15T10:30:00Z",
			EndTime:   "2024-01-15T10:30:01Z",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test-key", false, 30)
	require.NoError(t, err)

	syncResp, err := client.TriggerLDAPSync()
	require.NoError(t, err)
	require.NotNil(t, syncResp)

	assert.Equal(t, "failed", syncResp.Status)
	assert.Equal(t, "Failed to connect to LDAP server", syncResp.Error)
}

func TestApplicationConfiguration_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		errorResponse ErrorResponse
		expectedError string
	}{
		{
			name:       "401 Unauthorized",
			statusCode: http.StatusUnauthorized,
			errorResponse: ErrorResponse{
				Error: "Unauthorized",
			},
			expectedError: "HTTP 401: Unauthorized",
		},
		{
			name:       "403 Forbidden",
			statusCode: http.StatusForbidden,
			errorResponse: ErrorResponse{
				Error: "Forbidden - insufficient permissions",
			},
			expectedError: "HTTP 403: Forbidden - insufficient permissions",
		},
		{
			name:       "404 Not Found",
			statusCode: http.StatusNotFound,
			errorResponse: ErrorResponse{
				Error: "Configuration not found",
			},
			expectedError: "HTTP 404: Configuration not found",
		},
		{
			name:       "500 Internal Server Error",
			statusCode: http.StatusInternalServerError,
			errorResponse: ErrorResponse{
				Error: "Internal server error",
			},
			expectedError: "HTTP 500: Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				err := json.NewEncoder(w).Encode(tt.errorResponse)
				require.NoError(t, err)
			}))
			defer server.Close()

			client, err := NewClient(server.URL, "test-key", false, 30)
			require.NoError(t, err)

			_, err = client.GetApplicationConfiguration()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}
