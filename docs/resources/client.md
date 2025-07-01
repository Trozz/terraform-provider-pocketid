---
page_title: "pocketid_client Resource - terraform-provider-pocketid"
subcategory: ""
description: |-
  Manages an OIDC client in Pocket-ID.
---

# pocketid_client (Resource)

Manages an OIDC client in Pocket-ID. OIDC clients are applications that can authenticate users through Pocket-ID using the OAuth2/OpenID Connect protocol.

~> **Note** The client secret is only available during resource creation and cannot be retrieved later. Store it securely in your secrets management system immediately after creation.

## Example Usage

### Basic Web Application

```hcl
resource "pocketid_client" "web_app" {
  name = "My Web Application"
  callback_urls = [
    "https://app.example.com/auth/callback",
    "http://localhost:3000/auth/callback"
  ]
}

# Output the client secret (only available on creation)
output "client_secret" {
  value     = pocketid_client.web_app.client_secret
  sensitive = true
}
```

### Single Page Application (SPA)

```hcl
resource "pocketid_client" "spa" {
  name = "My React App"
  callback_urls = [
    "https://spa.example.com/callback",
    "http://localhost:3000/callback"
  ]
  is_public    = true  # Public client for SPA
  pkce_enabled = true  # PKCE is required for public clients
}
```

### Mobile Application

```hcl
resource "pocketid_client" "mobile_app" {
  name = "My Mobile App"
  callback_urls = [
    "com.example.myapp://callback",
    "myapp://auth/callback"
  ]
  logout_callback_urls = [
    "com.example.myapp://logout"
  ]
  is_public    = true
  pkce_enabled = true
}
```

### Restricted Client with Allowed Groups

```hcl
resource "pocketid_group" "developers" {
  name          = "developers"
  friendly_name = "Development Team"
}

resource "pocketid_group" "admins" {
  name          = "admins"
  friendly_name = "Administrators"
}

resource "pocketid_client" "internal_app" {
  name = "Internal Admin Portal"
  callback_urls = [
    "https://admin.internal.example.com/callback"
  ]
  
  # Only users in these groups can use this client
  allowed_user_groups = [
    pocketid_group.developers.id,
    pocketid_group.admins.id
  ]
}
```

### Complete Example with All Options

```hcl
resource "pocketid_client" "complete_example" {
  name = "Complete Example App"
  
  # Required: At least one callback URL
  callback_urls = [
    "https://app.example.com/callback",
    "https://app.example.com/auth/callback",
    "http://localhost:8080/callback"
  ]
  
  # Optional: Logout callbacks
  logout_callback_urls = [
    "https://app.example.com/logout",
    "http://localhost:8080/logout"
  ]
  
  # Optional: Client type configuration
  is_public    = false  # Default: false (confidential client)
  pkce_enabled = true   # Default: true (recommended for security)
  
  # Optional: Access restrictions
  allowed_user_groups = [
    pocketid_group.developers.id
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The display name of the OIDC client. This name is shown to users during the authentication process.

* `callback_urls` - (Required) List of allowed callback URLs for the OIDC client. At least one URL must be specified. These URLs are where Pocket-ID will redirect users after successful authentication. Must be valid URLs (except for mobile app deep links).

* `logout_callback_urls` - (Optional) List of allowed logout callback URLs for the OIDC client. These URLs are where Pocket-ID can redirect users after logout.

* `is_public` - (Optional) Whether this is a public client (no client secret). Defaults to `false`. Set to `true` for single-page applications (SPAs) and mobile apps that cannot securely store secrets.

* `pkce_enabled` - (Optional) Whether PKCE (Proof Key for Code Exchange) is enabled for this client. Defaults to `true`. PKCE is highly recommended for all clients and required for public clients.

* `allowed_user_groups` - (Optional) List of user group IDs that are allowed to use this client. If empty or not specified, all users can use this client. Use this to restrict access to specific groups of users.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier of the OIDC client. This is used as the `client_id` in OAuth2/OIDC flows.

* `client_secret` - The client secret for confidential clients. This value is sensitive and only available during resource creation. It cannot be retrieved afterward. This attribute is `null` for public clients.

* `has_logo` - Whether the client has a logo configured. Currently, logos must be uploaded through the Pocket-ID web interface.

## Import

OIDC clients can be imported using their ID:

```bash
terraform import pocketid_client.example <client-id>
```

For example:

```bash
terraform import pocketid_client.web_app 550e8400-e29b-41d4-a716-446655440000
```

~> **Note** When importing a client, the `client_secret` attribute will not be populated as it cannot be retrieved from the API. If you need the client secret, you'll need to generate a new one through the Pocket-ID interface or recreate the resource.

## Client Types and Security Considerations

### Confidential Clients (is_public = false)

Confidential clients can securely store credentials. Use this type for:
- Server-side web applications
- Backend services
- APIs that authenticate with Pocket-ID

**Security notes:**
- Store the client secret securely (e.g., in a secrets manager)
- Never expose the client secret in client-side code
- Rotate secrets periodically

### Public Clients (is_public = true)

Public clients cannot securely store credentials. Use this type for:
- Single-page applications (SPAs)
- Mobile applications
- Native desktop applications

**Security notes:**
- Always enable PKCE for public clients
- Use redirect URI validation to prevent authorization code interception
- Consider implementing refresh token rotation

## Redirect URI Guidelines

### Web Applications
- Always use HTTPS in production: `https://app.example.com/callback`
- Include localhost for development: `http://localhost:3000/callback`
- Be specific with paths to improve security

### Mobile Applications
- Use custom URL schemes: `com.example.app://callback`
- Consider universal links (iOS) or app links (Android)
- Register all possible callback formats

### Common Patterns
```hcl
# Development and production URLs
callback_urls = [
  "https://app.example.com/auth/callback",      # Production
  "https://staging.example.com/auth/callback",  # Staging  
  "http://localhost:3000/auth/callback",        # Local development
  "http://localhost:8080/auth/callback"         # Alternative port
]
```

## Working with Groups

To restrict client access to specific user groups:

1. Create the groups first
2. Reference them in the client configuration
3. Users must be members of at least one allowed group to authenticate

```hcl
# Create groups
resource "pocketid_group" "engineering" {
  name          = "engineering"
  friendly_name = "Engineering Team"
}

resource "pocketid_group" "product" {
  name          = "product"
  friendly_name = "Product Team"
}

# Create a client restricted to these groups
resource "pocketid_client" "internal_tool" {
  name          = "Internal Development Tool"
  callback_urls = ["https://devtool.internal/callback"]
  
  allowed_user_groups = [
    pocketid_group.engineering.id,
    pocketid_group.product.id
  ]
}
```

## Troubleshooting

### Client Secret Not Available
The client secret is only available when the resource is created. If you've lost it:
1. Generate a new secret through the Pocket-ID web interface, or
2. Delete and recreate the Terraform resource

### Redirect URI Mismatch Errors
Ensure your application's redirect URI exactly matches one in the `callback_urls` list, including:
- Protocol (http vs https)
- Domain and subdomain
- Port number
- Path

### Group Restrictions Not Working
- Verify the user is a member of at least one allowed group
- Check that group IDs are correct
- Remember that empty `allowed_user_groups` means all users are allowed