# Build Matrix for Terraform Registry

This document outlines the OS and architecture combinations that are built for the Pocket-ID Terraform provider.

## Terraform Registry Requirements

According to [Terraform Registry documentation](https://developer.hashicorp.com/terraform/registry/providers/os-arch), providers should build for the following platforms:

### Required Platforms ✅

- ✅ **Darwin (macOS) / AMD64** - Intel Macs
- ✅ **Darwin (macOS) / ARM64** - Apple Silicon Macs (M1/M2/M3)
- ✅ **Linux / AMD64** - Standard Linux x86_64 (required for HCP Terraform)
- ✅ **Linux / ARM64** - Linux on ARM64/AArch64
- ✅ **Linux / ARMv6** - Linux on ARM v6 (Raspberry Pi, etc.)
- ✅ **Windows / AMD64** - Windows 64-bit

### Recommended Platforms ✅

- ✅ **Linux / 386** - Linux 32-bit
- ✅ **Windows / 386** - Windows 32-bit
- ✅ **FreeBSD / 386** - FreeBSD 32-bit
- ✅ **FreeBSD / AMD64** - FreeBSD 64-bit

## Current Build Configuration

Based on `.goreleaser.yml`, the following platforms are built:

### Darwin (macOS)

- `darwin/amd64` - Intel Macs
- `darwin/arm64` - Apple Silicon Macs

### Linux

- `linux/386` - 32-bit
- `linux/amd64` - 64-bit (CGO disabled for HCP Terraform compatibility)
- `linux/arm64` - ARM64/AArch64
- `linux/arm/6` - ARMv6 (Raspberry Pi Zero, Pi 1)
- `linux/arm/7` - ARMv7 (Raspberry Pi 2+)

### Windows

- `windows/386` - 32-bit
- `windows/amd64` - 64-bit

### FreeBSD

- `freebsd/386` - 32-bit
- `freebsd/amd64` - 64-bit

## Build Settings

All binaries are built with:

- **CGO_ENABLED=0** - Static binaries with no external dependencies
- **-trimpath** flag - Reproducible builds
- **Proper versioning** - Version, commit, and date embedded in binaries

This configuration ensures compatibility with Terraform Registry requirements and HCP Terraform.
