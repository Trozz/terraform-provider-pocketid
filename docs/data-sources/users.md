---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "pocketid_users Data Source - pocketid"
subcategory: ""
description: |-
  Fetches all users from Pocket-ID.
---

# pocketid_users (Data Source)

Fetches all users from Pocket-ID.



<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `users` (Attributes List) List of all users. (see [below for nested schema](#nestedatt--users))

<a id="nestedatt--users"></a>
### Nested Schema for `users`

Read-Only:

- `disabled` (Boolean) Whether the user account is disabled.
- `email` (String) The email address of the user.
- `first_name` (String) The first name of the user.
- `groups` (Set of String) List of group IDs the user belongs to.
- `id` (String) The ID of the user.
- `is_admin` (Boolean) Whether the user has administrator privileges.
- `last_name` (String) The last name of the user.
- `locale` (String) The locale preference for the user.
- `username` (String) The username of the user.
