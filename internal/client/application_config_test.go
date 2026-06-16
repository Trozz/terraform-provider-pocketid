package client_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
)

func appConfigVariables() []client.AppConfigVariable {
	return []client.AppConfigVariable{
		{Key: "appName", Type: "string", Value: "My App"},
		{Key: "sessionDuration", Type: "number", Value: "60"},
		{Key: "homePageUrl", Type: "string", Value: "https://example.com"},
		{Key: "smtpPassword", Type: "string", Value: "s3cret"},
		{Key: "ldapEnabled", Type: "boolean", Value: "true"},
		{Key: "ldapBindPassword", Type: "string", Value: "ldapsecret"},
	}
}

func TestGetApplicationConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/application-configuration/all", r.URL.Path)
		assert.Equal(t, "test-token", r.Header.Get("X-API-KEY"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(appConfigVariables())
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	cfg, err := c.GetApplicationConfig()
	require.NoError(t, err)

	assert.Equal(t, "My App", cfg.AppName)
	assert.Equal(t, "60", cfg.SessionDuration)
	assert.Equal(t, "https://example.com", cfg.HomePageURL)
	assert.Equal(t, "s3cret", cfg.SmtpPassword)
	assert.Equal(t, "true", cfg.LdapEnabled)
	assert.Equal(t, "ldapsecret", cfg.LdapBindPassword)
	// Keys not present in the response stay at their zero value.
	assert.Equal(t, "", cfg.AccentColor)
}

func TestUpdateApplicationConfig(t *testing.T) {
	var receivedBody map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/api/application-configuration", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &receivedBody))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]client.AppConfigVariable{
			{Key: "appName", Type: "string", Value: "Updated App"},
			{Key: "smtpHost", Type: "string", Value: "smtp.example.com"},
		})
	}))
	defer server.Close()

	c, err := client.NewClient(server.URL, "test-token", false, 30)
	require.NoError(t, err)

	updated, err := c.UpdateApplicationConfig(&client.ApplicationConfig{
		AppName:  "Updated App",
		SmtpHost: "smtp.example.com",
	})
	require.NoError(t, err)

	// The PUT body should carry the JSON keys expected by Pocket-ID.
	assert.Equal(t, "Updated App", receivedBody["appName"])
	assert.Equal(t, "smtp.example.com", receivedBody["smtpHost"])

	// Response is parsed back from the key/value array.
	assert.Equal(t, "Updated App", updated.AppName)
	assert.Equal(t, "smtp.example.com", updated.SmtpHost)
}
