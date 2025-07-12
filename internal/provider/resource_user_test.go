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

func TestAccResourceUser_basic(t *testing.T) {
	resourceName := "pocketid_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceUserConfig_basic("test-user", "test@example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "username", "test-user"),
					resource.TestCheckResourceAttr(resourceName, "email", "test@example.com"),
					resource.TestCheckResourceAttr(resourceName, "first_name", "Test"),
					resource.TestCheckResourceAttr(resourceName, "last_name", "User"),
					resource.TestCheckResourceAttr(resourceName, "is_admin", "false"),
					resource.TestCheckResourceAttr(resourceName, "disabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccResourceUserConfig_basic("test-user", "updated@example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "email", "updated@example.com"),
				),
			},
		},
	})
}

func TestAccResourceUser_withGroups(t *testing.T) {
	resourceName := "pocketid_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create user with multiple groups
			{
				Config: testAccResourceUserConfig_withGroups("test-user", "test@example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "username", "test-user"),
					resource.TestCheckResourceAttr(resourceName, "email", "test@example.com"),
					resource.TestCheckResourceAttr(resourceName, "groups.#", "2"),
					testAccCheckUserGroupsSet(resourceName),
				),
			},
			// Verify no changes on re-apply (groups ordering should not matter)
			{
				Config:   testAccResourceUserConfig_withGroups("test-user", "test@example.com"),
				PlanOnly: true,
			},
			// Update groups
			{
				Config: testAccResourceUserConfig_withSingleGroup("test-user", "test@example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "groups.#", "1"),
				),
			},
			// Remove all groups
			{
				Config: testAccResourceUserConfig_basic("test-user", "test@example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "groups.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceUser_disabled(t *testing.T) {
	resourceName := "pocketid_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create disabled user
			{
				Config: testAccResourceUserConfig_disabled("disabled-user", "disabled@example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "username", "disabled-user"),
					resource.TestCheckResourceAttr(resourceName, "disabled", "true"),
				),
			},
			// Enable user
			{
				Config: testAccResourceUserConfig_basic("disabled-user", "disabled@example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "disabled", "false"),
				),
			},
		},
	})
}

// testAccCheckUserGroupsSet verifies that groups are stored as a set (unordered)
func testAccCheckUserGroupsSet(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		// Get all group attributes
		groupCount := 0
		groups := make(map[string]bool)
		groupsRegex := regexp.MustCompile(`^groups\.\d+$`)
		for key, value := range rs.Primary.Attributes {
			if groupsRegex.MatchString(key) {
				groups[value] = true
				groupCount++
			}
		}

		// Verify we have the expected number of unique groups
		if len(groups) != groupCount {
			return fmt.Errorf("duplicate groups found in state")
		}

		return nil
	}
}

func testAccResourceUserConfig_basic(username, email string) string {
	return fmt.Sprintf(`
resource "pocketid_user" "test" {
  username   = %[1]q
  email      = %[2]q
  first_name = "Test"
  last_name  = "User"
}
`, username, email)
}

func testAccResourceUserConfig_withGroups(username, email string) string {
	return fmt.Sprintf(`
resource "pocketid_group" "test1" {
  name          = "test-group-1"
  friendly_name = "Test group 1"
}

resource "pocketid_group" "test2" {
  name          = "test-group-2"
  friendly_name = "Test group 2"
}

resource "pocketid_user" "test" {
  username   = %[1]q
  email      = %[2]q
  first_name = "Test"
  last_name  = "User"

  groups = [
    pocketid_group.test2.id,
    pocketid_group.test1.id,
  ]
}
`, username, email)
}

func testAccResourceUserConfig_withSingleGroup(username, email string) string {
	return fmt.Sprintf(`
resource "pocketid_group" "test1" {
  name          = "test-group-1"
  friendly_name = "Test group 1"
}

resource "pocketid_group" "test2" {
  name          = "test-group-2"
  friendly_name = "Test group 2"
}

resource "pocketid_user" "test" {
  username   = %[1]q
  email      = %[2]q
  first_name = "Test"
  last_name  = "User"

  groups = [
    pocketid_group.test1.id,
  ]
}
`, username, email)
}

func testAccResourceUserConfig_disabled(username, email string) string {
	return fmt.Sprintf(`
resource "pocketid_user" "test" {
  username   = %[1]q
  email      = %[2]q
  first_name = "Test"
  last_name  = "User"
  disabled   = true
}
`, username, email)
}

func TestAccResourceUser_adminUser(t *testing.T) {
	resourceName := "pocketid_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create admin user
			{
				Config: testAccResourceUserConfig_admin("admin-user", "admin@example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "username", "admin-user"),
					resource.TestCheckResourceAttr(resourceName, "is_admin", "true"),
				),
			},
			// Remove admin privileges
			{
				Config: testAccResourceUserConfig_basic("admin-user", "admin@example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "is_admin", "false"),
				),
			},
		},
	})
}

func TestAccResourceUser_withLocale(t *testing.T) {
	resourceName := "pocketid_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create user with locale
			{
				Config: testAccResourceUserConfig_withLocale("locale-user", "locale@example.com", "fr-FR"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "username", "locale-user"),
					resource.TestCheckResourceAttr(resourceName, "locale", "fr-FR"),
				),
			},
			// Update locale
			{
				Config: testAccResourceUserConfig_withLocale("locale-user", "locale@example.com", "en-US"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "locale", "en-US"),
				),
			},
		},
	})
}

func TestAccResourceUser_invalidEmail(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceUserConfig_basic("test-user", "invalid-email"),
				ExpectError: regexp.MustCompile("Email must be a valid email address"),
			},
		},
	})
}

func TestAccResourceUser_duplicateUsername(t *testing.T) {
	username := "duplicate-user"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create first user
			{
				Config: testAccResourceUserConfig_duplicate(username, "user1@example.com", "first"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pocketid_user.first", "username", username),
				),
			},
			// Attempt to create duplicate user
			{
				Config:      testAccResourceUserConfig_duplicate(username, "user2@example.com", "second"),
				ExpectError: regexp.MustCompile("Username is already in use"),
			},
		},
	})
}

func TestAccResourceUser_updateImmutableField(t *testing.T) {
	resourceName := "pocketid_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create user
			{
				Config: testAccResourceUserConfig_basic("original-username", "test@example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "username", "original-username"),
				),
			},
			// Attempt to update username (should recreate resource)
			{
				Config: testAccResourceUserConfig_basic("new-username", "test@example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "username", "new-username"),
				),
			},
		},
	})
}

func testAccResourceUserConfig_admin(username, email string) string {
	return fmt.Sprintf(`
resource "pocketid_user" "test" {
  username   = %[1]q
  email      = %[2]q
  first_name = "Test"
  last_name  = "User"
  is_admin   = true
}
`, username, email)
}

func testAccResourceUserConfig_withLocale(username, email, locale string) string {
	return fmt.Sprintf(`
resource "pocketid_user" "test" {
  username   = %[1]q
  email      = %[2]q
  first_name = "Test"
  last_name  = "User"
  locale     = %[3]q
}
`, username, email, locale)
}

func testAccResourceUserConfig_duplicate(username, email, label string) string {
	if label == "first" {
		return fmt.Sprintf(`
resource "pocketid_user" "first" {
  username   = %[1]q
  email      = %[2]q
  first_name = "Test"
  last_name  = "User"
}
`, username, email)
	}
	return fmt.Sprintf(`
resource "pocketid_user" "first" {
  username   = %[1]q
  email      = "user1@example.com"
  first_name = "Test"
  last_name  = "User"
}

resource "pocketid_user" "second" {
  username   = %[1]q
  email      = %[2]q
  first_name = "Test"
  last_name  = "User2"
}
`, username, email)
}
