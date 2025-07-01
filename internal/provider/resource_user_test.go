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
		for key, value := range rs.Primary.Attributes {
			if matched, _ := regexp.MatchString(`^groups\.\d+$`, key); matched {
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
  name        = "test-group-1"
  description = "Test group 1"
}

resource "pocketid_group" "test2" {
  name        = "test-group-2"
  description = "Test group 2"
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
  name        = "test-group-1"
  description = "Test group 1"
}

resource "pocketid_group" "test2" {
  name        = "test-group-2"
  description = "Test group 2"
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
