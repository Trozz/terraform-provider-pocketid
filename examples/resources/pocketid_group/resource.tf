# Create a basic group
resource "pocketid_group" "developers" {
  name          = "developers"
  friendly_name = "Development Team"
}

# Create a group with custom claims included in member OIDC tokens.
# Setting custom_claims replaces all custom claims for the group.
# Reserved claim names (e.g. "email", "groups", "sub") are rejected by Pocket-ID.
resource "pocketid_group" "with_claims" {
  name          = "platform"
  friendly_name = "Platform Team"

  custom_claims = {
    role        = "admin"
    environment = "production"
  }
}
