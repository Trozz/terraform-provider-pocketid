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
  name = "${random_string.prefix.result}-developers"
}

resource "pocketid_group" "admins" {
  name = "${random_string.prefix.result}-admins"
}

resource "pocketid_group" "users" {
  name = "${random_string.prefix.result}-users"
}

# Create OIDC clients

# Public SPA client
resource "pocketid_client" "spa_app" {
  client_id = "${random_string.prefix.result}-spa-app"
  name      = "${random_string.prefix.result}-React SPA Application"
  redirect_uris = [
    "https://spa.example.com/callback",
    "http://localhost:3000/callback"
  ]
  post_logout_redirect_uris = [
    "https://spa.example.com/logout",
    "http://localhost:3000/logout"
  ]
  grant_types = ["authorization_code"]
  auth_method = "none"
  scopes      = ["openid", "profile", "email"]
  allowed_cors_origins = [
    "https://spa.example.com",
    "http://localhost:3000"
  ]
  require_pushed_authorization_requests = true
}

# Confidential web application
resource "pocketid_client" "web_app" {
  client_id = "${random_string.prefix.result}-web-app"
  name      = "${random_string.prefix.result}-Main Web Application"
  redirect_uris = [
    "https://app.example.com/auth/callback",
    "https://staging.example.com/auth/callback",
    "http://localhost:8080/auth/callback"
  ]
  post_logout_redirect_uris = [
    "https://app.example.com/logout"
  ]
  grant_types          = ["authorization_code", "refresh_token"]
  auth_method          = "client_secret_post"
  scopes               = ["openid", "profile", "email"]
  require_auth_time    = true
  allowed_cors_origins = ["https://app.example.com"]
}

# Mobile application client
resource "pocketid_client" "mobile_app" {
  client_id = "${random_string.prefix.result}-mobile-app"
  name      = "${random_string.prefix.result}-Mobile Application"
  redirect_uris = [
    "com.example.myapp://callback",
    "myapp://auth/callback"
  ]
  grant_types                           = ["authorization_code", "refresh_token"]
  auth_method                           = "none"
  scopes                                = ["openid", "profile", "email", "offline_access"]
  require_pushed_authorization_requests = true
}

# Admin portal with restricted access
resource "pocketid_client" "admin_portal" {
  client_id = "${random_string.prefix.result}-admin-portal"
  name      = "${random_string.prefix.result}-Admin Portal"
  redirect_uris = [
    "https://admin.example.com/callback"
  ]
  grant_types       = ["authorization_code", "refresh_token"]
  auth_method       = "client_secret_basic"
  scopes            = ["openid", "profile", "email", "groups"]
  require_auth_time = true
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

# Get a specific user by username
data "pocketid_user" "admin" {
  username = pocketid_user.admin_user.username
}

# List all users
data "pocketid_users" "all_users" {}

# List all users (filtering by group not supported in data source)
data "pocketid_users" "developers" {}

# Outputs

output "spa_client_id" {
  description = "Client ID for the SPA application"
  value       = pocketid_client.spa_app.client_id
}

output "web_app_client_id" {
  description = "Client ID for the web application"
  value       = pocketid_client.web_app.client_id
}

output "admin_portal_client_id" {
  description = "Client ID for the admin portal"
  value       = pocketid_client.admin_portal.client_id
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

# Create one-time access tokens for initial user setup
resource "pocketid_one_time_access_token" "admin_token" {
  user_id       = pocketid_user.admin_user.id
  expires_at    = timeadd(timestamp(), "24h")
  skip_recreate = true
}

resource "pocketid_one_time_access_token" "dev_onboarding" {
  user_id       = pocketid_user.developer.id
  expires_at    = timeadd(timestamp(), "168h") # 7 days
  skip_recreate = true
}

output "admin_token_value" {
  description = "One-time access token for admin user (sensitive)"
  value       = pocketid_one_time_access_token.admin_token.token
  sensitive   = true
}
