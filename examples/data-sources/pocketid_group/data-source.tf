# Look up a group by ID
data "pocketid_group" "by_id" {
  id = "grp_1234567890"
}

# Look up a group by name
data "pocketid_group" "developers" {
  name = "developers"
}

# Use the group data in other resources
resource "pocketid_user" "developer" {
  username   = "new.developer"
  email      = "new.developer@example.com"
  first_name = "New"
  last_name  = "Developer"

  groups = [data.pocketid_group.developers.id]
}

# Reference group information
output "developers_group_info" {
  value = {
    id            = data.pocketid_group.developers.id
    name          = data.pocketid_group.developers.name
    friendly_name = data.pocketid_group.developers.friendly_name
  }
}
