# Single Page Application (SPA) with PKCE Example

This example demonstrates how to configure an OIDC client for a browser-based Single Page Application using PKCE
(Proof Key for Code Exchange).

## Overview

This example creates:

- A public OIDC client suitable for SPAs
- PKCE enabled for enhanced security
- No client secret (as it cannot be securely stored in browsers)
- Appropriate callback URLs for SPA authentication flows

## Why PKCE?

PKCE is the recommended approach for OAuth2 in public clients (like SPAs) because:

- It prevents authorization code interception attacks
- It doesn't require a client secret
- It's specifically designed for clients that cannot securely store credentials

## Prerequisites

- Terraform >= 1.0
- A running Pocket-ID instance
- An API token with admin privileges

## Usage

1. Copy `terraform.tfvars.example` to `terraform.tfvars`:

   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

2. Edit `terraform.tfvars` with your configuration:
   - `base_url`: Your Pocket-ID instance URL
   - `api_token`: Your admin API token
   - `app_name`: Name for your SPA
   - `app_urls`: URLs where your SPA is hosted

3. Initialize and apply:

   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

## What This Creates

- A public OIDC client with:
  - PKCE enabled for secure browser-based authentication
  - No client secret (public client)
  - Callback URLs for your SPA
  - Logout URLs for sign-out flow

## Integration Example

In your SPA, you would use this client like:

```javascript
// Example using oidc-client-js
const settings = {
  authority: 'https://pocket-id.example.com',
  client_id: '<output from terraform>',
  redirect_uri: 'http://localhost:3000/callback',
  response_type: 'code',
  scope: 'openid profile email',
  // PKCE is automatically handled by most OIDC libraries
};
```

## Clean Up

To remove all resources:

```bash
terraform destroy
```

## Security Notes

- This creates a public client (no client secret)
- PKCE is enabled by default for security
- Never try to include a client secret in your SPA code
- Always use HTTPS in production
