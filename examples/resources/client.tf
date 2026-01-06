# Example: Basic OIDC Client Configuration
resource "pocketid_client" "basic_app" {
  client_id = "my-basic-app"
  name      = "My Basic Application"

  # Required: At least one redirect URI must be specified
  redirect_uris = [
    "https://myapp.example.com/callback",
    "https://myapp.example.com/oidc/callback"
  ]

  # Grant types
  grant_types = ["authorization_code", "refresh_token"]

  # Scopes
  scopes = ["openid", "profile", "email"]

  # Authentication method
  auth_method = "client_secret_post"
}

# Example: Advanced OIDC Client Configuration
resource "pocketid_client" "advanced_app" {
  client_id = "my-advanced-app"
  name      = "My Advanced Application"

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

  # OIDC scopes - includes custom groups scope
  scopes = [
    "openid",
    "profile",
    "email",
    "groups" # For group membership information
  ]

  # Security settings
  require_auth_time                     = true
  require_pushed_authorization_requests = false
  requires_reauthentication = true

  # CORS origins for SPAs
  allowed_cors_origins = [
    "https://app.example.com",
    "https://staging.app.example.com",
    "http://localhost:3000"
  ]

  # Authentication method
  auth_method = "client_secret_basic"
}

# Example: SPA (Single Page Application) Client
resource "pocketid_client" "spa_app" {
  client_id = "my-spa-app"
  name      = "My SPA Application"

  redirect_uris = [
    "https://spa.example.com/auth/callback",
    "http://localhost:8080/auth/callback"
  ]

  # SPA typically uses only authorization code with PKCE
  grant_types = ["authorization_code"]

  # SPAs use public client authentication (no secret)
  auth_method = "none"

  # SPAs should always use PKCE
  require_pushed_authorization_requests = true

  # SPAs typically don't use refresh tokens for security
  # If you do enable refresh tokens, ensure proper storage
  scopes = ["openid", "profile", "email"]

  # CORS settings for browser-based apps
  allowed_cors_origins = [
    "https://spa.example.com",
    "http://localhost:8080"
  ]
  launch_url = "https://spa.example.com/launch"
}

# Example: Mobile Application Client
resource "pocketid_client" "mobile_app" {
  client_id = "my-mobile-app"
  name      = "My Mobile Application"

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

  # Mobile apps use public client authentication (no secret)
  auth_method = "none"

  # Mobile apps should use PKCE
  require_pushed_authorization_requests = true

  # All available scopes for mobile
  scopes = ["openid", "profile", "email", "groups", "offline_access"]
}

# Example: Service-to-Service Client
resource "pocketid_client" "api_client" {
  client_id = "api-service-client"
  name      = "API Service Client"

  # Service clients typically don't need redirect URIs but provider may require it
  redirect_uris = ["https://api.example.com/callback"]

  # Service-to-service authentication
  grant_types = ["client_credentials"]
  auth_method = "client_secret_basic"
  scopes      = ["api:read", "api:write"]
}

# Output the client ID (client_secret is sensitive and write-only)
output "api_client_id" {
  value       = pocketid_client.api_client.client_id
  description = "The client ID for the API service"
}

# Example: Using with other resources
locals {
  app_config = {
    oidc = {
      issuer       = "https://pocket-id.example.com"
      client_id    = pocketid_client.advanced_app.client_id
      redirect_uri = pocketid_client.advanced_app.redirect_uris[0]
      scopes       = join(" ", pocketid_client.advanced_app.scopes)
    }
  }
}

# You could then use local.app_config to configure your application
# For example, storing in a secret manager or configuration service

# Example: Client with specific user group restrictions
resource "pocketid_client" "restricted_app" {
  client_id = "restricted-admin-app"
  name      = "Restricted Admin Application"

  redirect_uris = ["https://admin.example.com/callback"]
  grant_types   = ["authorization_code", "refresh_token"]
  auth_method   = "client_secret_post"
  scopes        = ["openid", "profile", "email", "groups"]

  # Restrict access to specific groups (requires group IDs)
  # allowed_user_groups = [pocketid_group.admins.id]
  # Optional: require reauthentication for admin/restricted clients
  requires_reauthentication = true
}
