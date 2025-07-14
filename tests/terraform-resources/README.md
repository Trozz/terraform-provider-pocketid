# Resource Creation Tests

This test creates all types of resources supported by the PocketID provider.

## Resources Created

- **Groups**: 3 groups (administrators, developers, users)
- **Users**: 3 users with different states and group memberships
- **One-time Access Tokens**: 2 tokens with different expiration and recreation settings
- **OAuth2 Clients**: 4 clients representing different application types

## Usage

1. Copy and configure variables:

   ```bash
   cp terraform.tfvars.example terraform.tfvars
   # Edit terraform.tfvars with your PocketID instance details
   ```

2. Initialize and apply:

   ```bash
   terraform init
   terraform apply
   ```

3. Save the outputs for data source testing:

   ```bash
   terraform output -json > ../terraform-data-sources/resource-ids.json
   ```

## Outputs

The configuration outputs all created resource IDs, which can be used by the data source tests to verify lookups work correctly.
