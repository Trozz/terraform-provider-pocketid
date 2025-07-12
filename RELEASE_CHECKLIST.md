# Release Checklist for Terraform Provider for Pocket-ID

This checklist ensures a smooth release process for new versions of the provider.

## Pre-Release Checklist

### Code Quality

- [ ] All unit tests pass locally (`make test`)
- [ ] All acceptance tests pass locally (`make test-acc`)
- [ ] No linting errors (`make lint`)
- [ ] Code coverage is acceptable (aim for >80%)
- [ ] All PR feedback has been addressed

### Documentation

- [ ] CHANGELOG.md is updated with all changes
- [ ] README.md is up to date
- [ ] Resource documentation is complete
- [ ] Data source documentation is complete
- [ ] Examples are working and up to date
- [ ] Run `tfplugindocs generate` to update generated docs
- [ ] Review generated documentation for accuracy

### Version Preparation

- [ ] Decide on version number following semantic versioning
  - [ ] MAJOR version for incompatible API changes
  - [ ] MINOR version for backwards-compatible functionality
  - [ ] PATCH version for backwards-compatible bug fixes
- [ ] Update version references if hardcoded anywhere
- [ ] Review and update compatibility matrix if needed

### Testing

- [ ] Test upgrade path from previous version
- [ ] Test with minimum supported Terraform version
- [ ] Test with latest Terraform version
- [ ] Test all resource CRUD operations
- [ ] Test all data sources
- [ ] Test import functionality
- [ ] Verify sensitive values are properly masked

## Release Process

### 1. Final Checks

- [ ] Ensure main branch is up to date
- [ ] No uncommitted changes (`git status`)
- [ ] All CI checks are passing on main branch

### 2. Create Release Tag

```bash
# For a new release (e.g., v0.1.0)
git tag -a v0.1.0 -m "Release v0.1.0"

# For a pre-release (e.g., v0.1.0-rc.1)
git tag -a v0.1.0-rc.1 -m "Pre-release v0.1.0-rc.1"

# Push the tag
git push origin v0.1.0
```

### 3. Monitor Release Workflow

- [ ] Check GitHub Actions release workflow is running
- [ ] Verify GPG signing is successful
- [ ] Confirm all platform binaries are built
- [ ] Check that release assets are properly uploaded

### 4. Verify GitHub Release

- [ ] Release appears on GitHub releases page
- [ ] Release notes are properly formatted
- [ ] All required assets are present:
  - [ ] Binary archives for all platforms
  - [ ] SHA256SUMS file
  - [ ] SHA256SUMS.sig file
  - [ ] Manifest JSON file
- [ ] Download and verify one binary works

### 5. Terraform Registry (First Time Only)

- [ ] Sign in to Terraform Registry
- [ ] Publish provider following PUBLISHING.md
- [ ] Verify webhook is created
- [ ] Confirm provider page is live

### 6. Verify Registry Release

- [ ] New version appears on Terraform Registry
- [ ] Documentation is properly rendered
- [ ] Installation instructions are correct
- [ ] Test installation with new version:

```hcl
terraform {
  required_providers {
    pocketid = {
      source  = "trozz/pocketid"
      version = "0.1.0"  # Use actual version
    }
  }
}
```

## Post-Release Checklist

### Communication

- [ ] Create GitHub discussion/announcement
- [ ] Update any pinned issues
- [ ] Notify Pocket-ID community (if applicable)
- [ ] Update project board/milestones

### Documentation Updates

- [ ] Update README.md badge from "pending" to active
- [ ] Update installation examples to use Registry source
- [ ] Archive any outdated documentation
- [ ] Update compatibility matrix

### Monitoring

- [ ] Monitor GitHub issues for problems
- [ ] Check Terraform Registry for any issues
- [ ] Watch for user feedback
- [ ] Track download statistics

### Housekeeping

- [ ] Close milestone for this release
- [ ] Create milestone for next release
- [ ] Update project board
- [ ] Plan next release features

## Rollback Plan

If issues are discovered after release:

1. **Do NOT delete or modify the release** - this breaks checksums
2. **Document the issue** in:
   - GitHub release notes (edit to add warning)
   - README.md if critical
   - Open a GitHub issue
3. **Release a patch version** with the fix
4. **Communicate** the issue and fix to users

## Emergency Contacts

- Terraform Registry Support: <terraform-registry@hashicorp.com>
- GitHub Support: <https://support.github.com>
- GPG Key Issues: Check PUBLISHING.md troubleshooting

## Version History

Track releases here for reference:

| Version | Date | Type | Notes |
|---------|------|------|-------|
| v0.1.0  | TBD  | Initial | First public release |

---

**Remember**: Once a version is released, it cannot be changed. Always release a new version for fixes.
