# Outputs for verification and use by data source tests

output "group_ids" {
  description = "IDs of created groups"
  value = {
    admin      = pocketid_group.admin.id
    developers = pocketid_group.developers.id
    users      = pocketid_group.users.id
  }
}

output "user_ids" {
  description = "IDs of created users"
  value = {
    admin    = pocketid_user.admin_user.id
    dev      = pocketid_user.dev_user.id
    disabled = pocketid_user.disabled_user.id
  }
}

output "client_ids" {
  description = "IDs of created OAuth2 clients"
  value = {
    web_app         = pocketid_client.web_app.id
    mobile_app      = pocketid_client.mobile_app.id
    service_account = pocketid_client.service_account.id
    spa_app         = pocketid_client.spa_app.id
  }
}

output "token_info" {
  description = "Information about created tokens"
  value = {
    admin_token_expires = pocketid_one_time_access_token.admin_token.expires_at
    dev_token_expires   = pocketid_one_time_access_token.dev_token.expires_at
  }
  sensitive = true
}
