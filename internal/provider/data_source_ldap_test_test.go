//go:build acc
// +build acc

package provider_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLDAPTestDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// First configure LDAP
			{
				Config: testAccLDAPConfigBasic(),
			},
			// Test connection
			{
				Config: testAccLDAPTestDataSourceBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pocketid_ldap_test.test", "id", "ldap-test"),
					resource.TestCheckResourceAttr("data.pocketid_ldap_test.test", "connection_success", "true"),
					resource.TestCheckResourceAttr("data.pocketid_ldap_test.test", "bind_success", "true"),
					resource.TestCheckResourceAttrSet("data.pocketid_ldap_test.test", "user_count"),
					resource.TestCheckResourceAttrSet("data.pocketid_ldap_test.test", "group_count"),
				),
			},
		},
	})
}

func TestAccLDAPTestDataSource_invalidCredentials(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Configure LDAP with wrong password
			{
				Config: testAccLDAPConfigInvalidCredentials(),
			},
			// Test connection - should show bind failure
			{
				Config: testAccLDAPTestDataSourceBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pocketid_ldap_test.test", "connection_success", "true"),
					resource.TestCheckResourceAttr("data.pocketid_ldap_test.test", "bind_success", "false"),
					resource.TestCheckResourceAttrSet("data.pocketid_ldap_test.test", "error"),
				),
			},
		},
	})
}

func TestAccLDAPTestDataSource_ldapDisabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Configure LDAP as disabled
			{
				Config: testAccLDAPConfigDisabled(),
			},
			// Test connection - should fail
			{
				Config:      testAccLDAPTestDataSourceBasic(),
				ExpectError: regexp.MustCompile("LDAP is not configured or disabled"),
			},
		},
	})
}

func testAccLDAPTestDataSourceBasic() string {
	return testAccProviderConfig() + `
data "pocketid_ldap_test" "test" {}
`
}

func testAccLDAPConfigInvalidCredentials() string {
	return testAccProviderConfig() + `
resource "pocketid_ldap_config" "test" {
  enabled = true

  url               = "ldap://localhost:3890"
  bind_dn          = "cn=admin,dc=example,dc=com"
  bind_password    = "wrong-password"
  base_dn          = "dc=example,dc=com"
  skip_cert_verify = true

  user_search_filter = "(objectClass=inetOrgPerson)"
}
`
}
