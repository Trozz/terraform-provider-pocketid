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

# Create a basic OIDC client
resource "pocketid_client" "example" {
  name                 = var.client_name
  callback_urls        = var.callback_urls
  logout_callback_urls = var.logout_callback_urls

  # Optional: Add allowed user groups to restrict access
  # allowed_user_groups = ["developers", "admins"]
}

# Data source to verify the client was created
data "pocketid_client" "example" {
  client_id = pocketid_client.example.id

  depends_on = [pocketid_client.example]
}
