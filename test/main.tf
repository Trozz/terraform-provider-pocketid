terraform {
  required_providers {
    pocketid = {
      source  = "registry.terraform.io/trozz/pocketid"
      version = "0.0.1"
    }
  }
}

provider "pocketid" {
  base_url  = "https://local.leer.dev"
  api_token = "tBwTzY4xjg2Apq1oCB9REpzihmN5dm7v"
}

# Create user groups
resource "pocketid_group" "admins" {
  name          = "admins"
  friendly_name = "Administrators"
}

resource "pocketid_group" "developers" {
  name          = "developers"
  friendly_name = "Development Team"
}

resource "pocketid_group" "users" {
  name          = "users"
  friendly_name = "Regular Users"
}

# Create users
resource "pocketid_user" "admin_user" {
  username   = "admin"
  email      = "admin@example.com"
  first_name = "Admin"
  last_name  = "User"
  is_admin   = true
  groups     = [pocketid_group.admins.id]
}

resource "pocketid_user" "john_doe" {
  username   = "john.doe"
  email      = "john.doe@example.com"
  first_name = "John"
  last_name  = "Doe"
  is_admin   = false
  groups = [
    pocketid_group.developers.id,
    pocketid_group.users.id
  ]
}

resource "pocketid_user" "jane_smith" {
  username   = "jane.smith"
  email      = "jane.smith@example.com"
  first_name = "Jane"
  last_name  = "Smith"
  disabled   = true # Example of a disabled user
  groups     = [pocketid_group.users.id]
}

# Create a simple OIDC client
resource "pocketid_client" "example" {
  name = "Terraform Test Client"

  callback_urls = [
    "https://example.com/callback",
    "https://example.com/oidc/callback"
  ]

  logout_callback_urls = [
    "https://example.com/logout"
  ]

  is_public    = false
  pkce_enabled = true
}

# Create a public client (SPA)
resource "pocketid_client" "spa_client" {
  name = "Test SPA Client"

  callback_urls = [
    "http://localhost:3000/callback",
    "https://spa.example.com/callback"
  ]

  is_public    = true
  pkce_enabled = true
}

# Create a client restricted to specific groups
resource "pocketid_client" "restricted_client" {
  name = "Restricted App"

  callback_urls = [
    "https://restricted.example.com/callback"
  ]

  is_public    = false
  pkce_enabled = true

  allowed_user_groups = [
    pocketid_group.admins.id,
    pocketid_group.developers.id
  ]
}

# Output group information
output "admin_group_id" {
  value       = pocketid_group.admins.id
  description = "The ID of the administrators group"
}

output "developer_users" {
  value = {
    group_id = pocketid_group.developers.id
    users    = [pocketid_user.john_doe.username]
  }
  description = "Developers group information"
}

# Output user information
output "admin_user_info" {
  value = {
    id       = pocketid_user.admin_user.id
    username = pocketid_user.admin_user.username
    email    = pocketid_user.admin_user.email
  }
  description = "Admin user information"
}

# Output client details
output "example_client_id" {
  value       = pocketid_client.example.id
  description = "The ID of the example client"
}

output "example_client_secret" {
  value       = pocketid_client.example.client_secret
  sensitive   = true
  description = "The secret of the example client"
}

output "spa_client_id" {
  value       = pocketid_client.spa_client.id
  description = "The ID of the SPA client"
}

output "restricted_client_info" {
  value = {
    id                  = pocketid_client.restricted_client.id
    allowed_group_count = length(pocketid_client.restricted_client.allowed_user_groups)
  }
  description = "Restricted client information"
}
