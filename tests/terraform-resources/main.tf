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
  name = "test-administrators"
}

resource "pocketid_group" "developers" {
  name = "test-developers"
}

resource "pocketid_group" "users" {
  name = "test-users"
}

# Users
resource "pocketid_user" "admin_user" {
  email      = "test-admin@example.com"
  first_name = "Test"
  last_name  = "Admin"
  enabled    = true
  groups     = [pocketid_group.admin.id]
}

resource "pocketid_user" "dev_user" {
  email      = "test-developer@example.com"
  first_name = "Test"
  last_name  = "Developer"
  enabled    = true
  groups     = [pocketid_group.developers.id, pocketid_group.users.id]
}

resource "pocketid_user" "disabled_user" {
  email      = "test-disabled@example.com"
  first_name = "Test"
  last_name  = "Disabled"
  enabled    = false
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
  client_id                             = "test-web-app"
  name                                  = "Test Web Application"
  grant_types                           = ["authorization_code", "refresh_token"]
  redirect_uris                         = ["https://app.example.com/callback", "https://app.example.com/auth"]
  post_logout_redirect_uris             = ["https://app.example.com/logout", "https://app.example.com/"]
  require_auth_time                     = true
  require_pushed_authorization_requests = false
  allowed_cors_origins                  = ["https://app.example.com"]
  auth_method                           = "client_secret_post"
  scopes                                = ["openid", "profile", "email"]
}

resource "pocketid_client" "mobile_app" {
  client_id     = "test-mobile-app"
  name          = "Test Mobile Application"
  grant_types   = ["authorization_code", "refresh_token"]
  redirect_uris = ["com.example.app://callback", "com.example.app://auth"]
  auth_method   = "none"
  scopes        = ["openid", "profile", "email", "offline_access"]
}

resource "pocketid_client" "service_account" {
  client_id   = "test-service-account"
  name        = "Test Service Account"
  grant_types = ["client_credentials"]
  auth_method = "client_secret_basic"
  scopes      = ["api:read", "api:write"]
}

resource "pocketid_client" "spa_app" {
  client_id                             = "test-spa"
  name                                  = "Test Single Page Application"
  grant_types                           = ["authorization_code"]
  redirect_uris                         = ["https://spa.example.com/callback"]
  post_logout_redirect_uris             = ["https://spa.example.com/"]
  allowed_cors_origins                  = ["https://spa.example.com"]
  auth_method                           = "none"
  scopes                                = ["openid", "profile", "email"]
  require_pushed_authorization_requests = true
}
