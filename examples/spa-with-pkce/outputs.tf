output "client_id" {
  description = "The client ID to use in your SPA"
  value       = pocketid_client.spa.id
}

output "client_name" {
  description = "The name of the client"
  value       = pocketid_client.spa.name
}

output "oidc_configuration" {
  description = "OIDC configuration for your SPA"
  value = {
    authority     = var.base_url
    client_id     = pocketid_client.spa.id
    redirect_uri  = pocketid_client.spa.callback_urls[0]
    scope         = "openid profile email"
    response_type = "code"
  }
}

output "integration_example" {
  description = "Example JavaScript configuration for oidc-client-js"
  value       = <<-EOT
    const oidcConfig = {
      authority: '${var.base_url}',
      client_id: '${pocketid_client.spa.id}',
      redirect_uri: '${pocketid_client.spa.callback_urls[0]}',
      post_logout_redirect_uri: '${pocketid_client.spa.logout_callback_urls[0]}',
      response_type: 'code',
      scope: 'openid profile email',
      // PKCE is automatically handled by the library
    };
  EOT
}
