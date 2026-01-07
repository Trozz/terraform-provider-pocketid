# Trigger LDAP sync on every apply
resource "pocketid_ldap_sync" "sync" {
  triggers = {
    timestamp = timestamp()
  }
}

# Trigger LDAP sync only when config changes
resource "pocketid_ldap_sync" "on_config_change" {
  triggers = {
    config_id = pocketid_ldap_config.main.id
  }
}

# Manual sync trigger (change the value to trigger)
resource "pocketid_ldap_sync" "manual" {
  triggers = {
    manual = "change-this-to-trigger-sync"
  }
}
