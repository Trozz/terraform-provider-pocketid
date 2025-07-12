# Basic OIDC Client Example

This example demonstrates the minimal configuration required to create an OIDC client in Pocket-ID.

## Overview

This example creates:

- A basic OIDC client for a web application
- Callback URLs for authentication flow
- Logout callback URLs for sign-out flow

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
   - `client_name`: Name for your OIDC client
   - `callback_urls`: OAuth2 callback URLs for your application

3. Initialize and apply:

   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

## What This Creates

- An OIDC client with:
  - A unique client ID (generated)
  - A client secret (generated, marked as sensitive)
  - Specified callback URLs for OAuth2 flow
  - Logout callback URLs for sign-out

## Clean Up

To remove all resources:

```bash
terraform destroy
```

## Notes

- The client secret is marked as sensitive and won't be displayed in logs
- Store the client ID and secret securely for use in your application
- This creates a confidential client (not public) by default
