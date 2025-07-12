# Security Policy

## Supported Versions

We follow semantic versioning (semver) for our releases. Currently, we only support the most recent
version as we are in pre-1.0 development.

| Version | Supported          |
| ------- | ------------------ |
| 0.x.x (current) | :white_check_mark: |
| < current | :x:                |

Once we reach version 1.0.0, we will implement a more comprehensive support policy for previous versions.

## Reporting a Vulnerability

We take the security of terraform-provider-pocketid seriously. If you believe you have found a
security vulnerability, please report it to us as described below.

### Please do NOT

- Open a public GitHub issue
- Disclose the vulnerability publicly before a fix is available

### Please DO

- Email us at <security@leer.dev> with details of the vulnerability
- Include the steps to reproduce the issue
- Include the impact of the vulnerability
- Allow us reasonable time to respond and fix the issue before public disclosure

## Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 5 business days
- **Resolution Target**:
  - Critical (CVSS 9.0-10.0): 7 days
  - High (CVSS 7.0-8.9): 14 days
  - Medium (CVSS 4.0-6.9): 30 days
  - Low (CVSS 0.1-3.9): 90 days

## Security Best Practices

When using this Terraform provider:

### 1. Credential Management

- **Never** commit API tokens or credentials to version control
- Use environment variables or secure secret management solutions
- Rotate API tokens regularly
- Use the principle of least privilege for API tokens

### 2. State File Security

- Encrypt Terraform state files at rest
- Use remote state backends with proper access controls
- Never commit state files to version control
- Regularly audit who has access to state files

### 3. Provider Configuration

```hcl
# Good - Using environment variable
provider "pocketid" {
  # api_token sourced from POCKETID_API_TOKEN environment variable
}

# Bad - Hardcoded credentials
provider "pocketid" {
  api_token = "your-api-token-here" # NEVER DO THIS
}
```

### 4. Resource Security

- Regularly review and audit group memberships
- Implement proper RBAC (Role-Based Access Control)
- Use descriptive names for groups to avoid confusion
- Document the purpose of each group and its permissions

### 5. CI/CD Security

- Use GitHub Actions secrets for sensitive data
- Pin GitHub Actions to specific commit SHAs (as implemented)
- Enable branch protection rules
- Require code reviews for all changes
- Run security scans in CI pipeline

## Security Features

This provider implements several security features:

1. **Secure API Communication**: All API calls use HTTPS
2. **Input Validation**: Strict validation of all resource inputs
3. **Error Handling**: Sensitive information is not exposed in error messages
4. **Testing**: Comprehensive test coverage including security scenarios

## Vulnerability Disclosure Policy

When we receive a security bug report, we will:

1. Confirm the problem and determine affected versions
2. Audit code to find similar problems
3. Prepare fixes for all supported versions
4. Release new versions with the fix
5. Prominently announce the fix in release notes

## Security Updates

Security updates will be released as:

- Patch versions for non-breaking fixes
- Clear security advisory in GitHub
- Notification to users via GitHub watch notifications

## Contact

For security concerns, please email: <security@leer.dev>

For general bugs and feature requests, use [GitHub Issues](https://github.com/trozz/terraform-provider-pocketid/issues).

## Acknowledgments

We appreciate the security research community and will acknowledge reporters who:

- Follow responsible disclosure practices
- Allow us time to fix issues before public disclosure
- Provide clear reproduction steps

Thank you for helping keep terraform-provider-pocketid secure!
