terraform {
  required_providers {
    pocketid = {
      source  = "Trozz/pocketid"
      version = "~> 0.1.0"
    }
  }
}

provider "pocketid" {
  base_url  = var.pocketid_base_url
  api_token = var.pocketid_api_token
}

# Variables
variable "pocketid_base_url" {
  description = "Base URL for PocketID instance"
  type        = string
  default     = "http://localhost:1411"
}

variable "pocketid_api_token" {
  description = "API token for PocketID authentication"
  type        = string
  sensitive   = true
}

# Groups
resource "pocketid_group" "admin" {
  name          = "test-administrators"
  friendly_name = "Test Administrators"
}

resource "pocketid_group" "developers" {
  name          = "test-developers"
  friendly_name = "Test Developers"
}

resource "pocketid_group" "users" {
  name          = "test-users"
  friendly_name = "Test Users"
}

# Users
resource "pocketid_user" "admin_user" {
  username   = "test-admin"
  email      = "test-admin@example.com"
  first_name = "Test"
  last_name  = "Admin"
  # disabled = false (default)
  groups = [pocketid_group.admin.id]
}

resource "pocketid_user" "dev_user" {
  username   = "test-developer"
  email      = "test-developer@example.com"
  first_name = "Test"
  last_name  = "Developer"
  # disabled = false (default)
  groups = [pocketid_group.developers.id, pocketid_group.users.id]
}

resource "pocketid_user" "disabled_user" {
  username   = "test-disabled"
  email      = "test-disabled@example.com"
  first_name = "Test"
  last_name  = "Disabled"
  disabled   = true
  groups     = [pocketid_group.users.id]
}

# One-time access tokens
resource "pocketid_one_time_access_token" "admin_token" {
  user_id       = pocketid_user.admin_user.id
  expires_at    = timeadd(timestamp(), "24h")
  skip_recreate = true
}

resource "pocketid_one_time_access_token" "dev_token" {
  user_id       = pocketid_user.dev_user.id
  expires_at    = timeadd(timestamp(), "1h")
  skip_recreate = false
}

# OAuth2 Clients
resource "pocketid_client" "web_app" {
  name = "Test Web Application"
  callback_urls = [
    "https://app.example.com/callback",
    "https://app.example.com/auth"
  ]
  logout_callback_urls = [
    "https://app.example.com/logout",
    "https://app.example.com/"
  ]
  is_public    = false
  pkce_enabled = true
}

resource "pocketid_client" "mobile_app" {
  name = "Test Mobile Application"
  callback_urls = [
    "com.example.app://callback",
    "com.example.app://auth"
  ]
  is_public    = true
  pkce_enabled = true
}

resource "pocketid_client" "service_account" {
  name = "Test Service Account"
  callback_urls = [
    "https://localhost/callback"
  ]
  is_public    = false
  pkce_enabled = false
}

resource "pocketid_client" "spa_app" {
  name = "Test Single Page Application"
  callback_urls = [
    "https://spa.example.com/callback"
  ]
  logout_callback_urls = [
    "https://spa.example.com/"
  ]
  is_public    = true
  pkce_enabled = true
}
