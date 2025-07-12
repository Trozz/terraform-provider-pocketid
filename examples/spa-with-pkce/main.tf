terraform {
  required_version = ">= 1.0"
  required_providers {
    pocketid = {
      source = "trozz/pocketid"
    }
  }
}

provider "pocketid" {
  base_url  = var.base_url
  api_token = var.api_token
}

# Create a public OIDC client for SPA with PKCE
resource "pocketid_client" "spa" {
  name = var.app_name

  # Public client for browser-based applications
  is_public = true

  # Enable PKCE for enhanced security
  pkce_enabled = true

  # Callback URLs for OAuth2 flow
  callback_urls = [
    for url in var.app_urls : "${url}/auth/callback"
  ]

  # Logout callback URLs
  logout_callback_urls = var.app_urls

  # Optional: Restrict to specific user groups
  # allowed_user_groups = var.allowed_groups
}

# Verify the client configuration
data "pocketid_client" "spa" {
  client_id = pocketid_client.spa.id

  depends_on = [pocketid_client.spa]
}
