---
name: Provider Crash Report
about: Report a provider crash or panic
title: '[CRASH] '
labels: 'bug, crash, priority-high'
assignees: ''

---

## Crash Summary

Brief description of what you were doing when the provider crashed.

## Terraform Version

Run `terraform -v` to show the version.

## Provider Version

Version of terraform-provider-pocketid that crashed.

## Terraform Configuration

```hcl
# Minimal configuration that reproduces the crash
# Please remove any sensitive information
```

## Crash Output

```
# Please paste the full panic output here
# This typically starts with "panic:" and includes a stack trace
```

## Steps to Reproduce

1.
2.
3.

## Debug Logs

Please provide a link to a GitHub Gist containing the complete debug output leading up to the crash: <https://www.terraform.io/docs/internals/debugging.html>

## Environment Details

- Operating System:
- Architecture (x86_64, arm64, etc.):
- Any special network configuration (proxy, VPN, etc.):

## Workaround

Have you found any way to avoid the crash? If so, please describe.

## Note

Provider crashes are high-priority issues. We'll investigate as soon as possible.
