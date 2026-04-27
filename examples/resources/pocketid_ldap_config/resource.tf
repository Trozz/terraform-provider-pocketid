# Basic LDAP configuration
resource "pocketid_ldap_config" "main" {
  enabled        = true
  sync_on_change = true

  # Connection settings
  url              = "ldaps://ldap.example.com:636"
  bind_dn          = "cn=admin,dc=example,dc=com"
  bind_password    = var.ldap_bind_password
  base_dn          = "dc=example,dc=com"
  skip_cert_verify = false

  # Search filters
  user_search_filter       = "(objectClass=person)"
  user_group_search_filter = "(objectClass=groupOfNames)"

  # User attribute mappings
  user_attributes {
    unique_identifier = "objectGUID"
    username          = "sAMAccountName"
    email             = "mail"
    first_name        = "givenName"
    last_name         = "sn"
  }

  # Group attribute mappings
  group_attributes {
    member            = "member"
    unique_identifier = "objectGUID"
    name              = "cn"
    admin_group       = "PocketID-Admins"
  }

  # Behavior settings
  soft_delete_users = true
}

variable "ldap_bind_password" {
  type      = string
  sensitive = true
}
