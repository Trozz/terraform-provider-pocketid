# User and Group Management Example

This example demonstrates comprehensive user and group management in Pocket-ID using Terraform.

## Overview

This example shows how to:

- Create user groups for different roles
- Create users with various attributes
- Assign users to groups
- Manage admin privileges
- Use data sources to query existing users and groups

## Features Demonstrated

1. **Group Management**
   - Creating groups with friendly names
   - Organizing users by role or department

2. **User Management**
   - Creating users with full profiles
   - Setting admin privileges
   - Managing user status (enabled/disabled)
   - Assigning users to multiple groups

3. **Data Sources**
   - Querying all users and groups
   - Finding specific users by username
   - Retrieving group information

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
   - Update the Pocket-ID connection details
   - Customize the users and groups for your organization

3. Initialize and apply:

   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

## What This Creates

### Groups

- `developers` - Development team members
- `admins` - System administrators
- `support` - Support team members

### Users

- Admin user with full privileges
- Developer users assigned to the developers group
- Support user with limited access
- Disabled user account (for demonstration)

## Managing Users

### Adding a New User

Add a new user block to `main.tf`:

```hcl
resource "pocketid_user" "new_user" {
  username   = "newuser"
  email      = "newuser@example.com"
  first_name = "New"
  last_name  = "User"
  groups     = [pocketid_group.developers.id]
}
```

### Disabling a User

Set the `disabled` attribute to `true`:

```hcl
disabled = true
```

### Making a User Admin

Set the `is_admin` attribute to `true`:

```hcl
is_admin = true
```

## Clean Up

To remove all resources:

```bash
terraform destroy
```

## Important Notes

- Users created through Terraform won't have passwords or passkeys set
- Users will need to complete registration through the Pocket-ID UI
- Admin privileges should be granted sparingly
- Group memberships are managed through the user resource, not the group resource
