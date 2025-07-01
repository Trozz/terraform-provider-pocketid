# Example: Basic OIDC Client Configuration
resource "pocketid_client" "basic_app" {
  name        = "My Basic Application"
  description = "A simple OIDC client for demonstration"

  # Required: At least one redirect URI must be specified
  redirect_uris = [
    "https://myapp.example.com/callback",
    "https://myapp.example.com/oidc/callback"
  ]

  # Default grant types if not specified: ["authorization_code"]
  grant_types = ["authorization_code", "refresh_token"]

  # Default response types if not specified: ["code"]
  response_types = ["code"]

  # Default scopes if not specified: ["openid"]
  scopes = ["openid", "profile", "email"]
}

# Example: Advanced OIDC Client Configuration
resource "pocketid_client" "advanced_app" {
  name        = "My Advanced Application"
  description = "An OIDC client with all configuration options"

  # Multiple redirect URIs for different environments
  redirect_uris = [
    "https://app.example.com/callback",
    "https://staging.app.example.com/callback",
    "http://localhost:3000/callback" # For local development
  ]

  # Post-logout redirect URIs
  post_logout_redirect_uris = [
    "https://app.example.com/",
    "https://app.example.com/logout-success",
    "http://localhost:3000/"
  ]

  # OAuth2 grant types
  grant_types = [
    "authorization_code",
    "refresh_token",
    "implicit" # Use with caution
  ]

  # OAuth2 response types
  response_types = [
    "code",
    "id_token",
    "token"
  ]

  # OIDC scopes - includes custom groups scope
  scopes = [
    "openid",
    "profile",
    "email",
    "groups" # For group membership information
  ]

  # Security settings
  require_consent = true # Require user consent on first login
  require_pkce    = true # Require PKCE for authorization code flow (recommended)

  # Token lifetimes (in seconds)
  access_token_lifetime  = 3600   # 1 hour
  refresh_token_lifetime = 604800 # 7 days
}

# Example: SPA (Single Page Application) Client
resource "pocketid_client" "spa_app" {
  name        = "My SPA Application"
  description = "Single Page Application using PKCE"

  redirect_uris = [
    "https://spa.example.com/auth/callback",
    "http://localhost:8080/auth/callback"
  ]

  # SPA typically uses only authorization code with PKCE
  grant_types    = ["authorization_code"]
  response_types = ["code"]

  # SPAs should always use PKCE
  require_pkce = true

  # SPAs typically don't use refresh tokens for security
  # If you do enable refresh tokens, ensure proper storage
  scopes = ["openid", "profile", "email"]

  # Shorter token lifetime for SPAs
  access_token_lifetime = 900 # 15 minutes
}

# Example: Mobile Application Client
resource "pocketid_client" "mobile_app" {
  name        = "My Mobile Application"
  description = "iOS and Android mobile application"

  # Mobile apps use custom URL schemes or universal links
  redirect_uris = [
    "com.example.myapp://oauth/callback",       # iOS custom scheme
    "https://myapp.example.com/oauth/callback", # Universal link
    "myapp://oauth/callback"                    # Android custom scheme
  ]

  post_logout_redirect_uris = [
    "com.example.myapp://logout",
    "myapp://logout"
  ]

  grant_types = [
    "authorization_code",
    "refresh_token" # Mobile apps typically use refresh tokens
  ]

  response_types = ["code"]

  # Mobile apps must use PKCE
  require_pkce = true

  # All available scopes for mobile
  scopes = ["openid", "profile", "email", "groups"]

  # Longer token lifetimes for mobile apps
  access_token_lifetime  = 3600    # 1 hour
  refresh_token_lifetime = 2592000 # 30 days
}

# Example: Using client credentials in other resources
resource "pocketid_client" "api_client" {
  name          = "API Service Client"
  description   = "Client for service-to-service authentication"
  redirect_uris = ["https://api.example.com/callback"]

  # Minimal configuration for API client
  grant_types = ["client_credentials"]
  scopes      = ["openid"]
}

# Output the client credentials (be careful with sensitive data!)
output "api_client_id" {
  value       = pocketid_client.api_client.client_id
  description = "The client ID for the API service"
}

output "api_client_secret" {
  value       = pocketid_client.api_client.client_secret
  description = "The client secret for the API service"
  sensitive   = true
}

# Example: Using with other resources
locals {
  app_config = {
    oidc = {
      issuer        = "https://pocket-id.example.com"
      client_id     = pocketid_client.advanced_app.client_id
      client_secret = pocketid_client.advanced_app.client_secret
      redirect_uri  = pocketid_client.advanced_app.redirect_uris[0]
    }
  }
}

# You could then use local.app_config to configure your application
# For example, storing in a secret manager or configuration service
