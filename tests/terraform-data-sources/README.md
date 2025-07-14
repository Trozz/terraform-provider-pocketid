# Data Source Tests

This test verifies all data source functionality by looking up resources created in the resource test.

## Prerequisites

1. Run the resource creation test first:

   ```bash
   cd ../terraform-resources
   terraform apply
   terraform output -json > ../terraform-data-sources/outputs.json
   ```

2. Extract the resource IDs from the output and add them to your `terraform.tfvars`.

## Test Coverage

### Individual Lookups

- User lookup by ID
- Group lookup by ID
- Client lookup by ID

### List Operations

- List all users
- List all groups
- List all clients

### Filtered Queries

- Filter users by enabled status
- Count users by status

## Usage

1. Configure variables:

   ```bash
   cp terraform.tfvars.example terraform.tfvars
   # Edit terraform.tfvars with:
   # - Your PocketID instance details
   # - Resource IDs from the resource creation test
   ```

2. Run the test:

   ```bash
   terraform init
   terraform apply
   ```

3. Review outputs to verify data sources are working correctly:

   ```bash
   terraform output
   ```

## Expected Results

The outputs should show:

- Individual resource details matching what was created
- Correct counts for total resources
- Proper filtering (enabled vs disabled users)
- All resource attributes properly populated
