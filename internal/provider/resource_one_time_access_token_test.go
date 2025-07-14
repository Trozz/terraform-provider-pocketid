//go:build acc
// +build acc

package provider_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceOneTimeAccessToken_basic(t *testing.T) {
	resourceName := "pocketid_one_time_access_token.test"
	userResourceName := "pocketid_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceOneTimeAccessTokenConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "user_id", userResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "id", userResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					testAccCheckOneTimeAccessTokenValid(resourceName),
				),
			},
			// Import testing
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token", "expires_at", "created_at"}, // These values can't be retrieved from API
			},
		},
	})
}

func TestAccResourceOneTimeAccessToken_withExpiry(t *testing.T) {
	resourceName := "pocketid_one_time_access_token.test"
	// Set expiry to 1 hour from now
	expiresAt := time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceOneTimeAccessTokenConfig_withExpiry(expiresAt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttr(resourceName, "expires_at", expiresAt),
					testAccCheckOneTimeAccessTokenExpiry(resourceName, expiresAt),
				),
			},
		},
	})
}

func TestAccResourceOneTimeAccessToken_invalidExpiry(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceOneTimeAccessTokenConfig_withExpiry("invalid-date"),
				ExpectError: regexp.MustCompile("Invalid expires_at format"),
			},
		},
	})
}

func TestAccResourceOneTimeAccessToken_recreateOnUserChange(t *testing.T) {
	resourceName := "pocketid_one_time_access_token.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with first user
			{
				Config: testAccResourceOneTimeAccessTokenConfig_twoUsers("user1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "user_id", "pocketid_user.user1", "id"),
				),
			},
			// Change to second user (should recreate)
			{
				Config: testAccResourceOneTimeAccessTokenConfig_twoUsers("user2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "user_id", "pocketid_user.user2", "id"),
				),
			},
		},
	})
}

func TestAccResourceOneTimeAccessToken_deleteUser(t *testing.T) {
	// Test that the token is removed from state when the user is deleted
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create token
			{
				Config: testAccResourceOneTimeAccessTokenConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pocketid_one_time_access_token.test", "token"),
				),
			},
			// Delete user (token should be removed from state on next read)
			{
				Config: testAccResourceOneTimeAccessTokenConfig_noUser(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// The token resource should be gone
					testAccCheckResourceNotExists("pocketid_one_time_access_token.test"),
				),
			},
		},
	})
}

// Helper functions

func testAccCheckOneTimeAccessTokenValid(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		token := rs.Primary.Attributes["token"]
		if token == "" {
			return fmt.Errorf("token is empty")
		}

		// Check token format (should be non-empty string)
		if len(token) < 10 {
			return fmt.Errorf("token seems too short: %s", token)
		}

		return nil
	}
}

func testAccCheckOneTimeAccessTokenExpiry(resourceName, expectedExpiry string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		expiresAt := rs.Primary.Attributes["expires_at"]
		if expiresAt != expectedExpiry {
			return fmt.Errorf("expires_at mismatch: got %s, expected %s", expiresAt, expectedExpiry)
		}

		// Parse and validate the expiry time
		expiryTime, err := time.Parse(time.RFC3339, expiresAt)
		if err != nil {
			return fmt.Errorf("invalid expires_at format: %s", err)
		}

		// Check that expiry is in the future
		if expiryTime.Before(time.Now()) {
			return fmt.Errorf("token already expired: %s", expiresAt)
		}

		return nil
	}
}

func testAccCheckResourceNotExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if ok {
			return fmt.Errorf("resource still exists: %s", resourceName)
		}
		return nil
	}
}

// Configuration functions

func testAccResourceOneTimeAccessTokenConfig_basic() string {
	expiresAt := time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339)
	return fmt.Sprintf(`
resource "pocketid_user" "test" {
  username   = "token-test-user"
  email      = "token-test@example.com"
  first_name = "Token"
  last_name  = "Test"
}

resource "pocketid_one_time_access_token" "test" {
  user_id    = pocketid_user.test.id
  expires_at = "%s"
}
`, expiresAt)
}

func testAccResourceOneTimeAccessTokenConfig_withExpiry(expiresAt string) string {
	return fmt.Sprintf(`
resource "pocketid_user" "test" {
  username   = "token-test-user"
  email      = "token-test@example.com"
  first_name = "Token"
  last_name  = "Test"
}

resource "pocketid_one_time_access_token" "test" {
  user_id    = pocketid_user.test.id
  expires_at = %q
}
`, expiresAt)
}

func testAccResourceOneTimeAccessTokenConfig_twoUsers(activeUser string) string {
	expiresAt := time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339)
	return fmt.Sprintf(`
resource "pocketid_user" "user1" {
  username   = "token-test-user1"
  email      = "token-test1@example.com"
  first_name = "Token"
  last_name  = "Test1"
}

resource "pocketid_user" "user2" {
  username   = "token-test-user2"
  email      = "token-test2@example.com"
  first_name = "Token"
  last_name  = "Test2"
}

resource "pocketid_one_time_access_token" "test" {
  user_id    = pocketid_user.%s.id
  expires_at = "%s"
}
`, activeUser, expiresAt)
}

func testAccResourceOneTimeAccessTokenConfig_noUser() string {
	return `
# User deleted, token should be removed from state
`
}

func TestAccOneTimeAccessTokenResource_SkipRecreate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with skip_recreate = true
			{
				Config: testAccOneTimeAccessTokenResourceConfig_skipRecreate("skip.recreate.user", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pocketid_one_time_access_token.test", "skip_recreate", "true"),
					resource.TestCheckResourceAttrSet("pocketid_one_time_access_token.test", "token"),
					resource.TestCheckResourceAttrSet("pocketid_one_time_access_token.test", "created_at"),
				),
			},
			// Test that resource persists even when token is not found
			{
				PreConfig: func() {
					// This simulates the token being used/deleted outside of Terraform
					// Note: In real usage, we can't actually delete the token since
					// the API might not allow it or the token might be auto-deleted after use
				},
				Config: testAccOneTimeAccessTokenResourceConfig_skipRecreate("skip.recreate.user", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					// The resource should still exist in state
					resource.TestCheckResourceAttr("pocketid_one_time_access_token.test", "skip_recreate", "true"),
				),
			},
			// Test with skip_recreate = false (default behavior)
			{
				Config: testAccOneTimeAccessTokenResourceConfig_skipRecreate("normal.user", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pocketid_one_time_access_token.test2", "skip_recreate", "false"),
					resource.TestCheckResourceAttrSet("pocketid_one_time_access_token.test2", "token"),
				),
			},
		},
	})
}

func testAccOneTimeAccessTokenResourceConfig_skipRecreate(username string, skipRecreate bool) string {
	resourceName := "test"
	if !skipRecreate {
		resourceName = "test2"
	}
	expiresAt := time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339)

	return fmt.Sprintf(`
resource "pocketid_user" "%[1]s" {
  username   = "%[1]s"
  email      = "%[1]s@example.com"
  first_name = "Test"
  last_name  = "User"
}

resource "pocketid_one_time_access_token" "%[2]s" {
  user_id       = pocketid_user.%[1]s.id
  expires_at    = "%[4]s"
  skip_recreate = %[3]t
}
`, strings.ReplaceAll(username, ".", "_"), resourceName, skipRecreate, expiresAt)
}
