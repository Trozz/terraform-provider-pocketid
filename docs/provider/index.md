---
page_title: "Provider: Pocket-ID"
description: |-
  The Pocket-ID provider is used to interact with Pocket-ID, a simple and secure passkey-only OIDC provider.
---

# Pocket-ID Provider

The Pocket-ID provider is used to interact with [Pocket-ID](https://github.com/pocket-id/pocket-id), a simple and secure passkey-only OpenID Connect (OIDC) provider. This provider allows you to manage OIDC clients, users, and groups in your Pocket-ID instance through Terraform.

## Key Features

- **OIDC Client Management**: Create and manage OAuth2/OIDC client applications
- **User Management**: Manage users (note: passkey registration must be done through the UI)
- **Group Management**: Create and manage user groups for access control
- **Secure Authentication**: Uses API tokens for secure provider authentication
- **Passkey-First**: Supports Pocket-ID's passkey-only authentication approach

## Example Usage

```hcl
terraform {
  required_providers {
    pocketid = {
      source  = "trozz/pocketid"
      version = "~> 1.0"
    }
  }
}

# Configure the Pocket-ID Provider
provider "pocketid" {
  base_url  = "https://auth.example.com"
  api_token = var.pocketid_api_token
}

# Create an OIDC client
resource "pocketid_client" "web_app" {
  name = "My Web Application"
  callback_urls = [
    "https://app.example.com/callback",
    "http://localhost:3000/callback"
  ]
  logout_callback_urls = [
    "https://app.example.com/logout"
  ]
  is_public    = false
  pkce_enabled = true
}
```

## Authentication

The Pocket-ID provider uses API token authentication. You need to generate an API token from your Pocket-ID instance's admin interface.

### Generating an API Token

1. Log in to your Pocket-ID admin interface
2. Navigate to Settings → API Keys
3. Click "Create New API Key"
4. Give your key a descriptive name (e.g., "Terraform")
5. Copy the generated token - you won't be able to see it again!

### Configuring Authentication

There are several ways to provide authentication credentials:

#### Provider Configuration (Recommended for variables)

```hcl
provider "pocketid" {
  base_url  = "https://auth.example.com"
  api_token = var.pocketid_api_token
}
```

#### Environment Variables (Recommended for CI/CD)

```bash
export POCKETID_BASE_URL="https://auth.example.com"
export POCKETID_API_TOKEN="your-api-token-here"
```

Then your provider configuration can be minimal:

```hcl
provider "pocketid" {
  # Configuration loaded from environment variables
}
```

## Argument Reference

The following arguments are supported in the `provider` block:

* `base_url` - (Optional) The base URL of your Pocket-ID instance. This can also be set via the `POCKETID_BASE_URL` environment variable. Required if not set via environment variable.

* `api_token` - (Optional) The API token for authentication. This can also be set via the `POCKETID_API_TOKEN` environment variable. Required if not set via environment variable. This value is sensitive and should be handled securely.

* `skip_tls_verify` - (Optional) Skip TLS certificate verification. Defaults to `false`. **Warning**: Only use this for development/testing with self-signed certificates. Never use in production.

* `timeout` - (Optional) HTTP client timeout in seconds. Defaults to `30`. Increase this value if you experience timeout errors with slow Pocket-ID instances.

## Environment Variables

The provider supports the following environment variables:

* `POCKETID_BASE_URL` - The base URL of your Pocket-ID instance
* `POCKETID_API_TOKEN` - The API token for authentication

Environment variables take precedence over empty provider configuration but are overridden by explicitly set provider arguments.

## Advanced Configuration Examples

### Development Environment with Self-Signed Certificate

```hcl
provider "pocketid" {
  base_url        = "https://localhost:8443"
  api_token       = var.dev_api_token
  skip_tls_verify = true  # Only for development!
  timeout         = 60    # Longer timeout for slower dev environment
}
```

### Multiple Provider Instances

You can configure multiple Pocket-ID providers for different instances:

```hcl
provider "pocketid" {
  alias     = "production"
  base_url  = "https://auth.example.com"
  api_token = var.prod_api_token
}

provider "pocketid" {
  alias     = "staging"
  base_url  = "https://auth-staging.example.com"
  api_token = var.staging_api_token
}

# Use with alias
resource "pocketid_client" "prod_app" {
  provider = pocketid.production
  name     = "Production App"
  # ...
}

resource "pocketid_client" "staging_app" {
  provider = pocketid.staging
  name     = "Staging App"
  # ...
}
```

### Using with Terraform Cloud

When using with Terraform Cloud, set your credentials as sensitive environment variables:

1. In your Terraform Cloud workspace, go to Variables
2. Add `POCKETID_BASE_URL` as an environment variable
3. Add `POCKETID_API_TOKEN` as a sensitive environment variable
4. Your configuration can then be simple:

```hcl
provider "pocketid" {
  # Credentials are provided by Terraform Cloud environment variables
}
```

## Security Considerations

1. **API Token Security**: Always treat your API token as a secret. Use Terraform variables or environment variables, and never commit tokens to version control.

2. **TLS Verification**: Only disable TLS verification (`skip_tls_verify = true`) in development environments. Always use proper certificates in production.

3. **Least Privilege**: Create API tokens with the minimum required permissions for your Terraform operations.

4. **Token Rotation**: Regularly rotate your API tokens and update your Terraform configurations accordingly.

## Debugging

To enable debug logging for the provider:

```bash
export TF_LOG=DEBUG
export TF_LOG_PATH=terraform-debug.log
terraform apply
```

This will provide detailed information about API requests and responses, which can be helpful for troubleshooting.

## Known Limitations

1. **Passkey Registration**: Users created through Terraform cannot have passkeys registered via the API. Users must complete passkey registration through the Pocket-ID web interface.

2. **Client Secret Retrieval**: OIDC client secrets are only available during resource creation. They cannot be retrieved later, so make sure to store them securely when created.

3. **API Rate Limiting**: The provider implements retry logic for transient failures, but be aware of any rate limits configured on your Pocket-ID instance.

## Getting Help

- **Provider Issues**: [GitHub Issues](https://github.com/trozz/terraform-provider-pocketid/issues)
- **Pocket-ID Issues**: [Pocket-ID GitHub](https://github.com/pocket-id/pocket-id/issues)
- **Documentation**: [Provider Documentation](https://registry.terraform.io/providers/trozz/pocketid/latest/docs)

## Contributing

Contributions are welcome! Please see the [GitHub repository](https://github.com/trozz/terraform-provider-pocketid) for contribution guidelines.