# Publishing the Terraform Provider for Pocket-ID

This guide walks through the process of publishing the Terraform Provider for Pocket-ID to the Terraform Registry.

## Prerequisites

- A GitHub account with admin access to the provider repository
- GPG installed on your system
- The provider repository must be public
- At least one release version ready to publish

## Step 1: Generate a GPG Signing Key

All provider releases must be signed with a GPG key. The Terraform Registry will validate signatures when publishing, and Terraform will verify them during `terraform init`.

### Generate the Key

```bash
# Generate a new GPG key (RSA, 4096 bits recommended)
gpg --full-generate-key
```

When prompted:
1. Select `(1) RSA and RSA`
2. Key size: `4096`
3. Key validity: `0` (does not expire) or your preferred expiration
4. Enter your real name
5. Enter your email address
6. Add a comment (optional, e.g., "Terraform Provider Signing")
7. Enter a secure passphrase

### List Your Keys

```bash
# List your GPG keys to find the key ID
gpg --list-secret-keys --keyid-format=long

# Example output:
# sec   rsa4096/ABCD1234EFGH5678 2024-01-20 [SC]
#       Key fingerprint = 1234 5678 90AB CDEF 1234  5678 90AB CDEF 1234 5678
# uid                 [ultimate] Your Name <your.email@example.com>
```

The key ID is the part after `rsa4096/` (in this example: `ABCD1234EFGH5678`).

### Export Your Public Key

```bash
# Export your public key in ASCII-armored format
gpg --armor --export ABCD1234EFGH5678 > terraform-provider-signing.asc

# Or export by email
gpg --armor --export your.email@example.com > terraform-provider-signing.asc
```

### Export Your Private Key (for GitHub Actions)

```bash
# Export your private key for GitHub Actions
gpg --armor --export-secret-keys ABCD1234EFGH5678 > private-key.asc

# IMPORTANT: Keep this file secure and delete it after adding to GitHub Secrets!
```

## Step 2: Add Your Public Key to Terraform Registry

1. Sign in to the [Terraform Registry](https://registry.terraform.io) with your GitHub account
2. Navigate to **User Settings** → **Signing Keys**
3. Click **Add a New Signing Key**
4. Paste the contents of `terraform-provider-signing.asc`
5. Give it a descriptive name (e.g., "Pocket-ID Provider Signing Key")
6. Click **Add Key**

## Step 3: Configure GitHub Repository Secrets

Add the following secrets to your GitHub repository for automated releases:

1. Go to your repository on GitHub
2. Navigate to **Settings** → **Secrets and variables** → **Actions**
3. Add these repository secrets:

### `GPG_PRIVATE_KEY`
- Click **New repository secret**
- Name: `GPG_PRIVATE_KEY`
- Value: Contents of `private-key.asc` (the entire ASCII-armored private key)

### `PASSPHRASE`
- Click **New repository secret**
- Name: `PASSPHRASE`
- Value: The passphrase you used when generating the GPG key

### Clean Up
```bash
# IMPORTANT: Delete the private key file after adding to GitHub!
rm private-key.asc
```

## Step 4: Create GitHub Actions Release Workflow

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: 'go.mod'
          cache: true

      - name: Import GPG key
        id: import_gpg
        uses: crazy-max/ghaction-import-gpg@v6
        with:
          gpg_private_key: ${{ secrets.GPG_PRIVATE_KEY }}
          passphrase: ${{ secrets.PASSPHRASE }}

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GPG_FINGERPRINT: ${{ steps.import_gpg.outputs.fingerprint }}
```

## Step 5: Test the Release Process

Before publishing to the Registry, test your release process:

```bash
# Create a test tag
git tag v0.1.0-alpha.1
git push origin v0.1.0-alpha.1

# Monitor the GitHub Actions workflow
```

Check that the release includes:
- Binary archives for all platforms
- `terraform-provider-pocketid_0.1.0-alpha.1_manifest.json`
- `terraform-provider-pocketid_0.1.0-alpha.1_SHA256SUMS`
- `terraform-provider-pocketid_0.1.0-alpha.1_SHA256SUMS.sig`

## Step 6: Publish to Terraform Registry

1. Sign in to the [Terraform Registry](https://registry.terraform.io)
2. Click **Publish** → **Provider** in the top navigation
3. Select your GitHub organization or username
4. Select the `terraform-provider-pocketid` repository
5. Review and accept the Terms of Use
6. Click **Publish Provider**

The Registry will:
- Create a webhook on your repository for future releases
- Ingest your existing releases
- Make your provider available at `registry.terraform.io/providers/[namespace]/pocketid`

## Step 7: Create Your First Official Release

Once published, create your first official release:

```bash
# Tag your release
git tag v0.1.0
git push origin v0.1.0

# The GitHub Actions workflow will automatically:
# 1. Build binaries for all platforms
# 2. Create checksums
# 3. Sign the release with your GPG key
# 4. Upload to GitHub Releases
# 5. Notify the Terraform Registry
```

## Step 8: Verify the Publication

1. Visit `https://registry.terraform.io/providers/[your-namespace]/pocketid`
2. Check that your provider appears with:
   - Correct documentation
   - Available versions
   - Installation instructions

3. Test installation:
```hcl
terraform {
  required_providers {
    pocketid = {
      source  = "[your-namespace]/pocketid"
      version = "~> 0.1.0"
    }
  }
}
```

## Troubleshooting

### Webhook Issues
If releases aren't appearing in the Registry:
1. Go to the provider page on Terraform Registry
2. Click **Settings** → **Resync**
3. Check your repository's webhook settings

### GPG Key Issues
- Ensure you're using RSA or DSA keys (not ECC)
- Verify the key hasn't expired
- Check that the fingerprint matches between GitHub Actions and your local key

### Release Asset Issues
Ensure your release includes all required files:
- Binary archives matching the pattern: `terraform-provider-pocketid_VERSION_OS_ARCH.zip`
- Manifest file: `terraform-provider-pocketid_VERSION_manifest.json`
- Checksums: `terraform-provider-pocketid_VERSION_SHA256SUMS`
- Signature: `terraform-provider-pocketid_VERSION_SHA256SUMS.sig`

## Best Practices

1. **Semantic Versioning**: Follow [semver](https://semver.org/) for version tags
   - `v0.x.y` for initial development
   - `v1.0.0` for first stable release
   - Include pre-release tags for testing: `v1.0.0-rc.1`

2. **Documentation**: Keep your docs up to date
   - Run `tfplugindocs generate` before releases
   - Review rendered docs on the Registry

3. **Testing**: Always test releases with alpha/beta tags first

4. **Security**: 
   - Never commit your GPG private key
   - Rotate keys periodically
   - Use a strong passphrase

## Support

If you encounter issues:
- Provider development: Open an issue in this repository
- Registry issues: Contact terraform-registry@hashicorp.com
- Documentation: See [Publishing Providers](https://developer.hashicorp.com/terraform/registry/providers/publishing)

## Next Steps

After publishing:
1. Add the Registry badge to your README
2. Update installation instructions to use the Registry source
3. Announce the availability to the Pocket-ID community
4. Monitor issues and feedback from users