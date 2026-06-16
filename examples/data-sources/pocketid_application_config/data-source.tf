# Read the global application configuration of the Pocket-ID instance.
data "pocketid_application_config" "current" {}

# Reference configuration values elsewhere.
output "application_name" {
  value = data.pocketid_application_config.current.app_name
}

output "ldap_enabled" {
  value = data.pocketid_application_config.current.ldap_enabled
}
