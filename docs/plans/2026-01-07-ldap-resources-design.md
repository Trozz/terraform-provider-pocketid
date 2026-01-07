# LDAP Resources Design

## Overview

Implementation of Terraform resources for managing LDAP configuration in Pocket ID.

**Issue:** [#3 - Add LDAP Configuration Resource](https://github.com/Trozz/terraform-provider-pocketid/issues/3)

**Branch:** `feature/3-LDAP-Config`

## Resources

| Resource | Purpose | API Endpoint |
|----------|---------|--------------|
| `pocketid_ldap_config` | Manage LDAP settings | `GET/PUT /api/application-configuration` |
| `pocketid_ldap_sync` | Trigger manual sync | `POST /api/application-configuration/sync-ldap` |

### Key Characteristics

- `ldap_config` is a **singleton** resource (one per PocketID instance, ID = "ldap")
- **Delete** behavior: Sets `enabled = false` (LDAP config always exists in PocketID)
- `ldap_config` includes `sync_on_change` option for automatic sync after updates
- `ldap_sync` uses **triggers pattern** for manual/scheduled sync control

## pocketid_ldap_config Schema

```hcl
resource "pocketid_ldap_config" "main" {
  # Core Settings
  enabled        = true   # Required
  sync_on_change = true   # Optional, default: false

  # Connection Settings
  url              = "ldaps://ldap.example.com:636"  # Required when enabled
  bind_dn          = "cn=admin,dc=example,dc=com"    # Required when enabled
  bind_password    = var.ldap_password               # Required when enabled, SENSITIVE
  base_dn          = "dc=example,dc=com"             # Required when enabled
  skip_cert_verify = false                           # Optional, default: false

  # Search Filters
  user_search_filter       = "(objectClass=person)"       # Optional
  user_group_search_filter = "(objectClass=groupOfNames)" # Optional

  # User Attribute Mappings (nested block)
  user_attributes {
    unique_identifier = "objectGUID"      # Required when enabled
    username          = "sAMAccountName"  # Required when enabled
    email             = "mail"            # Optional
    first_name        = "givenName"       # Optional
    last_name         = "sn"              # Optional
  }

  # Group Attribute Mappings (nested block)
  group_attributes {
    member            = "member"           # Optional, default: "member"
    unique_identifier = "objectGUID"       # Optional
    name              = "cn"               # Optional
    admin_group       = "PocketID-Admins"  # Optional
  }

  # Behavior
  soft_delete_users = true  # Optional, default: true
}
```

### Attribute Details

#### Core Settings

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `enabled` | Bool | Yes | - | Enable/disable LDAP integration |
| `sync_on_change` | Bool | No | false | Trigger sync after config updates |

#### Connection Settings

| Attribute | Type | Required | Sensitive | Description |
|-----------|------|----------|-----------|-------------|
| `url` | String | When enabled | No | LDAP server URL (ldap:// or ldaps://) |
| `bind_dn` | String | When enabled | No | DN for LDAP bind authentication |
| `bind_password` | String | When enabled | Yes | Password for bind DN |
| `base_dn` | String | When enabled | No | Base DN for LDAP searches |
| `skip_cert_verify` | Bool | No | No | Skip TLS certificate verification |

#### Search Filters

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `user_search_filter` | String | No | `(objectClass=person)` | Filter for finding users |
| `user_group_search_filter` | String | No | `(objectClass=groupOfNames)` | Filter for finding groups |

#### User Attributes (Nested Block)

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `unique_identifier` | String | When enabled | LDAP attribute for unique user ID |
| `username` | String | When enabled | LDAP attribute for username |
| `email` | String | No | LDAP attribute for email |
| `first_name` | String | No | LDAP attribute for first name |
| `last_name` | String | No | LDAP attribute for last name |

#### Group Attributes (Nested Block)

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `member` | String | No | `member` | LDAP attribute for group members |
| `unique_identifier` | String | No | - | LDAP attribute for unique group ID |
| `name` | String | No | - | LDAP attribute for group name |
| `admin_group` | String | No | - | Group name that grants admin role |

#### Behavior Settings

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `soft_delete_users` | Bool | No | true | Disable vs delete users not in LDAP |

### Validation Rules

- When `enabled = true`: `url`, `bind_dn`, `bind_password`, `base_dn`, `user_attributes.unique_identifier`, `user_attributes.username` are required
- `url` must be valid `ldap://` or `ldaps://` format
- `bind_dn` and `base_dn` must be valid DN format (contains `=`)

### CRUD Behavior

| Operation | Behavior |
|-----------|----------|
| Create | PUT config to API, optionally trigger sync |
| Read | GET config from API, map to state |
| Update | PUT config to API, optionally trigger sync |
| Delete | PUT config with `enabled = false` |

### Import

```bash
terraform import pocketid_ldap_config.main ldap
```

## pocketid_ldap_sync Schema

```hcl
resource "pocketid_ldap_sync" "manual" {
  triggers = {
    schedule  = timestamp()                    # Sync every apply
    config_id = pocketid_ldap_config.main.id   # Sync when config changes
    manual    = "2024-01-15"                   # Manual trigger
  }

  timeouts {
    create = "5m"
  }
}
```

### Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `triggers` | Map(String) | No | Values that trigger sync when changed |
| `id` | String | Computed | Resource identifier |
| `last_sync` | String | Computed | Timestamp of last successful sync |

### CRUD Behavior

| Operation | Behavior |
|-----------|----------|
| Create | POST sync-ldap, record timestamp |
| Read | Return stored state (no API call) |
| Update | If triggers changed, POST sync-ldap |
| Delete | No-op |

## Implementation Structure

### Files to Create/Modify

| File | Purpose |
|------|---------|
| `internal/client/models.go` | Add LDAP config structs |
| `internal/client/app_config.go` | API methods for app config |
| `internal/resources/ldap_config_resource.go` | LDAP config resource |
| `internal/resources/ldap_sync_resource.go` | LDAP sync resource |
| `internal/resources/validators.go` | LDAP URL and DN validators |
| `internal/provider/provider.go` | Register new resources |
| `examples/resources/pocketid_ldap_config/resource.tf` | Example config |
| `examples/resources/pocketid_ldap_sync/resource.tf` | Example sync |
| `templates/resources/pocketid_ldap_config.md.tmpl` | Doc template |
| `templates/resources/pocketid_ldap_sync.md.tmpl` | Doc template |

### Tests to Create

| File | Purpose |
|------|---------|
| `internal/resources/ldap_config_resource_test.go` | Unit tests for schema & CRUD |
| `internal/resources/ldap_sync_resource_test.go` | Unit tests for sync resource |
| `internal/resources/validators_test.go` | Validator unit tests |
| `internal/client/app_config_test.go` | API client tests |

## API Reference

### Endpoints

| Endpoint | Method | Auth | Purpose |
|----------|--------|------|---------|
| `/api/application-configuration` | GET | No | List public config |
| `/api/application-configuration/all` | GET | Yes | List all config |
| `/api/application-configuration` | PUT | Yes | Update config |
| `/api/application-configuration/sync-ldap` | POST | Yes | Trigger sync (returns 204) |

### Sources

- [PocketID LDAP Documentation](https://pocket-id.org/docs/configuration/ldap)
- [PocketID GitHub Repository](https://github.com/pocket-id/pocket-id)
- [app_config_controller.go](https://github.com/pocket-id/pocket-id/blob/main/backend/internal/controller/app_config_controller.go)
