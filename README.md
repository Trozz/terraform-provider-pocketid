# Terraform Provider for Pocket-ID

[![CI](https://github.com/Trozz/terraform-provider-pocketid/actions/workflows/ci.yml/badge.svg)](https://github.com/Trozz/terraform-provider-pocketid/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Trozz/terraform-provider-pocketid)](https://goreportcard.com/report/github.com/Trozz/terraform-provider-pocketid)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

The Terraform Provider for [Pocket-ID](https://github.com/pocket-id/pocket-id) enables you to manage OIDC clients, users, and groups in your Pocket-ID instance using Infrastructure as Code.

## What is Pocket-ID?

Pocket-ID is a simple, self-hosted OpenID Connect (OIDC) provider that uses passkeys for authentication instead of passwords. This makes it more secure and user-friendly than traditional authentication methods.

## Features

- 🔐 **OIDC Client Management**: Create and manage OAuth2/OIDC client applications
- 👥 **User Management**: Manage user accounts (passkey registration via UI)
- 👨‍👩‍👦‍👦 **Group Management**: Organize users and control access with groups
- 🔑 **Secure Authentication**: API token-based provider authentication
- 🚀 **Easy to Use**: Simple, intuitive resource definitions
- 📚 **Well Documented**: Comprehensive documentation and examples

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.20 (for development)
- A running [Pocket-ID](https://github.com/pocket-id/pocket-id) instance
- An API token from your Pocket-ID instance

## Installation

### Terraform Registry

```hcl
terraform {
  required_providers {
    pocketid = {
      source  = "trozz/pocketid"
      version = "~> 1.0"
    }
  }
}
```

### Manual Installation

1. Download the latest release from the [releases page](https://github.com/Trozz/terraform-provider-pocketid/releases)
2. Extract the archive
3. Move the binary to `~/.terraform.d/plugins/registry.terraform.io/trozz/pocketid/${VERSION}/${OS_ARCH}/`

## Quick Start

### 1. Configure the Provider

```hcl
# Using provider configuration
provider "pocketid" {
  base_url  = "https://auth.example.com"
  api_token = var.pocketid_api_token
}

# Or using environment variables
# export POCKETID_BASE_URL="https://auth.example.com"
# export POCKETID_API_TOKEN="your-api-token"
```

### 2. Create an OIDC Client

```hcl
resource "pocketid_client" "web_app" {
  name = "My Web Application"
  callback_urls = [
    "https://app.example.com/callback",
    "http://localhost:3000/callback"
  ]
  is_public    = false
  pkce_enabled = true
}

output "client_id" {
  value = pocketid_client.web_app.id
}

output "client_secret" {
  value     = pocketid_client.web_app.client_secret
  sensitive = true
}
```

### 3. Create Groups and Users

```hcl
# Create a group
resource "pocketid_group" "developers" {
  name          = "developers"
  friendly_name = "Development Team"
}

# Create a user
resource "pocketid_user" "john_doe" {
  username   = "johndoe"
  email      = "john@example.com"
  first_name = "John"
  last_name  = "Doe"
  groups     = [pocketid_group.developers.id]
}
```

## Resources

### Available Resources

- `pocketid_client` - Manages OIDC client applications
- `pocketid_user` - Manages user accounts
- `pocketid_group` - Manages user groups

### Available Data Sources

- `pocketid_client` - Queries a single OIDC client
- `pocketid_clients` - Lists all OIDC clients
- `pocketid_user` - Queries a single user
- `pocketid_users` - Lists users with optional filtering

## Documentation

Full documentation is available on the [Terraform Registry](https://registry.terraform.io/providers/trozz/pocketid/latest/docs).

### Quick Links

- [Provider Configuration](https://registry.terraform.io/providers/trozz/pocketid/latest/docs)
- [Resource: pocketid_client](https://registry.terraform.io/providers/trozz/pocketid/latest/docs/resources/client)
- [Resource: pocketid_user](https://registry.terraform.io/providers/trozz/pocketid/latest/docs/resources/user)
- [Resource: pocketid_group](https://registry.terraform.io/providers/trozz/pocketid/latest/docs/resources/group)

## Examples

See the [examples](examples/) directory for complete working examples:

- [Basic Provider Setup](examples/provider/)
- [Complete Example](examples/complete/) - Full setup with clients, users, and groups
- [Resource Examples](examples/resources/) - Individual resource examples

## Development

### Prerequisites

- Go 1.20+
- Terraform 1.0+
- A Pocket-ID instance for testing

### Building the Provider

```bash
# Clone the repository
git clone https://github.com/Trozz/terraform-provider-pocketid.git
cd terraform-provider-pocketid

# Install dependencies
make deps

# Build the provider
make build

# Install locally for testing
make install
```

### Running Tests

**Important Note**: Due to Pocket-ID's security model, acceptance tests cannot be run in CI/CD pipelines. Pocket-ID requires:
- Manual passkey registration through the web UI
- Manual API key generation through the admin interface
- No programmatic way to bootstrap an instance

#### Unit Tests (Automated)

```bash
# Run unit tests - these run in CI
make test
```

#### Acceptance Tests (Manual)

Acceptance tests require a manually configured Pocket-ID instance:

1. Start a Pocket-ID instance
2. Register a user with a passkey through the web UI
3. Generate an API key in the admin interface
4. Set environment variables:
   ```bash
   export POCKETID_BASE_URL="https://your-pocket-id-instance.com"
   export POCKETID_API_TOKEN="your-api-token"
   ```
5. Run acceptance tests:
   ```bash
   make test-acc
   ```

See [TESTING.md](TESTING.md) for detailed testing instructions and strategies.

### Local Development

1. Start a local Pocket-ID instance:
   ```bash
   make pocket-id-start
   ```

2. Build and install the provider:
   ```bash
   make dev
   ```

3. Use the provider in your Terraform configuration

### Debugging

Enable debug logging:
```bash
export TF_LOG=DEBUG
terraform apply
```

## Contributing

Contributions are welcome! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### How to Contribute

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Workflow

1. Write tests for your changes
2. Ensure all tests pass (`make test-all`)
3. Update documentation as needed
4. Follow the existing code style
5. Add yourself to the [CONTRIBUTORS](CONTRIBUTORS.md) file

## Roadmap

- [ ] Support for webhook resources
- [ ] Bulk user import functionality
- [ ] Enhanced policy management
- [ ] Session management features
- [ ] Automated passkey registration (when/if API supports it)

## Support

- **Issues**: [GitHub Issues](https://github.com/Trozz/terraform-provider-pocketid/issues)
- **Discussions**: [GitHub Discussions](https://github.com/Trozz/terraform-provider-pocketid/discussions)
- **Pocket-ID**: [Pocket-ID Repository](https://github.com/pocket-id/pocket-id)

## Security

### Reporting Security Issues

Please report security vulnerabilities to [security@example.com](mailto:security@example.com). Do not open public issues for security problems.

### Best Practices

1. **Never commit API tokens** to version control
2. Use environment variables or secure secret management
3. Enable TLS verification in production
4. Regularly rotate API tokens
5. Follow the principle of least privilege for API tokens

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- The [Pocket-ID](https://github.com/pocket-id/pocket-id) team for creating an awesome OIDC provider
- The [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) team
- All [contributors](CONTRIBUTORS.md) who have helped improve this provider

## Maintainers

- [@trozz](https://github.com/trozz)

---

Made with ❤️ by the Terraform Pocket-ID Provider community