# Test data sources after resources are created

# Data source: Fetch a specific OIDC client by ID
data "pocketid_client" "example_client" {
  id = pocketid_client.example.id
}

# Data source: Fetch all OIDC clients
data "pocketid_clients" "all" {}

# Data source: Fetch a specific user by ID
data "pocketid_user" "admin" {
  id = pocketid_user.admin_user.id
}

# Data source: Fetch all users
data "pocketid_users" "all" {}

# Output data source results
output "data_source_client_info" {
  value = {
    id           = data.pocketid_client.example_client.id
    name         = data.pocketid_client.example_client.name
    is_public    = data.pocketid_client.example_client.is_public
    pkce_enabled = data.pocketid_client.example_client.pkce_enabled
  }
  description = "Information about the example client from data source"
}

output "all_clients_count" {
  value       = length(data.pocketid_clients.all.clients)
  description = "Total number of OIDC clients"
}

output "all_clients_names" {
  value       = [for client in data.pocketid_clients.all.clients : client.name]
  description = "Names of all OIDC clients"
}

output "data_source_user_info" {
  value = {
    id       = data.pocketid_user.admin.id
    username = data.pocketid_user.admin.username
    email    = data.pocketid_user.admin.email
    is_admin = data.pocketid_user.admin.is_admin
  }
  description = "Information about the admin user from data source"
}

output "all_users_count" {
  value       = length(data.pocketid_users.all.users)
  description = "Total number of users"
}

output "all_users_summary" {
  value = [
    for user in data.pocketid_users.all.users : {
      username = user.username
      email    = user.email
      is_admin = user.is_admin
      disabled = user.disabled
    }
  ]
  description = "Summary of all users"
}

# Example: Find users in a specific group
output "users_in_developers_group" {
  value = [
    for user in data.pocketid_users.all.users : user.username
    if try(contains(user.groups, pocketid_group.developers.id), false)
  ]
  description = "Usernames of users in the developers group"
}

# Example: Find public clients
output "public_clients" {
  value = [
    for client in data.pocketid_clients.all.clients : {
      name = client.name
      id   = client.id
    }
    if client.is_public
  ]
  description = "List of public OIDC clients"
}

# Example: Find clients with group restrictions
output "restricted_clients" {
  value = [
    for client in data.pocketid_clients.all.clients : {
      name        = client.name
      group_count = try(length(client.allowed_user_groups), 0)
    }
    if try(length(client.allowed_user_groups) > 0, false)
  ]
  description = "Clients with group access restrictions"
}
