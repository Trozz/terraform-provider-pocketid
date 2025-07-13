# LDAP Configuration Examples

This directory contains examples of how to configure LDAP integration with Pocket-ID using Terraform.

## Prerequisites

1. A running Pocket-ID instance
2. Access to an LDAP/Active Directory server
3. Appropriate credentials for LDAP bind authentication

## Getting Started

1. Copy `terraform.tfvars.example` to `terraform.tfvars`:

   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

2. Update `terraform.tfvars` with your actual LDAP server details and credentials.

3. Set your Pocket-ID provider credentials:

   ```bash
   export POCKETID_BASE_URL="https://your-pocketid-instance.com"
   export POCKETID_API_TOKEN="your-api-token"
   ```

4. Initialize Terraform:

   ```bash
   terraform init
   ```

5. Review the planned changes:

   ```bash
   terraform plan
   ```

6. Apply the configuration:

   ```bash
   terraform apply
   ```

## Examples Included

### 1. Basic LDAP Configuration

A minimal LDAP setup with just the required fields.

### 2. Active Directory Configuration

A comprehensive Active Directory setup including:

- LDAPS with certificate verification
- Custom search filters for users and groups
- Full attribute mappings
- Admin group mapping

### 3. LDAP Connection Testing

Uses the `pocketid_ldap_test` data source to validate LDAP connectivity before enabling synchronization.

### 4. Conditional Configuration

Only enables LDAP if the connection test passes.

### 5. Automatic Synchronization

Triggers LDAP sync automatically when configuration changes.

### 6. Manual Sync Trigger

Provides a way to manually trigger LDAP synchronization by changing a trigger value.

## Important Notes

- The `bind_password` is marked as sensitive and won't be shown in Terraform output
- Setting `enabled = false` effectively disables LDAP without deleting the configuration
- The `skip_cert_verify` option should only be used in development environments
- Always test your LDAP configuration using the data source before enabling synchronization

## Troubleshooting

If LDAP connection fails:

1. Check the connection using the test data source
2. Verify network connectivity to the LDAP server
3. Ensure bind credentials are correct
4. Check that the base DN exists and is accessible
5. Verify search filters return expected results

## Security Considerations

- Store sensitive values like passwords in environment variables or a secure secret management system
- Use LDAPS (port 636) instead of plain LDAP (port 389) when possible
- Regularly rotate LDAP bind credentials
- Limit the permissions of the LDAP bind user to read-only access
