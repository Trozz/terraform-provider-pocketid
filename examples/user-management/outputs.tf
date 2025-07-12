# Group outputs
output "group_ids" {
  description = "Map of group names to their IDs"
  value = {
    developers = pocketid_group.developers.id
    admins     = pocketid_group.admins.id
    support    = pocketid_group.support.id
  }
}

# User outputs
output "admin_user" {
  description = "Admin user details"
  value = {
    id       = pocketid_user.admin.id
    username = pocketid_user.admin.username
    email    = pocketid_user.admin.email
  }
}

output "all_users" {
  description = "List of all users"
  value = [
    for user in data.pocketid_users.all.users : {
      username = user.username
      email    = user.email
      is_admin = user.is_admin
      disabled = user.disabled
    }
  ]
}

output "all_groups" {
  description = "List of all groups"
  value = [
    for group in data.pocketid_groups.all.groups : {
      name          = group.name
      friendly_name = group.friendly_name
    }
  ]
}

output "user_group_assignments" {
  description = "Summary of user-to-group assignments"
  value = {
    "${pocketid_user.admin.username}"         = [pocketid_group.admins.friendly_name]
    "${pocketid_user.developer1.username}"    = [pocketid_group.developers.friendly_name]
    "${pocketid_user.developer2.username}"    = [pocketid_group.developers.friendly_name, pocketid_group.admins.friendly_name]
    "${pocketid_user.support_user.username}"  = [pocketid_group.support.friendly_name]
    "${pocketid_user.disabled_user.username}" = []
  }
}

output "active_users_count" {
  description = "Number of active (non-disabled) users"
  value       = length([for user in data.pocketid_users.all.users : user if !user.disabled])
}

output "admin_users_count" {
  description = "Number of admin users"
  value       = length([for user in data.pocketid_users.all.users : user if user.is_admin])
}
