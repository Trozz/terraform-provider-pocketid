//go:build acc
// +build acc

package datasources_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGroupsDataSource_ReadAll(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create some test groups first
			{
				Config: testAccGroupsDataSourceConfig_CreateMultipleGroups(rName),
			},
			// Then retrieve all groups
			{
				Config: testAccGroupsDataSourceConfig_ReadAll(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check that we have at least the groups we created
					resource.TestMatchResourceAttr("data.pocketid_groups.all", "groups.#", regexp.MustCompile(`^[3-9]\d*|[1-9]\d+$`)), // At least 3
					// Check specific group attributes exist
					resource.TestCheckResourceAttrSet("data.pocketid_groups.all", "groups.0.id"),
					resource.TestCheckResourceAttrSet("data.pocketid_groups.all", "groups.0.name"),
					resource.TestCheckResourceAttrSet("data.pocketid_groups.all", "groups.0.friendly_name"),
				),
			},
		},
	})
}

func TestAccGroupsDataSource_FilterLocally(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create test groups with specific patterns
			{
				Config: testAccGroupsDataSourceConfig_CreateFilterableGroups(rName),
			},
			// Then filter them locally
			{
				Config: testAccGroupsDataSourceConfig_WithLocalFiltering(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check filtered outputs using TestCheckOutput
					resource.TestCheckOutput("admin_groups_count", "2"),
					resource.TestCheckOutput("dev_groups_count", "1"),
					// Test map output by checking specific keys exist
					resource.TestCheckOutput("has_admin_global", "true"),
					resource.TestCheckOutput("has_dev_team", "true"),
					// Don't check exact map size as it includes all groups in the system
					resource.TestCheckOutput("admin_global_id_not_empty", "true"),
				),
			},
		},
	})
}

func testAccGroupsDataSourceConfig_CreateMultipleGroups(rName string) string {
	return fmt.Sprintf(`
resource "pocketid_group" "developers" {
  name          = "%s_developers"
  friendly_name = "Test Developers"
}

resource "pocketid_group" "admins" {
  name          = "%s_admins"
  friendly_name = "Test Admins"
}

resource "pocketid_group" "users" {
  name          = "%s_users"
  friendly_name = "Test Users"
}
`, rName, rName, rName)
}

func testAccGroupsDataSourceConfig_ReadAll(rName string) string {
	return testAccGroupsDataSourceConfig_CreateMultipleGroups(rName) + `
data "pocketid_groups" "all" {
  depends_on = [
    pocketid_group.developers,
    pocketid_group.admins,
    pocketid_group.users
  ]
}
`
}

func testAccGroupsDataSourceConfig_CreateFilterableGroups(rName string) string {
	return fmt.Sprintf(`
resource "pocketid_group" "admin_global" {
  name          = "%s_admin_global"
  friendly_name = "Global Administrators"
}

resource "pocketid_group" "admin_local" {
  name          = "%s_admin_local"
  friendly_name = "Local Administrators"
}

resource "pocketid_group" "dev_team" {
  name          = "%s_dev_team"
  friendly_name = "Development Team"
}

resource "pocketid_group" "support" {
  name          = "%s_support"
  friendly_name = "Support Team"
}
`, rName, rName, rName, rName)
}

func testAccGroupsDataSourceConfig_WithLocalFiltering(rName string) string {
	return testAccGroupsDataSourceConfig_CreateFilterableGroups(rName) + fmt.Sprintf(`
data "pocketid_groups" "all" {
  depends_on = [
    pocketid_group.admin_global,
    pocketid_group.admin_local,
    pocketid_group.dev_team,
    pocketid_group.support
  ]
}

locals {
  admin_groups = [
    for group in data.pocketid_groups.all.groups : group
    if can(regex("%s_admin", group.name))
  ]

  dev_groups = [
    for group in data.pocketid_groups.all.groups : group
    if can(regex("%s_dev", group.name))
  ]

  group_name_to_id = {
    for group in data.pocketid_groups.all.groups :
    group.name => group.id
  }
}

output "admin_groups_count" {
  value = length(local.admin_groups)
}

output "dev_groups_count" {
  value = length(local.dev_groups)
}

output "group_map" {
  value = local.group_name_to_id
}

output "has_admin_global" {
  value = contains(keys(local.group_name_to_id), "%s_admin_global")
}

output "has_dev_team" {
  value = contains(keys(local.group_name_to_id), "%s_dev_team")
}

output "map_size" {
  value = length(local.group_name_to_id)
}

output "admin_global_id_not_empty" {
  value = local.group_name_to_id["%s_admin_global"] != ""
}
`, rName, rName, rName, rName, rName)
}
