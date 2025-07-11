//go:build acc
// +build acc

package datasources_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGroupsDataSource(t *testing.T) {
	t.Run("read all groups", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				// Create some test groups first
				{
					Config: testAccGroupsDataSourceConfig_CreateMultipleGroups(),
				},
				// Then retrieve all groups
				{
					Config: testAccGroupsDataSourceConfig_ReadAll(),
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
	})

	t.Run("groups can be filtered locally", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				// Create test groups with specific patterns
				{
					Config: testAccGroupsDataSourceConfig_CreateFilterableGroups(),
				},
				// Then filter them locally
				{
					Config: testAccGroupsDataSourceConfig_WithLocalFiltering(),
					Check: resource.ComposeAggregateTestCheckFunc(
						// Check filtered outputs
						resource.TestCheckResourceAttr("output.admin_groups_count", "value", "2"),
						resource.TestCheckResourceAttr("output.dev_groups_count", "value", "1"),
						resource.TestCheckResourceAttrSet("output.group_map", "value"),
					),
				},
			},
		})
	})
}

func testAccGroupsDataSourceConfig_CreateMultipleGroups() string {
	return `
resource "pocketid_group" "developers" {
  name          = "test_developers"
  friendly_name = "Test Developers"
}

resource "pocketid_group" "admins" {
  name          = "test_admins"
  friendly_name = "Test Admins"
}

resource "pocketid_group" "users" {
  name          = "test_users"
  friendly_name = "Test Users"
}
`
}

func testAccGroupsDataSourceConfig_ReadAll() string {
	return testAccGroupsDataSourceConfig_CreateMultipleGroups() + `
data "pocketid_groups" "all" {
  depends_on = [
    pocketid_group.developers,
    pocketid_group.admins,
    pocketid_group.users
  ]
}
`
}

func testAccGroupsDataSourceConfig_CreateFilterableGroups() string {
	return `
resource "pocketid_group" "admin_global" {
  name          = "test_admin_global"
  friendly_name = "Global Administrators"
}

resource "pocketid_group" "admin_local" {
  name          = "test_admin_local"
  friendly_name = "Local Administrators"
}

resource "pocketid_group" "dev_team" {
  name          = "test_dev_team"
  friendly_name = "Development Team"
}

resource "pocketid_group" "support" {
  name          = "test_support"
  friendly_name = "Support Team"
}
`
}

func testAccGroupsDataSourceConfig_WithLocalFiltering() string {
	return testAccGroupsDataSourceConfig_CreateFilterableGroups() + `
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
    if can(regex("admin", group.name))
  ]

  dev_groups = [
    for group in data.pocketid_groups.all.groups : group
    if can(regex("dev", group.name))
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
`
}
