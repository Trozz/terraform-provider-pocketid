//go:build acc
// +build acc

package provider_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// LDAP is not configured in the test environment, so triggering a sync fails.
// This verifies the resource is wired up and surfaces the API error. A
// successful sync can only be tested against a live LDAP server.
func TestAccResourceLdapSync_errorsWhenLdapDisabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "pocketid_ldap_sync" "test" {}
`,
				ExpectError: regexp.MustCompile("Error syncing LDAP"),
			},
		},
	})
}
