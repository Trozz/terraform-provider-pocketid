//go:build acc
// +build acc

package provider_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLDAPSync_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// First configure LDAP
			{
				Config: testAccLDAPConfigBasic(),
			},
			// Create sync resource
			{
				Config: testAccLDAPSyncBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pocketid_ldap_sync.test", "id", "ldap-sync"),
					resource.TestCheckResourceAttrSet("pocketid_ldap_sync.test", "last_sync_time"),
					resource.TestCheckResourceAttrSet("pocketid_ldap_sync.test", "sync_status"),
				),
			},
			// Update trigger to force new sync
			{
				Config: testAccLDAPSyncUpdated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pocketid_ldap_sync.test", "id", "ldap-sync"),
					resource.TestCheckResourceAttrSet("pocketid_ldap_sync.test", "last_sync_time"),
				),
			},
		},
	})
}

func TestAccLDAPSync_withTimeout(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// First configure LDAP
			{
				Config: testAccLDAPConfigBasic(),
			},
			// Create sync with timeout
			{
				Config: testAccLDAPSyncWithTimeout(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pocketid_ldap_sync.test", "id", "ldap-sync"),
				),
			},
		},
	})
}

func TestAccLDAPSync_ldapDisabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Configure LDAP as disabled
			{
				Config: testAccLDAPConfigDisabled(),
			},
			// Try to create sync - should fail
			{
				Config:      testAccLDAPSyncBasic(),
				ExpectError: regexp.MustCompile("LDAP is not enabled"),
			},
		},
	})
}

func testAccLDAPSyncBasic() string {
	return testAccLDAPConfigBasic() + fmt.Sprintf(`
resource "pocketid_ldap_sync" "test" {
  triggers = {
    timestamp = "%s"
  }
}
`, time.Now().Format(time.RFC3339))
}

func testAccLDAPSyncUpdated() string {
	return testAccLDAPConfigBasic() + fmt.Sprintf(`
resource "pocketid_ldap_sync" "test" {
  triggers = {
    timestamp = "%s"
    force     = "true"
  }
}
`, time.Now().Add(1*time.Minute).Format(time.RFC3339))
}

func testAccLDAPSyncWithTimeout() string {
	return testAccLDAPConfigBasic() + fmt.Sprintf(`
resource "pocketid_ldap_sync" "test" {
  triggers = {
    timestamp = "%s"
  }

  timeouts {
    create = "30s"
    update = "30s"
  }
}
`, time.Now().Format(time.RFC3339))
}
