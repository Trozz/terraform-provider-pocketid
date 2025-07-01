# Build Attestations

This project uses GitHub's build attestations to provide cryptographic proof that our release artifacts were built by our GitHub Actions workflows.

## What are Build Attestations?

Build attestations are cryptographic signatures that prove:
- The artifacts were built by a specific GitHub Actions workflow
- The exact source code commit that was used
- The build hasn't been tampered with since creation

## Why are Attestations Important?

For infrastructure tools like Terraform providers:
- **Supply Chain Security**: Verify that the provider you're downloading was actually built by our CI/CD pipeline
- **Compliance**: Many organizations require provenance verification for infrastructure tools
- **Trust**: Users can cryptographically verify the authenticity of releases

## Verifying Attestations

### Using GitHub CLI

You can verify the attestations of our releases using the GitHub CLI:

```bash
# Install GitHub CLI if you haven't already
# See: https://cli.github.com/

# Verify a specific release artifact
gh attestation verify terraform-provider-pocketid_v1.0.0_darwin_amd64.zip \
  --owner Trozz \
  --repo terraform-provider-pocketid
```

### Using the Attestations API

You can also verify attestations programmatically using GitHub's API:

```bash
# Get attestations for a specific artifact
curl -H "Accept: application/vnd.github+json" \
  https://api.github.com/repos/Trozz/terraform-provider-pocketid/attestations/sha256:YOUR_ARTIFACT_SHA
```

### What's Attested?

We generate attestations for:

1. **Build Artifacts**: All compiled binaries in the build job
2. **Pre-release Artifacts**: Development builds from the main branch
3. **Release Artifacts**: Final signed release archives and checksums

## Attestation Details

Each attestation includes:

- **Subject**: The artifact(s) being attested
- **Predicate Type**: SLSA provenance v1.0
- **Build Information**:
  - Source repository and commit
  - GitHub Actions workflow details
  - Build timestamp
  - Builder identity

## Example Verification Output

When you verify an attestation, you'll see output like:

```
Loaded attestation from GitHub API
✓ Verification succeeded!

Repository: Trozz/terraform-provider-pocketid
Commit: abc123def456...
Workflow: .github/workflows/ci.yml
```

## Security Considerations

- Attestations are signed using GitHub's Sigstore infrastructure
- The signing keys are ephemeral and tied to the specific workflow run
- Attestations cannot be forged without access to GitHub's signing infrastructure
- Always verify attestations before using providers in production environments

## Further Reading

- [GitHub Artifact Attestations Documentation](https://docs.github.com/en/actions/security-guides/using-artifact-attestations-to-establish-provenance-for-builds)
- [SLSA Framework](https://slsa.dev/)
- [Sigstore Project](https://www.sigstore.dev/)