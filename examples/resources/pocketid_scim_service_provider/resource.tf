# Configure SCIM provisioning for an OIDC client.
# Pocket-ID will provision users and groups to the external SCIM endpoint.
resource "pocketid_client" "example" {
  name          = "SCIM Enabled App"
  callback_urls = ["https://app.example.com/callback"]
}

resource "pocketid_scim_service_provider" "example" {
  client_id = pocketid_client.example.id
  endpoint  = "https://app.example.com/scim/v2"
  token     = var.scim_bearer_token
}

# The bearer token is sensitive and is read back from the API on refresh.
variable "scim_bearer_token" {
  description = "Bearer token used to authenticate against the SCIM endpoint"
  type        = string
  sensitive   = true
}

# Output the SCIM service provider ID (the token is sensitive)
output "scim_service_provider_id" {
  description = "The ID of the SCIM service provider configuration"
  value       = pocketid_scim_service_provider.example.id
}
