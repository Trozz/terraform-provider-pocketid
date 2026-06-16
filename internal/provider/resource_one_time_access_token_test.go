//go:build acc
// +build acc

package provider_test

import (
	"fmt"
	"regexp"
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
				Config: testAccResourceOneTimeAccessTokenConfig_basic("1h"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "user_id", userResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "id", userResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "ttl", "1h"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					testAccCheckOneTimeAccessTokenValid(resourceName),
					testAccCheckOneTimeAccessTokenExpiryInFuture(resourceName),
				),
			},
			// Import testing. The token value and computed timestamps cannot be
			// recovered from the API, and ttl is config-only, so they are ignored.
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token", "ttl", "expires_at", "created_at"},
			},
		},
	})
}

func TestAccResourceOneTimeAccessToken_invalidTTL(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceOneTimeAccessTokenConfig_basic("not-a-duration"),
				ExpectError: regexp.MustCompile("Invalid ttl"),
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
	// Removing the token (and its user) from config destroys it cleanly.
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceOneTimeAccessTokenConfig_basic("1h"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pocketid_one_time_access_token.test", "token"),
				),
			},
			{
				Config: testAccResourceOneTimeAccessTokenConfig_noUser(),
				Check: resource.ComposeAggregateTestCheckFunc(
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

		if rs.Primary.Attributes["token"] == "" {
			return fmt.Errorf("token is empty")
		}

		return nil
	}
}

func testAccCheckOneTimeAccessTokenExpiryInFuture(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		expiresAt := rs.Primary.Attributes["expires_at"]
		expiryTime, err := time.Parse(time.RFC3339, expiresAt)
		if err != nil {
			return fmt.Errorf("invalid expires_at format: %s", err)
		}
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

func testAccResourceOneTimeAccessTokenConfig_basic(ttl string) string {
	return fmt.Sprintf(`
resource "pocketid_user" "test" {
  username   = "token-test-user"
  email      = "token-test@example.com"
  first_name = "Token"
  last_name  = "Test"
}

resource "pocketid_one_time_access_token" "test" {
  user_id = pocketid_user.test.id
  ttl     = %q
}
`, ttl)
}

func testAccResourceOneTimeAccessTokenConfig_twoUsers(activeUser string) string {
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
  user_id = pocketid_user.%s.id
  ttl     = "1h"
}
`, activeUser)
}

func testAccResourceOneTimeAccessTokenConfig_noUser() string {
	return `
# User deleted, token should be removed from state
`
}
