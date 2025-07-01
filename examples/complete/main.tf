terraform {
  required_providers {
    pocketid = {
      source  = "trozz/pocketid"
      version = "~> 1.0"
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

# Create user groups
resource "pocketid_group" "developers" {
  name          = "developers"
  friendly_name = "Development Team"
}

resource "pocketid_group" "admins" {
  name          = "admins"
  friendly_name = "System Administrators"
}

resource "pocketid_group" "users" {
  name          = "users"
  friendly_name = "Regular Users"
}

# Create OIDC clients

# Public SPA client
resource "pocketid_client" "spa_app" {
  name = "React SPA Application"
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
  name = "Main Web Application"
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

  # Only developers and admins can access
  allowed_user_groups = [
    pocketid_group.developers.id,
    pocketid_group.admins.id
  ]
}

# Mobile application client
resource "pocketid_client" "mobile_app" {
  name = "Mobile Application"
  callback_urls = [
    "com.example.myapp://callback",
    "myapp://auth/callback"
  ]
  is_public    = true
  pkce_enabled = true
}

# Admin portal with restricted access
resource "pocketid_client" "admin_portal" {
  name = "Admin Portal"
  callback_urls = [
    "https://admin.example.com/callback"
  ]
  is_public    = false
  pkce_enabled = true

  # Only admins can access
  allowed_user_groups = [
    pocketid_group.admins.id
  ]
}

# Create users

# Admin user
resource "pocketid_user" "admin_user" {
  username   = "admin"
  email      = "admin@example.com"
  first_name = "Admin"
  last_name  = "User"
  is_admin   = true

  groups = [
    pocketid_group.admins.id
  ]
}

# Developer users
resource "pocketid_user" "dev_lead" {
  username   = "john.doe"
  email      = "john.doe@example.com"
  first_name = "John"
  last_name  = "Doe"

  groups = [
    pocketid_group.developers.id,
    pocketid_group.admins.id
  ]
}

resource "pocketid_user" "developer" {
  username   = "jane.smith"
  email      = "jane.smith@example.com"
  first_name = "Jane"
  last_name  = "Smith"

  groups = [
    pocketid_group.developers.id
  ]
}

# Regular user
resource "pocketid_user" "regular_user" {
  username   = "bob.wilson"
  email      = "bob.wilson@example.com"
  first_name = "Bob"
  last_name  = "Wilson"
  locale     = "en-US"

  groups = [
    pocketid_group.users.id
  ]
}

# Data sources to query existing resources

# Get a specific client by ID
data "pocketid_client" "existing_client" {
  client_id = pocketid_client.web_app.id
}

# List all clients
data "pocketid_clients" "all_clients" {}

# Get a specific user
data "pocketid_user" "admin" {
  username = pocketid_user.admin_user.username
}

# List all users
data "pocketid_users" "all_users" {}

# List users in a specific group
data "pocketid_users" "developers" {
  group_id = pocketid_group.developers.id
}

# Outputs

output "spa_client_id" {
  description = "Client ID for the SPA application"
  value       = pocketid_client.spa_app.id
}

output "web_app_client_id" {
  description = "Client ID for the web application"
  value       = pocketid_client.web_app.id
}

output "web_app_client_secret" {
  description = "Client secret for the web application (only available on creation)"
  value       = pocketid_client.web_app.client_secret
  sensitive   = true
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

  groups = [pocketid_group.users.id]
}
