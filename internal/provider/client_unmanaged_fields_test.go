//go:build acc
// +build acc

package provider_test

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
)

const (
	canaryDescription      = "set-outside-terraform"
	canaryAccessTokenMins  = int64(120)
	canaryRefreshTokenMins = int64(4320)
)

func testClient() (*client.Client, error) {
	return client.NewClient(
		os.Getenv("POCKETID_BASE_URL"),
		os.Getenv("POCKETID_API_TOKEN"),
		false,
		30,
	)
}

// setUnmanagedClientFields sets, directly through the API, the client settings
// the provider does not expose as attributes.
func setUnmanagedClientFields(id string) error {
	c, err := testClient()
	if err != nil {
		return err
	}

	current, err := c.GetClient(id)
	if err != nil {
		return err
	}

	req := &client.OIDCClientCreateRequest{
		Name:                        current.Name,
		CallbackURLs:                current.CallbackURLs,
		LogoutCallbackURLs:          current.LogoutCallbackURLs,
		IsPublic:                    current.IsPublic,
		PkceEnabled:                 current.PkceEnabled,
		IsGroupRestricted:           current.IsGroupRestricted,
		Credentials:                 current.Credentials,
		Description:                 canaryDescription,
		SkipConsent:                 true,
		AccessTokenDurationMinutes:  canaryAccessTokenMins,
		RefreshTokenDurationMinutes: canaryRefreshTokenMins,
	}

	_, err = c.UpdateClient(id, req)
	return err
}

// checkUnmanagedClientFieldsPreserved verifies the values seeded by
// setUnmanagedClientFields survived a provider-driven update.
func checkUnmanagedClientFieldsPreserved(id string) error {
	c, err := testClient()
	if err != nil {
		return err
	}

	got, err := c.GetClient(id)
	if err != nil {
		return err
	}

	if got.Description != canaryDescription {
		return fmt.Errorf("description was reset: got %q, want %q", got.Description, canaryDescription)
	}
	if !got.SkipConsent {
		return fmt.Errorf("skipConsent was reset to false")
	}
	if got.AccessTokenDurationMinutes != canaryAccessTokenMins {
		return fmt.Errorf("accessTokenDurationMinutes was reset: got %d, want %d",
			got.AccessTokenDurationMinutes, canaryAccessTokenMins)
	}
	if got.RefreshTokenDurationMinutes != canaryRefreshTokenMins {
		return fmt.Errorf("refreshTokenDurationMinutes was reset: got %d, want %d",
			got.RefreshTokenDurationMinutes, canaryRefreshTokenMins)
	}
	return nil
}

// checkClientSecretCount asserts how many secrets the client currently has.
func checkClientSecretCount(clientID string, want int) error {
	c, err := testClient()
	if err != nil {
		return err
	}

	secrets, err := c.ListClientSecrets(clientID)
	if err != nil {
		return err
	}

	if len(secrets) != want {
		ids := make([]string, 0, len(secrets))
		for _, s := range secrets {
			ids = append(ids, s.ID)
		}
		return fmt.Errorf("expected %d client secret(s), got %d: %v", want, len(secrets), ids)
	}
	return nil
}

// checkClientSecretAbsent asserts a specific secret has been revoked.
func checkClientSecretAbsent(clientID, secretID string) error {
	c, err := testClient()
	if err != nil {
		return err
	}

	secrets, err := c.ListClientSecrets(clientID)
	if err != nil {
		return err
	}

	for _, s := range secrets {
		if s.ID == secretID {
			return fmt.Errorf("superseded secret %s is still present (active=%v)", secretID, s.IsActive)
		}
	}
	return nil
}

// Public images Pocket ID can download in acceptance tests. It refuses URLs
// that resolve to private addresses, so a local test server cannot be used.
const (
	testLogoURL     = "https://raw.githubusercontent.com/pocket-id/pocket-id/v2.16.0/backend/resources/default-images/logoLight.svg"
	testDarkLogoURL = "https://raw.githubusercontent.com/pocket-id/pocket-id/v2.16.0/backend/resources/default-images/logoDark.svg"
)

// setClientLogoURL sets a client's logo directly through the API, the way an
// operator would in the Pocket ID UI.
func setClientLogoURL(id, logoURL string) error {
	c, err := testClient()
	if err != nil {
		return err
	}

	current, err := c.GetClient(id)
	if err != nil {
		return err
	}

	req := &client.OIDCClientCreateRequest{
		Name:               current.Name,
		CallbackURLs:       current.CallbackURLs,
		LogoutCallbackURLs: current.LogoutCallbackURLs,
		IsPublic:           current.IsPublic,
		PkceEnabled:        current.PkceEnabled,
		IsGroupRestricted:  current.IsGroupRestricted,
		Credentials:        current.Credentials,
		LogoURL:            &logoURL,
	}

	_, err = c.UpdateClient(id, req)
	return err
}

// checkClientLogoFlags asserts which logos Pocket ID reports for a client.
func checkClientLogoFlags(id string, wantLogo, wantDarkLogo bool) error {
	c, err := testClient()
	if err != nil {
		return err
	}

	got, err := c.GetClient(id)
	if err != nil {
		return err
	}

	if got.HasLogo != wantLogo {
		return fmt.Errorf("hasLogo: got %v, want %v", got.HasLogo, wantLogo)
	}
	if got.HasDarkLogo != wantDarkLogo {
		return fmt.Errorf("hasDarkLogo: got %v, want %v", got.HasDarkLogo, wantDarkLogo)
	}
	return nil
}

// checkNoClientNamed asserts no OIDC client with the given name exists.
func checkNoClientNamed(name string) error {
	c, err := testClient()
	if err != nil {
		return err
	}

	clients, err := c.ListClients()
	if err != nil {
		return err
	}

	for _, cl := range clients.Data {
		if cl.Name == name {
			return fmt.Errorf("client %q (%s) was left behind", name, cl.ID)
		}
	}
	return nil
}

// getClientLogo returns the logo image Pocket ID serves for a client.
func getClientLogo(id string, light bool) ([]byte, error) {
	url := fmt.Sprintf("%s/api/oidc/clients/%s/logo?light=%t", os.Getenv("POCKETID_BASE_URL"), id, light)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}
