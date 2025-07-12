---
name: Bug Report
about: Create a report to help us improve
title: '[BUG] '
labels: 'bug, needs-triage'
assignees: ''

---

## Bug Description

A clear and concise description of what the bug is.

## Terraform Version

Run `terraform -v` to show the version. If you are not running the latest version of Terraform, please upgrade because your issue may have already been fixed.

## Provider Version

If you are not running the latest version of the provider, please upgrade because your issue may have already been fixed.

## Affected Resource(s)

Please list the resources as a list, for example:

- pocketid_user
- pocketid_group

If this issue appears to affect multiple resources, it may be an issue with Terraform's core, so please mention this.

## Terraform Configuration Files

```hcl
# Copy-paste your Terraform configuration here.
# Please remove any sensitive information like API keys.
```

## Debug Output

Please provide a link to a GitHub Gist containing the complete debug output: <https://www.terraform.io/docs/internals/debugging.html>. Please do NOT paste the debug output in the issue; just paste a link to the Gist.

## Expected Behavior

What should have happened?

## Actual Behavior

What actually happened?

## Steps to Reproduce

Please list the steps required to reproduce the issue, for example:

1. `terraform apply`

## Important Factoids

Are there anything atypical about your accounts that we should know? For example: Running in a VPN environment, using a proxy, etc.

## References

Are there any other GitHub issues (open or closed) or Pull Requests that should be linked here? For example:

- #0000
