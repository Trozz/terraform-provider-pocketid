# Outputs to verify data source functionality

output "admin_user_lookup" {
  description = "Admin user details from lookup"
  value = {
    id         = data.pocketid_user.lookup_admin.id
    email      = data.pocketid_user.lookup_admin.email
    first_name = data.pocketid_user.lookup_admin.first_name
    last_name  = data.pocketid_user.lookup_admin.last_name
    enabled    = data.pocketid_user.lookup_admin.enabled
    groups     = data.pocketid_user.lookup_admin.groups
  }
}

output "dev_user_lookup" {
  description = "Developer user details from lookup"
  value = {
    id         = data.pocketid_user.lookup_dev.id
    email      = data.pocketid_user.lookup_dev.email
    first_name = data.pocketid_user.lookup_dev.first_name
    last_name  = data.pocketid_user.lookup_dev.last_name
    enabled    = data.pocketid_user.lookup_dev.enabled
    groups     = data.pocketid_user.lookup_dev.groups
  }
}

output "all_users_summary" {
  description = "Summary of all users"
  value = {
    total_count = length(data.pocketid_users.all_users.users)
    emails      = [for user in data.pocketid_users.all_users.users : user.email]
  }
}

output "user_counts_by_status" {
  description = "User counts by enabled status"
  value = {
    total    = length(data.pocketid_users.all_users.users)
    enabled  = length(data.pocketid_users.enabled_users.users)
    disabled = length(data.pocketid_users.disabled_users.users)
  }
}

output "all_groups_summary" {
  description = "Summary of all groups"
  value = {
    total_count = length(data.pocketid_groups.all_groups.groups)
    names       = [for group in data.pocketid_groups.all_groups.groups : group.name]
  }
}

output "all_clients_summary" {
  description = "Summary of all OAuth2 clients"
  value = {
    total_count = length(data.pocketid_clients.all_clients.clients)
    names       = [for client in data.pocketid_clients.all_clients.clients : client.name]
    client_ids  = [for client in data.pocketid_clients.all_clients.clients : client.client_id]
  }
}

output "web_client_details" {
  description = "Web client configuration from lookup"
  value = {
    id            = data.pocketid_client.lookup_web_client.id
    client_id     = data.pocketid_client.lookup_web_client.client_id
    name          = data.pocketid_client.lookup_web_client.name
    grant_types   = data.pocketid_client.lookup_web_client.grant_types
    redirect_uris = data.pocketid_client.lookup_web_client.redirect_uris
    scopes        = data.pocketid_client.lookup_web_client.scopes
  }
}
