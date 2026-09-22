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
