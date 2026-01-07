terraform {
  required_version = ">= 1.0"
  required_providers {
    pocketid = {
      source = "trozz/pocketid"
    }
  }
}

provider "pocketid" {
  base_url  = var.base_url
  api_token = var.api_token
}

# Create groups for different roles
resource "pocketid_group" "developers" {
  name          = "developers"
  friendly_name = "Development Team"
}

resource "pocketid_group" "admins" {
  name          = "admins"
  friendly_name = "System Administrators"
}

resource "pocketid_group" "support" {
  name          = "support"
  friendly_name = "Support Team"
}

# Create an admin user
resource "pocketid_user" "admin" {
  username   = var.admin_username
  email      = var.admin_email
  first_name = "Admin"
  last_name  = "User"
  is_admin   = true
  groups     = [pocketid_group.admins.id]
}

# Create developer users
resource "pocketid_user" "developer1" {
  username   = "john.doe"
  email      = "john.doe@example.com"
  first_name = "John"
  last_name  = "Doe"
  locale     = "en"
  groups     = [pocketid_group.developers.id]
}

resource "pocketid_user" "developer2" {
  username   = "jane.smith"
  email      = "jane.smith@example.com"
  first_name = "Jane"
  last_name  = "Smith"
  locale     = "en"
  groups = [
    pocketid_group.developers.id,
    pocketid_group.admins.id # Jane is both developer and admin
  ]
  is_admin = true
}

# Create a support user
resource "pocketid_user" "support_user" {
  username   = "support.user"
  email      = "support@example.com"
  first_name = "Support"
  last_name  = "Team"
  groups     = [pocketid_group.support.id]
}

# Create a user with a custom display name
resource "pocketid_user" "custom_display_name" {
  username     = "jsmith"
  email        = "j.smith@example.com"
  first_name   = "John"
  last_name    = "Smith"
  display_name = "JS" # Custom display name instead of "John Smith"
  groups       = [pocketid_group.developers.id]
}

# Create a user where display_name is auto-generated from first and last names
resource "pocketid_user" "auto_display_name" {
  username   = "jdoe"
  email      = "jane.doe@example.com"
  first_name = "Jane"
  last_name  = "Doe"
  # display_name will automatically be set to "Jane Doe" by the API
  groups = [pocketid_group.support.id]
}

# Example of a disabled user
resource "pocketid_user" "disabled_user" {
  username   = "former.employee"
  email      = "former@example.com"
  first_name = "Former"
  last_name  = "Employee"
  disabled   = true
  groups     = [] # No group memberships
}

# Data sources to query existing resources
data "pocketid_users" "all" {
  depends_on = [
    pocketid_user.admin,
    pocketid_user.developer1,
    pocketid_user.developer2,
    pocketid_user.support_user,
    pocketid_user.disabled_user
  ]
}

data "pocketid_groups" "all" {
  depends_on = [
    pocketid_group.developers,
    pocketid_group.admins,
    pocketid_group.support
  ]
}

# Example of finding a specific user
data "pocketid_user" "jane" {
  username = pocketid_user.developer2.username

  depends_on = [pocketid_user.developer2]
}
