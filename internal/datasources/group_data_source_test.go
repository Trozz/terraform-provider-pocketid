//go:build acc
// +build acc

package datasources_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/Trozz/terraform-provider-pocketid/internal/provider"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"pocketid": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// Check for required environment variables
	if v := os.Getenv("POCKETID_BASE_URL"); v == "" {
		t.Fatal("POCKETID_BASE_URL must be set for acceptance tests")
	}

	if v := os.Getenv("POCKETID_API_TOKEN"); v == "" {
		t.Fatal("POCKETID_API_TOKEN must be set for acceptance tests")
	}
}

func TestAccGroupDataSource(t *testing.T) {
	t.Run("lookup by ID", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				// First create a group
				{
					Config: testAccGroupDataSourceConfig_CreateGroup("test_group", "Test Group"),
				},
				// Then look it up by ID
				{
					Config: testAccGroupDataSourceConfig_LookupByID("test_group", "Test Group"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.pocketid_group.test", "name", "test_group"),
						resource.TestCheckResourceAttr("data.pocketid_group.test", "friendly_name", "Test Group"),
						resource.TestCheckResourceAttrSet("data.pocketid_group.test", "id"),
					),
				},
			},
		})
	})

	t.Run("lookup by name", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				// First create a group
				{
					Config: testAccGroupDataSourceConfig_CreateGroup("test_group_name", "Test Group By Name"),
				},
				// Then look it up by name
				{
					Config: testAccGroupDataSourceConfig_LookupByName("test_group_name", "Test Group By Name"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.pocketid_group.test", "name", "test_group_name"),
						resource.TestCheckResourceAttr("data.pocketid_group.test", "friendly_name", "Test Group By Name"),
						resource.TestCheckResourceAttrSet("data.pocketid_group.test", "id"),
					),
				},
			},
		})
	})

	t.Run("error when neither ID nor name provided", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccGroupDataSourceConfig_NoIdentifier(),
					ExpectError: regexp.MustCompile(`Either 'id' or 'name' must be provided`),
				},
			},
		})
	})

	t.Run("error when group not found", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccGroupDataSourceConfig_NotFound(),
					ExpectError: regexp.MustCompile(`No group found`),
				},
			},
		})
	})
}

func testAccGroupDataSourceConfig_CreateGroup(name, friendlyName string) string {
	return fmt.Sprintf(`
resource "pocketid_group" "test" {
  name          = %[1]q
  friendly_name = %[2]q
}
`, name, friendlyName)
}

func testAccGroupDataSourceConfig_LookupByID(name, friendlyName string) string {
	return fmt.Sprintf(`
resource "pocketid_group" "test" {
  name          = %[1]q
  friendly_name = %[2]q
}

data "pocketid_group" "test" {
  id = pocketid_group.test.id
}
`, name, friendlyName)
}

func testAccGroupDataSourceConfig_LookupByName(name, friendlyName string) string {
	return fmt.Sprintf(`
resource "pocketid_group" "test" {
  name          = %[1]q
  friendly_name = %[2]q
}

data "pocketid_group" "test" {
  name = pocketid_group.test.name
}
`, name, friendlyName)
}

func testAccGroupDataSourceConfig_NoIdentifier() string {
	return `
data "pocketid_group" "test" {
  # Neither ID nor name provided
}
`
}

func testAccGroupDataSourceConfig_NotFound() string {
	return `
data "pocketid_group" "test" {
  name = "non_existent_group_12345"
}
`
}
