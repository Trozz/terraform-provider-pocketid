# One-time access tokens can be imported using the user ID
terraform import pocketid_one_time_access_token.example "user-id-123"

# Note: The token value itself cannot be retrieved after creation,
# so it will not be available in the state after import.
