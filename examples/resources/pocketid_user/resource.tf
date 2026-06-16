# Create a basic user
resource "pocketid_user" "example" {
  username = "john.doe"
  email    = "john.doe@example.com"
}

# Create a user with custom claims included in OIDC tokens.
# Setting custom_claims replaces all custom claims for the user.
# Reserved claim names (e.g. "email", "groups", "sub") are rejected by Pocket-ID.
resource "pocketid_user" "with_claims" {
  username   = "claims.user"
  email      = "claims.user@example.com"
  first_name = "Claims"
  last_name  = "User"

  custom_claims = {
    department = "engineering"
    level      = "senior"
  }
}
