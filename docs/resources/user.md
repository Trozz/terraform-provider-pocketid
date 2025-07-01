---
page_title: "pocketid_user Resource - terraform-provider-pocketid"
subcategory: ""
description: |-
  Manages a user in Pocket-ID.
---

# pocketid_user (Resource)

Manages a user in Pocket-ID. This resource creates and manages user accounts, but note that passkey registration must be completed by the user through the Pocket-ID web interface.

~> **Important** Pocket-ID is a passkey-only authentication system. While this resource creates the user account, the user must visit the Pocket-ID web interface to register their passkey before they can authenticate.

## Example Usage

### Basic User

```hcl
resource "pocketid_user" "john_doe" {
  username   = "johndoe"
  email      = "john.doe@example.com"
  first_name = "John"
  last_name  = "Doe"
}
```

### Admin User

```hcl
resource "pocketid_user" "admin" {
  username   = "admin"
  email      = "admin@example.com"
  first_name = "Admin"
  last_name  = "User"
  is_admin   = true
}
```

### User with Groups

```hcl
resource "pocketid_group" "developers" {
  name          = "developers"
  friendly_name = "Development Team"
}

resource "pocketid_group" "managers" {
  name          = "managers"
  friendly_name = "Management Team"
}

resource "pocketid_user" "developer" {
  username   = "jane.developer"
  email      = "jane@example.com"
  first_name = "Jane"
  last_name  = "Developer"
  
  groups = [
    pocketid_group.developers.id,
    pocketid_group.managers.id
  ]
}
```

### Disabled User

```hcl
resource "pocketid_user" "inactive_user" {
  username   = "former.employee"
  email      = "former@example.com"
  first_name = "Former"
  last_name  = "Employee"
  disabled   = true  # User cannot authenticate
}
```

### User with Locale

```hcl
resource "pocketid_user" "german_user" {
  username   = "max.mustermann"
  email      = "max@example.de"
  first_name = "Max"
  last_name  = "Mustermann"
  locale     = "de-DE"
}
```

### Complete Example

```hcl
# Create groups first
resource "pocketid_group" "engineering" {
  name          = "engineering"
  friendly_name = "Engineering Team"
}

resource "pocketid_group" "senior_staff" {
  name          = "senior-staff"
  friendly_name = "Senior Staff"
}

# Create user with all options
resource "pocketid_user" "complete_example" {
  username   = "sarah.connor"
  email      = "sarah.connor@example.com"
  first_name = "Sarah"
  last_name  = "Connor"
  is_admin   = false
  locale     = "en-US"
  disabled   = false
  
  groups = [
    pocketid_group.engineering.id,
    pocketid_group.senior_staff.id
  ]
}

# Output user information
output "user_id" {
  value = pocketid_user.complete_example.id
}
```

## Argument Reference

The following arguments are supported:

* `username` - (Required) The username for the user. This must be unique within the Pocket-ID instance. The username is used for login and API operations.

* `email` - (Required) The email address of the user. This must be a valid email address and should be unique within the instance.

* `first_name` - (Optional) The user's first name. This is displayed in the Pocket-ID interface.

* `last_name` - (Optional) The user's last name. This is displayed in the Pocket-ID interface.

* `is_admin` - (Optional) Whether the user has administrative privileges. Defaults to `false`. Admin users can manage other users, groups, and OIDC clients.

* `locale` - (Optional) The user's preferred locale (e.g., "en-US", "de-DE"). This affects the language of the Pocket-ID interface for this user.

* `disabled` - (Optional) Whether the user account is disabled. Defaults to `false`. Disabled users cannot authenticate.

* `groups` - (Optional) List of group IDs that the user belongs to. Groups are used for access control with OIDC clients.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier of the user.

## Import

Users can be imported using their ID:

```bash
terraform import pocketid_user.example <user-id>
```

For example:

```bash
terraform import pocketid_user.john_doe 550e8400-e29b-41d4-a716-446655440000
```

## User Lifecycle Management

### Creating Users

When you create a user through Terraform:

1. The user account is created in Pocket-ID
2. The user receives an email (if email is configured in Pocket-ID)
3. The user must visit the Pocket-ID web interface to register their passkey
4. Only after passkey registration can the user authenticate

### Updating Users

You can update all user attributes except the username. Common update scenarios:

```hcl
# Promote user to admin
resource "pocketid_user" "user" {
  username = "johndoe"
  email    = "john@example.com"
  is_admin = true  # Changed from false
}

# Add user to groups
resource "pocketid_user" "user" {
  username = "johndoe"
  email    = "john@example.com"
  groups   = [pocketid_group.developers.id]  # Added group membership
}

# Disable user account
resource "pocketid_user" "user" {
  username = "johndoe"
  email    = "john@example.com"
  disabled = true  # Account disabled
}
```

### Deleting Users

When a user is deleted:
- All their passkeys are removed
- They are removed from all groups
- They can no longer authenticate
- Their data may be retained for audit purposes (check your Pocket-ID configuration)

## Group Management

### Adding Users to Groups

Users can be added to groups in two ways:

1. Through the user resource (shown in examples above)
2. Through the group resource's user list

Choose the approach that best fits your workflow:

```hcl
# Approach 1: Manage groups from user
resource "pocketid_user" "developer" {
  username = "developer"
  email    = "dev@example.com"
  groups   = [pocketid_group.engineering.id]
}

# Approach 2: Manage users from group (if supported)
# Check the pocketid_group resource documentation
```

### Group Membership Effects

Group membership affects:
- Which OIDC clients the user can authenticate to
- Access permissions within applications (if they check groups)
- Administrative capabilities (in conjunction with is_admin)

## Security Considerations

### Admin Users

Be careful when granting admin privileges:
- Admin users can manage all users and groups
- They can create and modify OIDC clients
- They can view audit logs
- Consider using groups for fine-grained permissions instead

### Disabled Accounts

Use the `disabled` flag for:
- Temporary suspension of access
- Preserving user data while preventing login
- Gradual offboarding processes

### Email Security

- Ensure email addresses are verified if your Pocket-ID instance sends emails
- Use organizational email addresses
- Update email addresses promptly when they change

## Common Patterns

### Onboarding New Employees

```hcl
locals {
  new_employees = {
    john = {
      username   = "john.smith"
      email      = "john.smith@company.com"
      first_name = "John"
      last_name  = "Smith"
      department = "engineering"
    }
    jane = {
      username   = "jane.doe"
      email      = "jane.doe@company.com"
      first_name = "Jane"
      last_name  = "Doe"
      department = "product"
    }
  }
}

resource "pocketid_user" "employees" {
  for_each = local.new_employees

  username   = each.value.username
  email      = each.value.email
  first_name = each.value.first_name
  last_name  = each.value.last_name
  
  groups = [
    pocketid_group.all_employees.id,
    pocketid_group.departments[each.value.department].id
  ]
}
```

### Bulk User Import

```hcl
# Load users from CSV or JSON
locals {
  users_csv = csvdecode(file("${path.module}/users.csv"))
}

resource "pocketid_user" "bulk_users" {
  for_each = { for u in local.users_csv : u.username => u }

  username   = each.value.username
  email      = each.value.email
  first_name = each.value.first_name
  last_name  = each.value.last_name
}
```

## Troubleshooting

### User Cannot Login
1. Ensure the user has registered their passkey
2. Check if the account is disabled
3. Verify group membership for restricted OIDC clients
4. Check Pocket-ID logs for authentication errors

### Email Not Received
- Verify Pocket-ID email configuration
- Check spam folders
- Ensure email address is correct
- Check Pocket-ID email logs

### Group Membership Not Working
- Verify group IDs are correct
- Check if groups are properly created
- Ensure user resource has been applied successfully
- Look for any group restrictions on OIDC clients

## Limitations

1. **Passkey Registration**: Cannot be automated through Terraform. Users must complete this step manually.

2. **Password Management**: Pocket-ID is passkey-only. There are no passwords to manage.

3. **Username Changes**: Usernames cannot be changed after creation. You must delete and recreate the user.

4. **Bulk Operations**: Large-scale user operations should be batched to avoid API rate limits.