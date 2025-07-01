# Testing Guide for Terraform Provider for Pocket-ID

## Overview

This document explains how to test the Terraform Provider for Pocket-ID, including the limitations due to Pocket-ID's security model and recommended testing strategies.

## Testing Limitations

### Why Acceptance Tests Can't Run in CI

Pocket-ID's security-first approach means that:

1. **No Default Credentials**: There are no default admin users or API keys
2. **Passkey-Only Authentication**: Users must register a passkey through the web UI
3. **Manual API Key Generation**: API keys must be generated manually through the admin interface
4. **No Bootstrap API**: There's no programmatic way to set up an initial admin user

These security features, while excellent for production use, mean that we cannot fully automate acceptance testing in CI/CD pipelines.

## Testing Strategy

### 1. Unit Tests (Automated in CI)

Unit tests run automatically in CI and cover:
- Provider configuration validation
- Resource schema definitions
- HTTP client functionality with mocked responses
- Error handling logic
- Data transformation functions

**Run unit tests:**
```bash
make test
```

### 2. Integration Tests (Manual Local Testing)

Integration tests require a running Pocket-ID instance with manual setup.

**Prerequisites:**
1. Docker installed locally
2. A Pocket-ID instance running
3. An admin user with a registered passkey
4. An API key generated for testing

### 3. Acceptance Tests (Manual Local Testing)

Acceptance tests verify the full provider functionality against a real Pocket-ID instance.

## Local Testing Setup

### Step 1: Start Pocket-ID

Using Docker:
```bash
docker run -d \
  --name pocket-id-test \
  -p 8080:80 \
  -v pocket-id-test-data:/app/backend/data \
  -e PUBLIC_APP_URL=http://localhost:8080 \
  ghcr.io/pocket-id/pocket-id:latest
```

Or using Docker Compose:
```yaml
# docker-compose.yml
version: '3.8'
services:
  pocket-id:
    image: ghcr.io/pocket-id/pocket-id:latest
    ports:
      - "8080:80"
    environment:
      - PUBLIC_APP_URL=http://localhost:8080
    volumes:
      - pocket-id-data:/app/backend/data

volumes:
  pocket-id-data:
```

### Step 2: Initial Setup

1. **Access Pocket-ID**: Navigate to http://localhost:8080
2. **Create Admin User**: 
   - Click "Register" or "Sign Up"
   - Enter username and email
   - Register a passkey when prompted
3. **Log In**: Use your passkey to authenticate
4. **Generate API Key**:
   - Go to Settings → API Keys
   - Click "Create New API Key"
   - Give it a descriptive name (e.g., "Terraform Testing")
   - Copy the API key immediately (you won't see it again!)

### Step 3: Configure Environment

Set the required environment variables:

```bash
export POCKETID_BASE_URL="http://localhost:8080"
export POCKETID_API_TOKEN="your-api-key-here"
```

Or create a `.env.test` file:
```bash
POCKETID_BASE_URL=http://localhost:8080
POCKETID_API_TOKEN=your-api-key-here
```

### Step 4: Run Tests

**Run all tests (unit + acceptance):**
```bash
make test-all
```

**Run only acceptance tests:**
```bash
make test-acc
```

**Run specific acceptance tests:**
```bash
TF_ACC=1 go test -v ./internal/provider -tags=acc -run TestAccResourceClient
```

## Writing Tests

### Unit Test Example

```go
func TestClient_CreateUser(t *testing.T) {
    // Create mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "POST", r.Method)
        assert.Equal(t, "/api/users", r.URL.Path)
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(&User{
            ID:       "test-user-id",
            Username: "testuser",
        })
    }))
    defer server.Close()

    // Test client
    client := NewClient(server.URL, "test-token", false, 30)
    user, err := client.CreateUser(&UserCreateRequest{
        Username: "testuser",
    })
    
    assert.NoError(t, err)
    assert.Equal(t, "test-user-id", user.ID)
}
```

### Acceptance Test Example

```go
func TestAccResourceUser_basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccResourceUserConfig_basic("testuser"),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("pocketid_user.test", "username", "testuser"),
                    resource.TestCheckResourceAttrSet("pocketid_user.test", "id"),
                ),
            },
        },
    })
}
```

## Test Data Management

### Cleaning Up Test Data

After running acceptance tests, clean up test data:

```bash
# Remove all test resources
cd test && terraform destroy -auto-approve

# Or manually through the Pocket-ID UI
```

### Test Isolation

Each test should:
1. Use unique resource names (prefix with test name)
2. Clean up after itself
3. Not depend on other tests' data

## Debugging Tests

### Enable Debug Logging

```bash
# Terraform debug logs
export TF_LOG=DEBUG
export TF_LOG_PATH=terraform-test.log

# Provider debug logs
export TF_LOG_PROVIDER=DEBUG
```

### Common Issues

#### "401 Unauthorized"
- Check API token is valid
- Ensure token has admin privileges
- Verify token hasn't expired

#### "Connection refused"
- Ensure Pocket-ID is running
- Check the base URL is correct
- Verify no firewall blocking

#### "Resource not found"
- Resource may have been deleted manually
- Check for timing issues in tests
- Ensure proper test sequencing

## Continuous Testing Strategy

Since we can't run acceptance tests in CI, we recommend:

1. **Pre-Release Testing**:
   - Run full acceptance test suite locally before releases
   - Document test results in release notes
   - Maintain a test checklist

2. **Community Testing**:
   - Encourage users to test pre-releases
   - Provide clear testing instructions
   - Create a testing matrix for different scenarios

3. **Monitoring**:
   - Track issues reported by users
   - Monitor provider usage patterns
   - Regular compatibility testing with new Pocket-ID versions

## Test Checklist

Before releasing a new version, ensure:

- [ ] All unit tests pass
- [ ] Acceptance tests pass locally
- [ ] No linting errors
- [ ] Documentation is updated
- [ ] Examples work correctly
- [ ] Import functionality works
- [ ] Resource updates don't cause replacements
- [ ] Sensitive values are properly masked
- [ ] Error messages are helpful

## Alternative Testing Approaches

### Mock Pocket-ID Server

Consider creating a mock Pocket-ID server for testing that:
- Implements the same API endpoints
- Allows programmatic user/key creation
- Used only for CI testing

### Test Fixtures

Maintain test fixtures with:
- Pre-configured Pocket-ID Docker images
- Seeded test data
- Known API keys for testing

**Note**: This would require coordination with the Pocket-ID project to ensure test fixtures don't compromise security.

## Contributing

When submitting PRs:

1. Include unit tests for new functionality
2. Document any manual testing performed
3. Update this guide if testing procedures change
4. Note any acceptance test failures and explanations

## Summary

While Pocket-ID's security model prevents fully automated testing, a combination of comprehensive unit tests and thorough local acceptance testing ensures provider quality. The extra manual effort is a worthwhile trade-off for Pocket-ID's superior security architecture.