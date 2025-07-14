# PocketID Provider Tests

This directory contains comprehensive Terraform configurations to test all provider functionality.

## Test Structure

The tests are split into two separate configurations:

1. **terraform-resources/** - Creates all types of resources
2. **terraform-data-sources/** - Tests all data source lookups using the created resources

## Running the Tests

### Step 1: Create Resources

```bash
cd terraform-resources
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your PocketID instance details

terraform init
terraform apply

# Save outputs for the next test
terraform output -json > ../terraform-data-sources/resource-ids.json
```

### Step 2: Test Data Sources

```bash
cd ../terraform-data-sources
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with:
# - Your PocketID instance details
# - Resource IDs from Step 1

terraform init
terraform apply

# Verify outputs
terraform output
```

### Step 3: Clean Up

```bash
# Clean up data source test first (no resources to destroy)
cd terraform-data-sources
terraform destroy

# Then clean up created resources
cd ../terraform-resources
terraform destroy
```

## What's Tested

### Resources

- Users (enabled/disabled states)
- Groups
- User-Group associations
- One-time access tokens (with different settings)
- OAuth2 clients (various types: web, mobile, SPA, service account)

### Data Sources

- Individual resource lookups by ID
- List all resources of each type
- Filtered queries (e.g., enabled users)
- Resource counts and aggregations

## Notes

- All test resources are prefixed with "test-" to avoid conflicts
- Sensitive outputs (tokens, secrets) are properly marked
- The tests are designed to be run in sequence (resources first, then data sources)
