terraform {
  required_providers {
    pocketid = {
      source = "Trozz/pocketid"
    }
  }
}

provider "pocketid" {
  # Configuration can be provided via environment variables:
  # POCKETID_BASE_URL
  # POCKETID_API_TOKEN
}

# Example 1: Basic LDAP configuration
resource "pocketid_ldap_config" "basic" {
  enabled = true

  url           = "ldap://ldap.example.com:389"
  bind_dn       = "cn=admin,dc=example,dc=com"
  bind_password = "secure-password"
  base_dn       = "dc=example,dc=com"

  user_attributes {
    unique_identifier = "uid"
    username          = "uid"
  }
}

# Example 2: Active Directory configuration
resource "pocketid_ldap_config" "active_directory" {
  enabled = true

  # LDAPS for secure connection
  url              = "ldaps://ad.company.com:636"
  bind_dn          = "CN=LDAP Reader,OU=Service Accounts,DC=company,DC=com"
  bind_password    = var.ad_bind_password
  base_dn          = "DC=company,DC=com"
  skip_cert_verify = false

  # AD-specific search filters
  user_search_filter       = "(&(objectClass=user)(objectCategory=person))"
  user_group_search_filter = "(objectClass=group)"

  # Keep disabled users instead of deleting
  soft_delete_users = true

  user_attributes {
    unique_identifier = "objectGUID"
    username          = "sAMAccountName"
    email             = "mail"
    first_name        = "givenName"
    last_name         = "sn"
    profile_picture   = "thumbnailPhoto"
  }

  group_attributes {
    member            = "member"
    unique_identifier = "objectGUID"
    name              = "cn"
    admin_group_name  = "PocketID-Admins"
  }
}

# Example 3: Test LDAP before enabling
data "pocketid_ldap_test" "validation" {}

output "ldap_test_results" {
  value = {
    connected     = data.pocketid_ldap_test.validation.connection_successful
    authenticated = data.pocketid_ldap_test.validation.bind_successful
    base_dn_valid = data.pocketid_ldap_test.validation.base_dn_found
    users_found   = data.pocketid_ldap_test.validation.users_found
    groups_found  = data.pocketid_ldap_test.validation.groups_found
  }
}

# Example 4: Conditional LDAP configuration
resource "pocketid_ldap_config" "conditional" {
  count = data.pocketid_ldap_test.validation.connection_successful ? 1 : 0

  enabled = true

  url           = var.ldap_url
  bind_dn       = var.ldap_bind_dn
  bind_password = var.ldap_bind_password
  base_dn       = var.ldap_base_dn

  user_attributes {
    unique_identifier = "uid"
    username          = "uid"
    email             = "mail"
  }
}

# Example 5: Automatic sync on configuration changes
resource "pocketid_ldap_sync" "auto_sync" {
  depends_on = [pocketid_ldap_config.basic]

  triggers = {
    # Sync whenever configuration changes
    config_change = pocketid_ldap_config.basic.id
    ldap_url      = pocketid_ldap_config.basic.url
    base_dn       = pocketid_ldap_config.basic.base_dn
  }
}

# Example 6: Manual sync trigger
resource "pocketid_ldap_sync" "manual" {
  triggers = {
    # Change this value to force a new sync
    manual_trigger = "2024-01-15-v1"
  }
}

# Variables for sensitive data
variable "ad_bind_password" {
  description = "Password for Active Directory bind DN"
  type        = string
  sensitive   = true
}

variable "ldap_url" {
  description = "LDAP server URL"
  type        = string
  default     = "ldap://ldap.example.com:389"
}

variable "ldap_bind_dn" {
  description = "LDAP bind DN"
  type        = string
  default     = "cn=admin,dc=example,dc=com"
}

variable "ldap_bind_password" {
  description = "LDAP bind password"
  type        = string
  sensitive   = true
}

variable "ldap_base_dn" {
  description = "LDAP base DN"
  type        = string
  default     = "dc=example,dc=com"
}
