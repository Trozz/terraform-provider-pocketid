output "client_id" {
  description = "The generated client ID"
  value       = pocketid_client.example.id
}

output "client_secret" {
  description = "The generated client secret (sensitive)"
  value       = pocketid_client.example.client_secret
  sensitive   = true
}

output "client_name" {
  description = "The name of the client"
  value       = pocketid_client.example.name
}

output "is_public" {
  description = "Whether this is a public client"
  value       = pocketid_client.example.is_public
}

output "pkce_enabled" {
  description = "Whether PKCE is enabled"
  value       = pocketid_client.example.pkce_enabled
}
