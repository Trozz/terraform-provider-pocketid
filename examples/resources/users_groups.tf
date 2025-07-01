# Example: Creating Groups
resource "pocketid_group" "administrators" {
  name        = "administrators"
  description = "System administrators with full access"
}

resource "pocketid_group" "developers" {
  name        = "developers"
  description = "Development team members"
}

resource "pocketid_group" "users" {
  name        = "users"
  description = "Regular users with standard access"
}

resource "pocketid_group" "api_consumers" {
  name        = "api-consumers"
  description = "Applications and services that consume APIs"
}

# Example: Creating Users
# Note: Users must complete passkey registration through the Pocket-ID UI
# This resource creates the user account, but authentication setup is done separately

resource "pocketid_user" "admin_user" {
  username   = "admin"
  email      = "admin@example.com"
  first_name = "Admin"
  last_name  = "User"

  # Assign to administrator group
  groups = [
    pocketid_group.administrators.id
  ]


}

resource "pocketid_user" "john_doe" {
  username   = "john.doe"
  email      = "john.doe@example.com"
  first_name = "John"
  last_name  = "Doe"

  # Assign to multiple groups
  groups = [
    pocketid_group.developers.id,
    pocketid_group.users.id
  ]


}

resource "pocketid_user" "jane_smith" {
  username   = "jane.smith"
  email      = "jane.smith@example.com"
  first_name = "Jane"
  last_name  = "Smith"

  groups = [
    pocketid_group.users.id
  ]


}

# Example: Disabled User Account
resource "pocketid_user" "inactive_user" {
  username   = "old.employee"
  email      = "old.employee@example.com"
  first_name = "Former"
  last_name  = "Employee"

  # User is disabled - cannot authenticate
  disabled = true
}

# Example: Service Account User
resource "pocketid_user" "service_account" {
  username = "ci-cd-service"
  email    = "ci-cd@example.com"

  # Service accounts might not need human names
  first_name = "CI/CD"
  last_name  = "Service"

  groups = [
    pocketid_group.api_consumers.id
  ]


}

# Example: Using Dynamic Groups
variable "departments" {
  description = "List of departments in the organization"
  type        = list(string)
  default     = ["engineering", "marketing", "sales", "support"]
}

resource "pocketid_group" "departments" {
  for_each = toset(var.departments)

  name        = each.value
  description = "${title(each.value)} department"
}

# Example: Creating Multiple Users from a List
variable "team_members" {
  description = "List of team members to create"
  type = list(object({
    username   = string
    email      = string
    first_name = string
    last_name  = string
    department = string
  }))
  default = [
    {
      username   = "alice.johnson"
      email      = "alice.johnson@example.com"
      first_name = "Alice"
      last_name  = "Johnson"
      department = "engineering"
    },
    {
      username   = "bob.wilson"
      email      = "bob.wilson@example.com"
      first_name = "Bob"
      last_name  = "Wilson"
      department = "marketing"
    }
  ]
}

resource "pocketid_user" "team_members" {
  for_each = { for member in var.team_members : member.username => member }

  username   = each.value.username
  email      = each.value.email
  first_name = each.value.first_name
  last_name  = each.value.last_name

  groups = [
    pocketid_group.users.id,
    pocketid_group.departments[each.value.department].id
  ]


}

# Example: Outputs for Integration
output "admin_group_id" {
  value       = pocketid_group.administrators.id
  description = "ID of the administrators group"
}

output "developer_users" {
  value = [
    for user in [pocketid_user.john_doe, pocketid_user.team_members["alice.johnson"]] :
    {
      username = user.username
      email    = user.email
    }
    if contains(user.groups, pocketid_group.developers.id)
  ]
  description = "List of users in the developers group"
}

# Example: Using Data Sources to Reference Existing Resources
# This would be in a separate configuration that imports existing users/groups

# data "pocketid_user" "existing_admin" {
#   username = "admin"
# }

# data "pocketid_group" "existing_admins" {
#   name = "administrators"
# }

# resource "pocketid_user" "new_admin" {
#   username    = "admin2"
#   email       = "admin2@example.com"
#   first_name  = "Second"
#   last_name = "Admin"
#
#   # Reference the existing group
#   groups = [
#     data.pocketid_group.existing_admins.id
#   ]
#
# }
# }
