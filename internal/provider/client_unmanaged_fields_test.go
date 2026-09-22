//go:build acc
// +build acc

package provider_test

import (
	"fmt"
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
