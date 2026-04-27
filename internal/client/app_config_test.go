package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
)

func TestGetLDAPConfig(t *testing.T) {
	// Mock API response - array of config variables
	configResponse := []map[string]any{
		{"key": "ldapEnabled", "type": "boolean", "value": true},
		{"key": "ldapUrl", "type": "string", "value": "ldaps://ldap.example.com:636"},
		{"key": "ldapBindDn", "type": "string", "value": "cn=admin,dc=example,dc=com"},
		{"key": "ldapBase", "type": "string", "value": "dc=example,dc=com"},
		{"key": "ldapSkipCertVerify", "type": "boolean", "value": false},
		{"key": "ldapUserSearchFilter", "type": "string", "value": "(objectClass=person)"},
		{"key": "ldapGroupSearchFilter", "type": "string", "value": "(objectClass=groupOfNames)"},
		{"key": "ldapAttributeUserUniqueIdentifier", "type": "string", "value": "objectGUID"},
		{"key": "ldapAttributeUserUsername", "type": "string", "value": "sAMAccountName"},
		{"key": "ldapAttributeUserEmail", "type": "string", "value": "mail"},
		{"key": "ldapAttributeUserFirstName", "type": "string", "value": "givenName"},
		{"key": "ldapAttributeUserLastName", "type": "string", "value": "sn"},
		{"key": "ldapAttributeGroupMember", "type": "string", "value": "member"},
		{"key": "ldapAttributeGroupUniqueIdentifier", "type": "string", "value": "objectGUID"},
		{"key": "ldapAttributeGroupName", "type": "string", "value": "cn"},
		{"key": "ldapAttributeAdminGroup", "type": "string", "value": "PocketID-Admins"},
		{"key": "ldapSoftDeleteUsers", "type": "boolean", "value": true},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/application-configuration/all", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(configResponse); err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	config, err := c.GetLDAPConfig()
	require.NoError(t, err)

	assert.True(t, config.Enabled)
	assert.Equal(t, "ldaps://ldap.example.com:636", config.URL)
	assert.Equal(t, "cn=admin,dc=example,dc=com", config.BindDN)
	assert.Equal(t, "dc=example,dc=com", config.BaseDN)
	assert.Equal(t, "objectGUID", config.UserUniqueAttribute)
	assert.Equal(t, "sAMAccountName", config.UserUsernameAttribute)
	assert.True(t, config.SoftDeleteUsers)
}

func TestUpdateLDAPConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/application-configuration", r.URL.Path)
		assert.Equal(t, "PUT", r.Method)

		var req client.LDAPConfigUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.True(t, req.LdapEnabled)
		assert.Equal(t, "ldaps://ldap.example.com:636", req.LdapUrl)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	updateReq := &client.LDAPConfigUpdateRequest{
		LdapEnabled: true,
		LdapUrl:     "ldaps://ldap.example.com:636",
		LdapBindDn:  "cn=admin,dc=example,dc=com",
		LdapBase:    "dc=example,dc=com",
	}

	err = c.UpdateLDAPConfig(updateReq)
	assert.NoError(t, err)
}

func TestSyncLDAP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/application-configuration/sync-ldap", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	err = c.SyncLDAP()
	assert.NoError(t, err)
}
