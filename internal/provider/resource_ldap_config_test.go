//go:build acc
// +build acc

package provider_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLDAPConfig_basic(t *testing.T) {
	resourceName := "pocketid_ldap_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccLDAPConfigBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "url", "ldap://localhost:3890"),
					resource.TestCheckResourceAttr(resourceName, "bind_dn", "cn=admin,dc=example,dc=com"),
					resource.TestCheckResourceAttr(resourceName, "base_dn", "dc=example,dc=com"),
					resource.TestCheckResourceAttr(resourceName, "skip_cert_verify", "true"),
					resource.TestCheckResourceAttr(resourceName, "user_search_filter", "(objectClass=inetOrgPerson)"),
					resource.TestCheckResourceAttr(resourceName, "id", "ldap"),
				),
			},
			// ImportState testing
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bind_password"},
			},
			// Update and Read testing
			{
				Config: testAccLDAPConfigUpdated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "user_search_filter", "(objectClass=person)"),
					resource.TestCheckResourceAttr(resourceName, "soft_delete_users", "false"),
				),
			},
			// Disable LDAP
			{
				Config: testAccLDAPConfigDisabled(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
		},
	})
}

func TestAccLDAPConfig_full(t *testing.T) {
	resourceName := "pocketid_ldap_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with all attributes
			{
				Config: testAccLDAPConfigFull(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "url", "ldap://localhost:3890"),
					resource.TestCheckResourceAttr(resourceName, "bind_dn", "cn=admin,dc=example,dc=com"),
					resource.TestCheckResourceAttr(resourceName, "base_dn", "dc=example,dc=com"),
					resource.TestCheckResourceAttr(resourceName, "skip_cert_verify", "true"),
					resource.TestCheckResourceAttr(resourceName, "user_search_filter", "(objectClass=inetOrgPerson)"),
					resource.TestCheckResourceAttr(resourceName, "user_group_search_filter", "(objectClass=groupOfUniqueNames)"),
					resource.TestCheckResourceAttr(resourceName, "soft_delete_users", "false"),
					// User attributes
					resource.TestCheckResourceAttr(resourceName, "user_attributes.unique_identifier", "entryUUID"),
					resource.TestCheckResourceAttr(resourceName, "user_attributes.username", "uid"),
					resource.TestCheckResourceAttr(resourceName, "user_attributes.email", "mail"),
					resource.TestCheckResourceAttr(resourceName, "user_attributes.first_name", "givenName"),
					resource.TestCheckResourceAttr(resourceName, "user_attributes.last_name", "sn"),
					resource.TestCheckResourceAttr(resourceName, "user_attributes.profile_picture", "jpegPhoto"),
					// Group attributes
					resource.TestCheckResourceAttr(resourceName, "group_attributes.member", "uniqueMember"),
					resource.TestCheckResourceAttr(resourceName, "group_attributes.unique_identifier", "entryUUID"),
					resource.TestCheckResourceAttr(resourceName, "group_attributes.name", "cn"),
					resource.TestCheckResourceAttr(resourceName, "group_attributes.admin_group_name", "admins"),
				),
			},
		},
	})
}

func TestAccLDAPConfig_validation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Invalid LDAP URL
			{
				Config:      testAccLDAPConfigInvalidURL(),
				ExpectError: regexp.MustCompile("must be a valid LDAP URL"),
			},
			// Invalid DN format
			{
				Config:      testAccLDAPConfigInvalidDN(),
				ExpectError: regexp.MustCompile("invalid DN format"),
			},
			// Missing required fields when enabled
			{
				Config:      testAccLDAPConfigMissingRequired(),
				ExpectError: regexp.MustCompile("url.*required when LDAP is enabled"),
			},
		},
	})
}

func testAccLDAPConfigBasic() string {
	return testAccProviderConfig() + `
resource "pocketid_ldap_config" "test" {
  enabled = true

  url               = "ldap://localhost:3890"
  bind_dn          = "cn=admin,dc=example,dc=com"
  bind_password    = "password"
  base_dn          = "dc=example,dc=com"
  skip_cert_verify = true

  user_search_filter = "(objectClass=inetOrgPerson)"
}
`
}

func testAccLDAPConfigUpdated() string {
	return testAccProviderConfig() + `
resource "pocketid_ldap_config" "test" {
  enabled = true

  url               = "ldap://localhost:3890"
  bind_dn          = "cn=admin,dc=example,dc=com"
  bind_password    = "password"
  base_dn          = "dc=example,dc=com"
  skip_cert_verify = true

  user_search_filter = "(objectClass=person)"
  soft_delete_users  = false
}
`
}

func testAccLDAPConfigDisabled() string {
	return testAccProviderConfig() + `
resource "pocketid_ldap_config" "test" {
  enabled = false
}
`
}

func testAccLDAPConfigFull() string {
	return testAccProviderConfig() + `
resource "pocketid_ldap_config" "test" {
  enabled = true

  url               = "ldap://localhost:3890"
  bind_dn          = "cn=admin,dc=example,dc=com"
  bind_password    = "password"
  base_dn          = "dc=example,dc=com"
  skip_cert_verify = true

  user_search_filter       = "(objectClass=inetOrgPerson)"
  user_group_search_filter = "(objectClass=groupOfUniqueNames)"

  user_attributes {
    unique_identifier = "entryUUID"
    username         = "uid"
    email           = "mail"
    first_name      = "givenName"
    last_name       = "sn"
    profile_picture = "jpegPhoto"
  }

  group_attributes {
    member             = "uniqueMember"
    unique_identifier  = "entryUUID"
    name              = "cn"
    admin_group_name  = "admins"
  }

  soft_delete_users = false
}
`
}

func testAccLDAPConfigInvalidURL() string {
	return testAccProviderConfig() + `
resource "pocketid_ldap_config" "test" {
  enabled = true
  url     = "http://invalid-url.com"
  bind_dn = "cn=admin,dc=example,dc=com"
  bind_password = "password"
  base_dn = "dc=example,dc=com"
}
`
}

func testAccLDAPConfigInvalidDN() string {
	return testAccProviderConfig() + `
resource "pocketid_ldap_config" "test" {
  enabled = true
  url     = "ldap://localhost:3890"
  bind_dn = "invalid-dn-format"
  bind_password = "password"
  base_dn = "dc=example,dc=com"
}
`
}

func testAccLDAPConfigMissingRequired() string {
	return testAccProviderConfig() + `
resource "pocketid_ldap_config" "test" {
  enabled = true
}
`
}
