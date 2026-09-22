//go:build acc
// +build acc

package provider_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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
					resource.TestCheckTypeSetElemAttrPair(resourceName, "allowed_user_groups.*", groupResourceName, "id"),
				),
			},
		},
	})
}

func TestAccResourceClient_groupOrderingStable(t *testing.T) {
	resourceName := "pocketid_client.test"

	groupA, groupB, groupC := "pocketid_group.a.id", "pocketid_group.b.id", "pocketid_group.c.id"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceClientConfig_groupOrdering(groupA, groupB, groupC),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "allowed_user_groups.#", "3"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "allowed_user_groups.*", "pocketid_group.a", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "allowed_user_groups.*", "pocketid_group.b", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "allowed_user_groups.*", "pocketid_group.c", "id"),
				),
			},
			{
				// Same three groups, reversed. Ordering is not a change.
				Config: testAccResourceClientConfig_groupOrdering(groupC, groupB, groupA),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				// Dropping a group is a real change: set semantics must not mask it.
				Config: testAccResourceClientConfig_groupOrdering(groupA, groupB),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "allowed_user_groups.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "allowed_user_groups.*", "pocketid_group.a", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "allowed_user_groups.*", "pocketid_group.b", "id"),
				),
			},
		},
	})
}

func TestAccResourceClient_emptyAllowedGroups(t *testing.T) {
	resourceName := "pocketid_client.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceClientConfig_emptyAllowedGroups(),
				Check: resource.TestCheckResourceAttr(
					resourceName, "allowed_user_groups.#", "0",
				),
			},
			{
				// An empty set must stay empty, not null.
				Config: testAccResourceClientConfig_emptyAllowedGroups(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
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

func testAccResourceClientConfig_groupOrdering(groupRefs ...string) string {
	return fmt.Sprintf(`
resource "pocketid_group" "a" {
  name          = "ordering-group-a"
  friendly_name = "Ordering Group A"
}

resource "pocketid_group" "b" {
  name          = "ordering-group-b"
  friendly_name = "Ordering Group B"
}

resource "pocketid_group" "c" {
  name          = "ordering-group-c"
  friendly_name = "Ordering Group C"
}

resource "pocketid_client" "test" {
  name          = "ordering-client"
  callback_urls = ["https://example.com/callback"]

  allowed_user_groups = [%[1]s]
}
`, strings.Join(groupRefs, ", "))
}

func testAccResourceClientConfig_emptyAllowedGroups() string {
	return `
resource "pocketid_client" "test" {
  name          = "empty-groups-client"
  callback_urls = ["https://example.com/callback"]

  allowed_user_groups = []
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

func TestAccResourceClient_federatedIdentities(t *testing.T) {
	resourceName := "pocketid_client.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceClientConfig_federatedIdentities("fed-client"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "fed-client"),
					resource.TestCheckResourceAttr(resourceName, "federated_identities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "federated_identities.0.issuer", "https://issuer.example.com"),
					resource.TestCheckResourceAttr(resourceName, "federated_identities.0.subject", "subject-1"),
					resource.TestCheckResourceAttr(resourceName, "federated_identities.0.audience", "audience-1"),
				),
			},
		},
	})
}

func testAccResourceClientConfig_federatedIdentities(name string) string {
	return fmt.Sprintf(`
resource "pocketid_client" "test" {
  name          = %[1]q
  callback_urls = ["https://example.com/callback"]

  federated_identities = [
    {
      issuer   = "https://issuer.example.com"
      subject  = "subject-1"
      audience = "audience-1"
    },
  ]
}
`, name)
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

func TestAccResourceClient_requiresPushedAuthorizationRequests(t *testing.T) {
	resourceName := "pocketid_client.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Set PAR to true. On Pocket-ID <= v2.8.0 the API does not echo the
			// field, so the provider preserves the configured value (no
			// inconsistent-result error, no perpetual diff).
			{
				Config: testAccResourceClientConfig_par("par-client", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "requires_pushed_authorization_requests", "true"),
				),
			},
			// Flip it back to false and confirm it applies cleanly.
			{
				Config: testAccResourceClientConfig_par("par-client", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "requires_pushed_authorization_requests", "false"),
				),
			},
		},
	})
}

func TestAccResourceClient_customSecret(t *testing.T) {
	resourceName := "pocketid_client.test"
	customSecret := "custom-secret-0123456789abcdef"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with an explicit client_secret and verify it is honored
			// rather than overwritten by a generated value.
			{
				Config: testAccResourceClientConfig_withSecret("secret-client", customSecret),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "client_secret", customSecret),
				),
			},
			// Re-applying the same config should not produce a diff.
			{
				Config:   testAccResourceClientConfig_withSecret("secret-client", customSecret),
				PlanOnly: true,
			},
		},
	})
}

func TestAccResourceClient_secretRotation(t *testing.T) {
	resourceName := "pocketid_client.test"
	initialSecret := "initial-secret-0123456789abcdef"
	rotatedSecret := "rotated-secret-abcdef0123456789"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceClientConfig_withSecret("rotate-client", initialSecret),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "client_secret", initialSecret),
				),
			},
			// Changing client_secret must rotate it in place via an Update
			// call, not force replacement of the client.
			{
				Config: testAccResourceClientConfig_withSecret("rotate-client", rotatedSecret),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "client_secret", rotatedSecret),
				),
			},
		},
	})
}

func TestAccResourceClient_secretPreservedWhenRemovedFromConfig(t *testing.T) {
	resourceName := "pocketid_client.test"
	customSecret := "preserved-secret-0123456789abcd"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with an explicit secret.
			{
				Config: testAccResourceClientConfig_withSecret("keep-secret-client", customSecret),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "client_secret", customSecret),
				),
			},
			// Remove client_secret from the config entirely. Since it can't
			// be retrieved from the API after creation, the provider must
			// preserve the previously configured value rather than rotating
			// it or producing an inconsistent-result error.
			{
				Config: testAccResourceClientConfig_basic("keep-secret-client", "https://example.com/callback"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "client_secret", customSecret),
				),
			},
			// Re-applying without client_secret should not keep producing a
			// diff (i.e. no perpetual drift/rotation attempts).
			{
				Config:   testAccResourceClientConfig_basic("keep-secret-client", "https://example.com/callback"),
				PlanOnly: true,
			},
		},
	})
}

func testAccResourceClientConfig_withSecret(name, secret string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "pocketid_client" "test" {
  name          = %[1]q
  callback_urls = ["https://example.com/callback"]
  is_public     = false
  client_secret = %[2]q
}
`, name, secret)
}

func testAccResourceClientConfig_par(name string, par bool) string {
	// PAR applies to confidential clients; Pocket-ID coerces it to false for
	// public clients, so this must use a non-public client.
	return testAccProviderConfig() + fmt.Sprintf(`
resource "pocketid_client" "test" {
  name          = %[1]q
  callback_urls = ["https://example.com/callback"]
  is_public     = false

  requires_pushed_authorization_requests = %[2]t
}
`, name, par)
}

func TestAccResourceClient_parPublicConflict(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// PAR on a public client is rejected at plan time.
				Config: testAccProviderConfig() + `
resource "pocketid_client" "test" {
  name          = "par-public-conflict"
  callback_urls = ["https://example.com/callback"]
  is_public     = true

  requires_pushed_authorization_requests = true
}
`,
				ExpectError: regexp.MustCompile("Invalid PAR configuration"),
			},
		},
	})
}

// TestAccResourceClient_preservesUnmanagedFields checks that updating a managed
// attribute does not reset settings the provider does not expose. The update
// endpoint replaces the client in full, so anything the provider omits from the
// payload is reset server-side.
func TestAccResourceClient_preservesUnmanagedFields(t *testing.T) {
	resourceName := "pocketid_client.test"

	var clientID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceClientConfig_basic("preserve-test", "https://example.com/callback"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "id", func(v string) error {
						clientID = v
						return nil
					}),
				),
			},
			{
				// Set the unmanaged fields out of band, the way an operator
				// would in the Pocket ID UI.
				PreConfig: func() {
					if err := setUnmanagedClientFields(clientID); err != nil {
						t.Fatalf("seeding unmanaged fields: %v", err)
					}
				},
				// Change a managed attribute so the resource performs an Update.
				Config: testAccResourceClientConfig_basic("preserve-test-renamed", "https://example.com/callback"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "preserve-test-renamed"),
					func(*terraform.State) error {
						return checkUnmanagedClientFieldsPreserved(clientID)
					},
				),
			},
		},
	})
}
