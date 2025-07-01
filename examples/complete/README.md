# Complete Pocket-ID Provider Example

This example demonstrates a complete setup of the Pocket-ID Terraform provider, showcasing all available resources and data sources.

## What This Example Creates

### Groups
- **developers** - Development team with access to development resources
- **admins** - System administrators with full access
- **users** - Regular users with basic access

### OIDC Clients
- **SPA Application** - Public client for single-page applications
- **Web Application** - Confidential client with group restrictions
- **Mobile Application** - Public client for mobile apps
- **Admin Portal** - Restricted client for administrators only

### Users
- **Admin User** - System administrator
- **Developer Lead** - Developer with admin access
- **Developer** - Regular developer
- **Regular User** - Standard user
- **Test Users** - Multiple test users created using `for_each`

## Prerequisites

1. A running Pocket-ID instance
2. An API token with administrative privileges
3. Terraform 1.0 or later

## Usage

1. **Set up your variables**

   Create a `terraform.tfvars` file:
   ```hcl
   pocketid_base_url  = "https://auth.example.com"
   pocketid_api_token = "your-api-token-here"
   ```

   Or use environment variables:
   ```bash
   export TF_VAR_pocketid_base_url="https://auth.example.com"
   export TF_VAR_pocketid_api_token="your-api-token-here"
   ```

2. **Initialize Terraform**
   ```bash
   terraform init
   ```

3. **Review the plan**
   ```bash
   terraform plan
   ```

4. **Apply the configuration**
   ```bash
   terraform apply
   ```

5. **Save the client secrets**
   
   After applying, save the client secrets securely:
   ```bash
   terraform output -raw web_app_client_secret > web_app_secret.txt
   terraform output -raw admin_portal_client_secret > admin_portal_secret.txt
   ```

   ⚠️ **Important**: Client secrets are only available during resource creation!

## What Gets Created

### Resource Hierarchy

```
Groups
├── developers
├── admins
└── users

Users
├── admin (member of: admins)
├── john.doe (member of: developers, admins)
├── jane.smith (member of: developers)
├── bob.wilson (member of: users)
└── test[1-3] (members of: users)

OIDC Clients
├── React SPA Application (public, no restrictions)
├── Main Web Application (confidential, restricted to developers & admins)
├── Mobile Application (public, no restrictions)
└── Admin Portal (confidential, restricted to admins only)
```

### Access Matrix

| Client | Admin User | Dev Lead | Developer | Regular User | Test Users |
|--------|------------|----------|-----------|--------------|------------|
| SPA App | ✅ | ✅ | ✅ | ✅ | ✅ |
| Web App | ✅ | ✅ | ✅ | ❌ | ❌ |
| Mobile App | ✅ | ✅ | ✅ | ✅ | ✅ |
| Admin Portal | ✅ | ✅ | ❌ | ❌ | ❌ |

## Outputs

The configuration provides several outputs:

- `spa_client_id` - Client ID for the SPA
- `web_app_client_id` - Client ID for the web application
- `web_app_client_secret` - Client secret (sensitive)
- `admin_portal_client_id` - Client ID for admin portal
- `admin_portal_client_secret` - Client secret (sensitive)
- `total_clients` - Total number of OIDC clients
- `total_users` - Total number of users
- `developers_count` - Number of users in developers group

View all outputs:
```bash
terraform output
```

## Important Notes

### Passkey Registration

After creating users, they must:
1. Visit your Pocket-ID instance URL
2. Log in with their username
3. Register a passkey through the web interface

Users cannot authenticate until they complete passkey registration!

### Client Secrets

- Client secrets are only shown during resource creation
- They cannot be retrieved later through the API
- Store them securely immediately after creation
- If lost, you'll need to recreate the client or generate a new secret through the UI

### Group Restrictions

- Users must be members of allowed groups to authenticate to restricted clients
- The web app is restricted to developers and admins
- The admin portal is restricted to admins only
- Users can belong to multiple groups

## Customization

### Adding More Users

Add to the `test_users` local variable:
```hcl
locals {
  test_users = {
    "test4" = { email = "test4@example.com", first_name = "Test", last_name = "User4" }
    "test5" = { email = "test5@example.com", first_name = "Test", last_name = "User5" }
  }
}
```

### Creating Department Groups

Add new groups:
```hcl
resource "pocketid_group" "engineering" {
  name          = "engineering"
  friendly_name = "Engineering Department"
}

resource "pocketid_group" "sales" {
  name          = "sales"
  friendly_name = "Sales Department"
}
```

### Adding Environment-Specific Clients

Create clients for different environments:
```hcl
resource "pocketid_client" "staging_app" {
  name = "Staging Application"
  callback_urls = [
    "https://staging.example.com/callback"
  ]
  is_public = false
  allowed_user_groups = [
    pocketid_group.developers.id
  ]
}
```

## Testing the Setup

1. **Test SPA Authentication**
   ```
   https://auth.example.com/authorize?
     client_id=<spa_client_id>&
     redirect_uri=https://spa.example.com/callback&
     response_type=code&
     scope=openid profile email&
     code_challenge=<challenge>&
     code_challenge_method=S256
   ```

2. **Test Confidential Client**
   Use the client ID and secret to exchange authorization code for tokens

3. **Verify Group Restrictions**
   Try authenticating with different users to verify access controls

## Clean Up

To remove all created resources:
```bash
terraform destroy
```

⚠️ **Warning**: This will delete all users, groups, and OIDC clients created by this configuration!

## Troubleshooting

### "User cannot authenticate"
- Ensure the user has registered their passkey
- Check if the user is in the correct groups for restricted clients

### "Invalid redirect URI"
- Verify the callback URL exactly matches one in the client configuration
- Check for trailing slashes or protocol mismatches

### "Client not found"
- Ensure terraform apply completed successfully
- Check the client ID in the outputs

### Rate Limiting
- If you encounter rate limits, add delays between resource creation
- Consider creating resources in smaller batches

## Next Steps

1. Integrate your applications with the created OIDC clients
2. Set up monitoring for authentication events
3. Implement proper secret rotation procedures
4. Create additional groups and access policies as needed
5. Document your group structure and access policies