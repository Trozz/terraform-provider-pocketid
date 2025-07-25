terraform {
  required_providers {
    pocketid = {
      source = "trozz/pocketid"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
}

# Configure the Pocket-ID Provider
provider "pocketid" {
  base_url  = var.pocketid_base_url
  api_token = var.pocketid_api_token
}

# Variables
variable "pocketid_base_url" {
  description = "The base URL of your Pocket-ID instance"
  type        = string
}

variable "pocketid_api_token" {
  description = "API token for Pocket-ID authentication"
  type        = string
  sensitive   = true
}

# Generate a random prefix for test resources to avoid conflicts
resource "random_string" "prefix" {
  length  = 6
  special = false
  upper   = false
}

# Create user groups
resource "pocketid_group" "developers" {
  name          = "${random_string.prefix.result}-developers"
  friendly_name = "Developers Group"
}

resource "pocketid_group" "admins" {
  name          = "${random_string.prefix.result}-admins"
  friendly_name = "Administrators Group"
}

resource "pocketid_group" "users" {
  name          = "${random_string.prefix.result}-users"
  friendly_name = "Regular Users Group"
}

# Create OIDC clients

# Public SPA client
resource "pocketid_client" "spa_app" {
  name = "${random_string.prefix.result}-React SPA Application"
  callback_urls = [
    "https://spa.example.com/callback",
    "http://localhost:3000/callback"
  ]
  logout_callback_urls = [
    "https://spa.example.com/logout",
    "http://localhost:3000/logout"
  ]
  is_public    = true
  pkce_enabled = true
}

# Confidential web application
resource "pocketid_client" "web_app" {
  name = "${random_string.prefix.result}-Main Web Application"
  callback_urls = [
    "https://app.example.com/auth/callback",
    "https://staging.example.com/auth/callback",
    "http://localhost:8080/auth/callback"
  ]
  logout_callback_urls = [
    "https://app.example.com/logout"
  ]
  is_public    = false
  pkce_enabled = true
}

# Mobile application client
resource "pocketid_client" "mobile_app" {
  name = "${random_string.prefix.result}-Mobile Application"
  callback_urls = [
    "com.example.myapp://callback",
    "myapp://auth/callback"
  ]
  is_public    = true
  pkce_enabled = true
}

# Admin portal with restricted access
resource "pocketid_client" "admin_portal" {
  name = "${random_string.prefix.result}-Admin Portal"
  callback_urls = [
    "https://admin.example.com/callback"
  ]
  is_public    = false
  pkce_enabled = false
  allowed_user_groups = [
    pocketid_group.admins.id
  ]
}

# Create users

# Admin user
resource "pocketid_user" "admin_user" {
  username   = "${random_string.prefix.result}-admin"
  email      = "${random_string.prefix.result}-admin@example.com"
  first_name = "Admin"
  last_name  = "User"
  # disabled = false (default)

  groups = [
    pocketid_group.admins.id
  ]
}

# Developer users
resource "pocketid_user" "dev_lead" {
  username   = "${random_string.prefix.result}-john.doe"
  email      = "${random_string.prefix.result}-john.doe@example.com"
  first_name = "John"
  last_name  = "Doe"
  # disabled = false (default)

  groups = [
    pocketid_group.developers.id,
    pocketid_group.admins.id
  ]
}

resource "pocketid_user" "developer" {
  username   = "${random_string.prefix.result}-jane.smith"
  email      = "${random_string.prefix.result}-jane.smith@example.com"
  first_name = "Jane"
  last_name  = "Smith"
  # disabled = false (default)

  groups = [
    pocketid_group.developers.id
  ]
}

# Regular user
resource "pocketid_user" "regular_user" {
  username   = "${random_string.prefix.result}-bob.wilson"
  email      = "${random_string.prefix.result}-bob.wilson@example.com"
  first_name = "Bob"
  last_name  = "Wilson"
  # disabled = false (default)

  groups = [
    pocketid_group.users.id
  ]
}

# Data sources to query existing resources

# Get a specific client by ID
data "pocketid_client" "existing_client" {
  id = pocketid_client.web_app.id
}

# List all clients
data "pocketid_clients" "all_clients" {}

# Get a specific user by id
data "pocketid_user" "admin" {
  id = pocketid_user.admin_user.id
}

# List all users
data "pocketid_users" "all_users" {}

# List all users (filtering by group not supported in data source)
data "pocketid_users" "developers" {}

# Outputs

output "spa_client_id" {
  description = "Client ID for the SPA application"
  value       = pocketid_client.spa_app.id
}

output "web_app_client_id" {
  description = "Client ID for the web application"
  value       = pocketid_client.web_app.id
}

output "admin_portal_client_id" {
  description = "Client ID for the admin portal"
  value       = pocketid_client.admin_portal.id
}

output "total_clients" {
  description = "Total number of OIDC clients"
  value       = length(data.pocketid_clients.all_clients.clients)
}

output "total_users" {
  description = "Total number of users"
  value       = length(data.pocketid_users.all_users.users)
}

output "developers_count" {
  description = "Number of users in developers group"
  value       = length(data.pocketid_users.developers.users)
}

# Example of creating multiple similar resources
locals {
  test_users = {
    "test1" = { email = "test1@example.com", first_name = "Test", last_name = "User1" }
    "test2" = { email = "test2@example.com", first_name = "Test", last_name = "User2" }
    "test3" = { email = "test3@example.com", first_name = "Test", last_name = "User3" }
  }
}

resource "pocketid_user" "test_users" {
  for_each = local.test_users

  username   = each.key
  email      = each.value.email
  first_name = each.value.first_name
  last_name  = each.value.last_name
  # disabled = false (default)

  groups = [pocketid_group.users.id]
}

# Note: one_time_access_token resource is not available in v0.1.5
# Uncomment when using a newer version that supports this resource
#
# # Create one-time access tokens for initial user setup
# resource "pocketid_one_time_access_token" "admin_token" {
#   user_id       = pocketid_user.admin_user.id
#   expires_at    = timeadd(timestamp(), "24h")
#   skip_recreate = true
# }
#
# resource "pocketid_one_time_access_token" "dev_onboarding" {
#   user_id       = pocketid_user.developer.id
#   expires_at    = timeadd(timestamp(), "168h") # 7 days
#   skip_recreate = true
# }
#
# output "admin_token_value" {
#   description = "One-time access token for admin user (sensitive)"
#   value       = pocketid_one_time_access_token.admin_token.token
#   sensitive   = true
# }
