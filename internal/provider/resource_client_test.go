//go:build acc
// +build acc

package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceClient_basic(t *testing.T) {
	resourceName := "pocketid_client.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceClientConfig_basic("test-client", "https://example.com/callback"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-client"),
					resource.TestCheckResourceAttr(resourceName, "callback_urls.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "callback_urls.0", "https://example.com/callback"),
					resource.TestCheckResourceAttr(resourceName, "is_public", "false"),
					resource.TestCheckResourceAttr(resourceName, "pkce_enabled", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "client_secret"),
					resource.TestCheckResourceAttr(resourceName, "has_logo", "false"),
				),
			},
			// ImportState testing
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_secret"}, // Secret can't be retrieved after creation
			},
			// Update and Read testing
			{
				Config: testAccResourceClientConfig_basic("updated-client", "https://example.com/callback"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "updated-client"),
				),
			},
		},
	})
}

func TestAccResourceClient_publicClient(t *testing.T) {
	resourceName := "pocketid_client.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceClientConfig_public("public-client"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "public-client"),
					resource.TestCheckResourceAttr(resourceName, "is_public", "true"),
					resource.TestCheckResourceAttr(resourceName, "pkce_enabled", "true"),
					resource.TestCheckNoResourceAttr(resourceName, "client_secret"),
				),
			},
		},
	})
}

func TestAccResourceClient_multipleCallbackURLs(t *testing.T) {
	resourceName := "pocketid_client.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceClientConfig_multipleCallbacks("multi-callback-client"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "multi-callback-client"),
					resource.TestCheckResourceAttr(resourceName, "callback_urls.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "callback_urls.0", "https://example.com/callback1"),
					resource.TestCheckResourceAttr(resourceName, "callback_urls.1", "https://example.com/callback2"),
					resource.TestCheckResourceAttr(resourceName, "callback_urls.2", "https://example.com/callback3"),
				),
			},
		},
	})
}

func TestAccResourceClient_withLogoutCallbacks(t *testing.T) {
	resourceName := "pocketid_client.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceClientConfig_withLogoutCallbacks("logout-client"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "logout-client"),
					resource.TestCheckResourceAttr(resourceName, "logout_callback_urls.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "logout_callback_urls.0", "https://example.com/logout1"),
					resource.TestCheckResourceAttr(resourceName, "logout_callback_urls.1", "https://example.com/logout2"),
				),
			},
		},
	})
}

func TestAccResourceClient_withAllowedGroups(t *testing.T) {
	resourceName := "pocketid_client.test"
	groupResourceName := "pocketid_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceClientConfig_withAllowedGroups(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "restricted-client"),
					resource.TestCheckResourceAttr(resourceName, "allowed_user_groups.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "allowed_user_groups.0", groupResourceName, "id"),
				),
			},
		},
	})
}

func TestAccResourceClient_pkceDisabled(t *testing.T) {
	resourceName := "pocketid_client.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceClientConfig_pkceDisabled("no-pkce-client"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "no-pkce-client"),
					resource.TestCheckResourceAttr(resourceName, "pkce_enabled", "false"),
				),
			},
		},
	})
}

func TestAccResourceClient_invalidCallbackURL(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceClientConfig_invalidCallback(),
				ExpectError: regexp.MustCompile("invalid callback URL"),
			},
		},
	})
}

func TestAccResourceClient_emptyName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceClientConfig_emptyName(),
				ExpectError: regexp.MustCompile("Attribute name string length must be between 1 and 50"),
			},
		},
	})
}

// Test configuration functions

func testAccResourceClientConfig_basic(name, callbackURL string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "pocketid_client" "test" {
  name          = %[1]q
  callback_urls = [%[2]q]
}
`, name, callbackURL)
}

func testAccResourceClientConfig_public(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "pocketid_client" "test" {
  name          = %[1]q
  callback_urls = ["https://example.com/callback"]
  is_public     = true
}
`, name)
}

func testAccResourceClientConfig_multipleCallbacks(name string) string {
	return fmt.Sprintf(`
resource "pocketid_client" "test" {
  name = %[1]q
  callback_urls = [
    "https://example.com/callback1",
    "https://example.com/callback2",
    "https://example.com/callback3"
  ]
}
`, name)
}

func testAccResourceClientConfig_withLogoutCallbacks(name string) string {
	return fmt.Sprintf(`
resource "pocketid_client" "test" {
  name          = %[1]q
  callback_urls = ["https://example.com/callback"]
  logout_callback_urls = [
    "https://example.com/logout1",
    "https://example.com/logout2"
  ]
}
`, name)
}

func testAccResourceClientConfig_withAllowedGroups() string {
	return `
resource "pocketid_group" "test" {
  name          = "test-group"
  friendly_name = "Test Group"
}

resource "pocketid_client" "test" {
  name          = "restricted-client"
  callback_urls = ["https://example.com/callback"]
  allowed_user_groups = [pocketid_group.test.id]
}
`
}

func testAccResourceClientConfig_pkceDisabled(name string) string {
	return fmt.Sprintf(`
resource "pocketid_client" "test" {
  name          = %[1]q
  callback_urls = ["https://example.com/callback"]
  pkce_enabled  = false
}
`, name)
}

func testAccResourceClientConfig_invalidCallback() string {
	return `
resource "pocketid_client" "test" {
  name          = "invalid-client"
  callback_urls = ["not-a-valid-url"]
}
`
}

func testAccResourceClientConfig_emptyName() string {
	return `
resource "pocketid_client" "test" {
  name          = ""
  callback_urls = ["https://example.com/callback"]
}
`
}

func TestAccResourceClient_updateCallbackURLs(t *testing.T) {
	resourceName := "pocketid_client.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with one callback URL
			{
				Config: testAccResourceClientConfig_basic("test-client", "https://example.com/callback"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "callback_urls.#", "1"),
				),
			},
			// Add more callback URLs
			{
				Config: testAccResourceClientConfig_multipleCallbacks("test-client"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "callback_urls.#", "3"),
				),
			},
			// Remove callback URLs
			{
				Config: testAccResourceClientConfig_basic("test-client", "https://example.com/callback"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "callback_urls.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceClient_generateSecret(t *testing.T) {
	resourceName := "pocketid_client.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create client and verify secret is set
			{
				Config: testAccResourceClientConfig_basic("test-client", "https://example.com/callback"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "client_secret"),
					testAccCheckClientSecretNotEmpty(resourceName),
				),
			},
		},
	})
}

func TestAccResourceClient_longName(t *testing.T) {
	longName := "This is a very long client name that might exceed typical length limits and tests the system's ability to handle long strings"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceClientConfig_basic(longName, "https://example.com/callback"),
				ExpectError: regexp.MustCompile("Attribute name string length must be between 1 and 50"),
			},
		},
	})
}

// Test helper functions

func testAccCheckClientSecretNotEmpty(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		secret := rs.Primary.Attributes["client_secret"]
		if secret == "" {
			return fmt.Errorf("client_secret is empty")
		}

		return nil
	}
}
