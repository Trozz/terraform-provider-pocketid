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

func TestAccResourceGroup_basic(t *testing.T) {
	resourceName := "pocketid_group.test"
	groupName := "test-group"
	friendlyName := "Test Group"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceGroupConfig_basic(groupName, friendlyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "friendly_name", friendlyName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					testAccCheckGroupExists(resourceName),
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
				Config: testAccResourceGroupConfig_basic(groupName, "Updated Test Group"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "friendly_name", "Updated Test Group"),
				),
			},
		},
	})
}

func TestAccResourceGroup_withSpecialCharacters(t *testing.T) {
	resourceName := "pocketid_group.test"
	groupName := "test-group-special"
	friendlyName := "Test Group with Special Characters: @#$%"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGroupConfig_basic(groupName, friendlyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "friendly_name", friendlyName),
				),
			},
		},
	})
}

func TestAccResourceGroup_emptyFriendlyName(t *testing.T) {
	resourceName := "pocketid_group.test"
	groupName := "test-group-empty-friendly"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGroupConfig_emptyFriendlyName(groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "friendly_name", ""),
				),
			},
		},
	})
}

func TestAccResourceGroup_invalidName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceGroupConfig_basic("", "Test Group"),
				ExpectError: regexp.MustCompile("Attribute name string length must be at least 1"),
			},
		},
	})
}

func TestAccResourceGroup_duplicateName(t *testing.T) {
	groupName := "test-duplicate-group"
	friendlyName := "Duplicate Test Group"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create first group
			{
				Config: testAccResourceGroupConfig_duplicate(groupName, friendlyName, "first"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pocketid_group.first", "name", groupName),
				),
			},
			// Attempt to create duplicate group
			{
				Config:      testAccResourceGroupConfig_duplicate(groupName, friendlyName, "second"),
				ExpectError: regexp.MustCompile("group already exists"),
			},
		},
	})
}

func TestAccResourceGroup_longFriendlyName(t *testing.T) {
	resourceName := "pocketid_group.test"
	groupName := "test-long-friendly-name"
	// Create a 256+ character friendly name
	longFriendlyName := "This is a very long friendly name that exceeds the typical length limits that might be imposed by the system. " +
		"It contains multiple sentences and goes on and on to test whether the system properly handles long text inputs. " +
		"This helps ensure that the application gracefully handles edge cases with string length validation."

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGroupConfig_basic(groupName, longFriendlyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "friendly_name", longFriendlyName),
				),
			},
		},
	})
}

func TestAccResourceGroup_multipleGroups(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGroupConfig_multiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check first group
					resource.TestCheckResourceAttr("pocketid_group.developers", "name", "developers"),
					resource.TestCheckResourceAttr("pocketid_group.developers", "friendly_name", "Development Team"),
					// Check second group
					resource.TestCheckResourceAttr("pocketid_group.admins", "name", "admins"),
					resource.TestCheckResourceAttr("pocketid_group.admins", "friendly_name", "Administrators"),
					// Check third group
					resource.TestCheckResourceAttr("pocketid_group.users", "name", "users"),
					resource.TestCheckResourceAttr("pocketid_group.users", "friendly_name", "Regular Users"),
				),
			},
		},
	})
}

// Helper function to check if group exists
func testAccCheckGroupExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Group ID is set")
		}

		// Here you would typically make an API call to verify the group exists
		// For now, we just check that the ID is set
		return nil
	}
}

// Configuration functions
func testAccResourceGroupConfig_basic(name, friendlyName string) string {
	return fmt.Sprintf(`
resource "pocketid_group" "test" {
  name          = %[1]q
  friendly_name = %[2]q
}
`, name, friendlyName)
}

func testAccResourceGroupConfig_emptyFriendlyName(name string) string {
	return fmt.Sprintf(`
resource "pocketid_group" "test" {
  name          = %[1]q
  friendly_name = ""
}
`, name)
}

func testAccResourceGroupConfig_duplicate(name, friendlyName, label string) string {
	if label == "first" {
		return fmt.Sprintf(`
resource "pocketid_group" "first" {
  name          = %[1]q
  friendly_name = %[2]q
}
`, name, friendlyName)
	}
	return fmt.Sprintf(`
resource "pocketid_group" "first" {
  name          = %[1]q
  friendly_name = %[2]q
}

resource "pocketid_group" "second" {
  name          = %[1]q
  friendly_name = %[2]q
}
`, name, friendlyName)
}

func testAccResourceGroupConfig_multiple() string {
	return `
resource "pocketid_group" "developers" {
  name          = "developers"
  friendly_name = "Development Team"
}

resource "pocketid_group" "admins" {
  name          = "admins"
  friendly_name = "Administrators"
}

resource "pocketid_group" "users" {
  name          = "users"
  friendly_name = "Regular Users"
}
`
}
