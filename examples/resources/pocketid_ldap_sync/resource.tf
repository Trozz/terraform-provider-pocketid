# Configure LDAP via the application configuration.
resource "pocketid_application_config" "this" {
  app_name     = "My Pocket-ID"
  ldap_enabled = "true"
  ldap_url     = "ldaps://ldap.example.com:636"
  ldap_base    = "dc=example,dc=com"
  ldap_bind_dn = "cn=service,dc=example,dc=com"
  # ... other required configuration ...
}

# Run an LDAP sync whenever the LDAP configuration changes.
# Changing any value in `triggers` forces a new sync.
resource "pocketid_ldap_sync" "this" {
  triggers = {
    ldap_url  = pocketid_application_config.this.ldap_url
    ldap_base = pocketid_application_config.this.ldap_base
  }
}
