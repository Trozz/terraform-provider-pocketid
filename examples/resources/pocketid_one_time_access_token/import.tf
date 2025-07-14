# One-time access tokens can be imported using the user ID
# terraform import pocketid_one_time_access_token.example "user-id-123"

# Note: The token value itself cannot be retrieved after creation,
# so it will not be available in the state after import.

# Example import block (Terraform 1.5+):
import {
  to = pocketid_one_time_access_token.example
  id = "user-id-123"
}

# The resource block that the import will populate:
resource "pocketid_one_time_access_token" "example" {
  # These values will be populated from the import
  # except for the token which cannot be retrieved
  user_id       = "user-id-123"
  expires_at    = "2024-12-31T23:59:59Z"
  skip_recreate = true
}
