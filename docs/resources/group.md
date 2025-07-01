---
page_title: "pocketid_group Resource - terraform-provider-pocketid"
subcategory: ""
description: |-
  Manages a user group in Pocket-ID.
---

# pocketid_group (Resource)

Manages a user group in Pocket-ID. Groups are used to organize users and control access to OIDC clients. Users can belong to multiple groups, and OIDC clients can be configured to only allow authentication from users in specific groups.

## Example Usage

### Basic Group

```hcl
resource "pocketid_group" "developers" {
  name          = "developers"
  friendly_name = "Development Team"
}
```

### Multiple Groups

```hcl
resource "pocketid_group" "engineering" {
  name          = "engineering"
  friendly_name = "Engineering Department"
}

resource "pocketid_group" "senior_engineers" {
  name          = "senior-engineers"
  friendly_name = "Senior Engineering Team"
}

resource "pocketid_group" "architects" {
  name          = "architects"
  friendly_name = "Software Architects"
}
```

### Groups with Users

```hcl
# Create groups
resource "pocketid_group" "frontend_team" {
  name          = "frontend"
  friendly_name = "Frontend Development Team"
}

resource "pocketid_group" "backend_team" {
  name          = "backend"
  friendly_name = "Backend Development Team"
}

# Create users and assign to groups
resource "pocketid_user" "frontend_dev" {
  username   = "alice.frontend"
  email      = "alice@example.com"
  first_name = "Alice"
  last_name  = "Developer"
  
  groups = [pocketid_group.frontend_team.id]
}

resource "pocketid_user" "fullstack_dev" {
  username   = "bob.fullstack"
  email      = "bob@example.com"
  first_name = "Bob"
  last_name  = "Engineer"
  
  # User belongs to both groups
  groups = [
    pocketid_group.frontend_team.id,
    pocketid_group.backend_team.id
  ]
}
```

### Groups for Access Control

```hcl
# Create groups for different access levels
resource "pocketid_group" "admins" {
  name          = "admins"
  friendly_name = "System Administrators"
}

resource "pocketid_group" "regular_users" {
  name          = "users"
  friendly_name = "Regular Users"
}

resource "pocketid_group" "readonly_users" {
  name          = "readonly"
  friendly_name = "Read-Only Users"
}

# Create OIDC clients with group restrictions
resource "pocketid_client" "admin_portal" {
  name          = "Admin Portal"
  callback_urls = ["https://admin.example.com/callback"]
  
  # Only admins can access this client
  allowed_user_groups = [pocketid_group.admins.id]
}

resource "pocketid_client" "user_app" {
  name          = "User Application"
  callback_urls = ["https://app.example.com/callback"]
  
  # Regular users and admins can access
  allowed_user_groups = [
    pocketid_group.regular_users.id,
    pocketid_group.admins.id
  ]
}
```

### Department-Based Groups

```hcl
locals {
  departments = {
    engineering = "Engineering Department"
    sales       = "Sales Department"
    marketing   = "Marketing Department"
    hr          = "Human Resources"
    finance     = "Finance Department"
  }
}

resource "pocketid_group" "departments" {
  for_each = local.departments

  name          = each.key
  friendly_name = each.value
}

# Reference specific department groups
resource "pocketid_user" "engineer" {
  username = "john.engineer"
  email    = "john@example.com"
  groups   = [pocketid_group.departments["engineering"].id]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The unique name of the group. This is used as an identifier and should be URL-safe (lowercase letters, numbers, and hyphens). Cannot be changed after creation.

* `friendly_name` - (Required) The human-readable display name of the group. This is shown in the Pocket-ID user interface.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier of the group. This is used when assigning users to groups or restricting OIDC client access.

## Import

Groups can be imported using their ID:

```bash
terraform import pocketid_group.example <group-id>
```

For example:

```bash
terraform import pocketid_group.developers 550e8400-e29b-41d4-a716-446655440000
```

## Group Management Patterns

### Hierarchical Groups

While Pocket-ID doesn't support nested groups directly, you can create a hierarchical structure through naming conventions:

```hcl
resource "pocketid_group" "eng_all" {
  name          = "eng-all"
  friendly_name = "All Engineering"
}

resource "pocketid_group" "eng_frontend" {
  name          = "eng-frontend"
  friendly_name = "Engineering - Frontend"
}

resource "pocketid_group" "eng_backend" {
  name          = "eng-backend"
  friendly_name = "Engineering - Backend"
}

resource "pocketid_group" "eng_devops" {
  name          = "eng-devops"
  friendly_name = "Engineering - DevOps"
}

# Users can belong to multiple groups to simulate hierarchy
resource "pocketid_user" "frontend_engineer" {
  username = "frontend.engineer"
  email    = "frontend@example.com"
  groups = [
    pocketid_group.eng_all.id,      # Parent group
    pocketid_group.eng_frontend.id   # Specific team
  ]
}
```

### Role-Based Access Control (RBAC)

Implement RBAC patterns using groups:

```hcl
# Define roles as groups
resource "pocketid_group" "role_viewer" {
  name          = "role-viewer"
  friendly_name = "Viewer Role"
}

resource "pocketid_group" "role_editor" {
  name          = "role-editor"
  friendly_name = "Editor Role"
}

resource "pocketid_group" "role_admin" {
  name          = "role-admin"
  friendly_name = "Admin Role"
}

# Create application clients with role-based access
resource "pocketid_client" "app" {
  name          = "Main Application"
  callback_urls = ["https://app.example.com/callback"]
  
  # All roles can access the app
  allowed_user_groups = [
    pocketid_group.role_viewer.id,
    pocketid_group.role_editor.id,
    pocketid_group.role_admin.id
  ]
}

# The application can check group membership to determine permissions
```

### Project-Based Groups

Organize users by projects:

```hcl
locals {
  projects = ["alpha", "beta", "gamma"]
}

resource "pocketid_group" "projects" {
  for_each = toset(local.projects)

  name          = "project-${each.value}"
  friendly_name = "Project ${title(each.value)}"
}

# Create project-specific OIDC clients
resource "pocketid_client" "project_apps" {
  for_each = pocketid_group.projects

  name          = "${each.value.friendly_name} App"
  callback_urls = ["https://${each.key}.example.com/callback"]
  
  # Only project members can access their app
  allowed_user_groups = [each.value.id]
}
```

## Best Practices

### Naming Conventions

1. **Use lowercase with hyphens** for the `name` field:
   - Good: `frontend-team`, `senior-developers`, `project-alpha`
   - Avoid: `Frontend Team`, `Senior_Developers`, `ProjectAlpha`

2. **Be descriptive** in `friendly_name`:
   - Good: "Senior Frontend Development Team"
   - Avoid: "Team 1", "Group A"

3. **Use prefixes** for organization:
   - Departments: `dept-engineering`, `dept-sales`
   - Roles: `role-admin`, `role-viewer`
   - Projects: `project-alpha`, `project-beta`
   - Locations: `loc-us-east`, `loc-eu-west`

### Group Strategy

1. **Start Simple**: Begin with basic groups and expand as needed
2. **Avoid Over-Grouping**: Too many groups can become hard to manage
3. **Document Purpose**: Use friendly names that clearly indicate the group's purpose
4. **Regular Audits**: Periodically review and clean up unused groups

### Security Considerations

1. **Principle of Least Privilege**: Create groups that grant minimum necessary access
2. **Separation of Duties**: Use different groups for different responsibilities
3. **Regular Reviews**: Audit group memberships periodically
4. **Clear Ownership**: Document who is responsible for managing each group

## Common Use Cases

### Multi-Tenant Applications

```hcl
# Create tenant groups
resource "pocketid_group" "tenants" {
  for_each = {
    acme    = "ACME Corporation"
    globex  = "Globex Industries"
    initech = "Initech Systems"
  }

  name          = "tenant-${each.key}"
  friendly_name = "Tenant: ${each.value}"
}

# Create tenant-specific clients
resource "pocketid_client" "tenant_apps" {
  for_each = pocketid_group.tenants

  name          = "${each.value.friendly_name} Portal"
  callback_urls = ["https://${each.key}.app.example.com/callback"]
  
  allowed_user_groups = [each.value.id]
}
```

### Environment-Based Access

```hcl
resource "pocketid_group" "env_dev" {
  name          = "env-development"
  friendly_name = "Development Environment Access"
}

resource "pocketid_group" "env_staging" {
  name          = "env-staging"
  friendly_name = "Staging Environment Access"
}

resource "pocketid_group" "env_prod" {
  name          = "env-production"
  friendly_name = "Production Environment Access"
}
```

## Troubleshooting

### Group Not Appearing in UI
- Ensure the group resource has been successfully applied
- Check for any API errors in Terraform output
- Verify the Pocket-ID instance is accessible

### Users Can't Access Restricted Clients
- Verify the user is a member of at least one allowed group
- Check the OIDC client's `allowed_user_groups` configuration
- Ensure group IDs are correctly referenced
- Look for typos in group assignments

### Import Issues
- Ensure you're using the correct group ID (not the name)
- Verify the group exists in Pocket-ID
- Check API permissions for reading groups

## Limitations

1. **No Nested Groups**: Pocket-ID doesn't support hierarchical group structures. Use naming conventions and multiple group assignments instead.

2. **No Dynamic Membership**: Group membership must be explicitly managed. There's no support for rule-based automatic assignment.

3. **Name Immutability**: Group names cannot be changed after creation. Plan your naming convention carefully.

4. **No Group Attributes**: Groups only have name and friendly_name. Additional metadata must be managed externally.