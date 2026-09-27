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
		{Key: "autoCreateOidcClientSecret", Type: "boolean", Value: "false"},
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
	assert.Equal(t, "false", cfg.AutoCreateOIDCClientSecret)
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

func TestClient_UpdateApplicationImage(t *testing.T) {
	image := []byte("\x89PNG\r\n\x1a\nnot-really-a-png")

	tests := []struct {
		image     client.ApplicationImage
		wantPath  string
		wantLight string
	}{
		{client.ApplicationImageLogo, "/api/application-images/logo", "true"},
		{client.ApplicationImageDarkLogo, "/api/application-images/logo", "false"},
		{client.ApplicationImageEmailLogo, "/api/application-images/email", ""},
		{client.ApplicationImageBackground, "/api/application-images/background", ""},
		{client.ApplicationImageFavicon, "/api/application-images/favicon", ""},
		{client.ApplicationImageDefaultProfilePicture, "/api/application-images/default-profile-picture", ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.image), func(t *testing.T) {
			called := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				assert.Equal(t, "PUT", r.Method)
				assert.Equal(t, tt.wantPath, r.URL.Path)
				assert.Equal(t, tt.wantLight, r.URL.Query().Get("light"))
				assert.Equal(t, "test-token", r.Header.Get("X-API-KEY"))

				filename, data := readUploadedFile(t, r)
				assert.Equal(t, "image.png", filename)
				assert.Equal(t, image, data)
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			c, err := client.NewClient(server.URL, "test-token", false, 30)
			require.NoError(t, err)

			require.NoError(t, c.UpdateApplicationImage(tt.image, "image.png", image))
			assert.True(t, called)
		})
	}
}

func TestClient_DeleteApplicationImage(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		wantErr     bool
		errContains string
	}{
		{"reset", http.StatusNoContent, false, ""},
		// No custom image is set.
		{"not set", http.StatusNotFound, false, ""},
		{"server error", http.StatusInternalServerError, true, "HTTP 500"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "DELETE", r.Method)
				assert.Equal(t, "/api/application-images/logo", r.URL.Path)
				assert.Equal(t, "false", r.URL.Query().Get("light"))
				w.WriteHeader(tt.status)
			}))
			defer server.Close()

			c, err := client.NewClient(server.URL, "test-token", false, 30)
			require.NoError(t, err)

			err = c.DeleteApplicationImage(client.ApplicationImageDarkLogo)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}
			assert.NoError(t, err)
		})
	}
}
