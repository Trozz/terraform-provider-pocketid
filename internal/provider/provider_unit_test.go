package provider_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pocketidprovider "github.com/Trozz/terraform-provider-pocketid/internal/provider"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name    string
		version string
	}{
		{
			name:    "with_version",
			version: "1.0.0",
		},
		{
			name:    "with_dev_version",
			version: "dev",
		},
		{
			name:    "with_test_version",
			version: "test",
		},
		{
			name:    "empty_version",
			version: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			providerFunc := pocketidprovider.New(tc.version)
			assert.NotNil(t, providerFunc)

			provider := providerFunc()
			assert.NotNil(t, provider)
		})
	}
}

func TestProvider_Metadata(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name            string
		version         string
		expectedVersion string
	}{
		{
			name:            "version_1.0.0",
			version:         "1.0.0",
			expectedVersion: "1.0.0",
		},
		{
			name:            "version_dev",
			version:         "dev",
			expectedVersion: "dev",
		},
		{
			name:            "version_test",
			version:         "test",
			expectedVersion: "test",
		},
		{
			name:            "empty_version",
			version:         "",
			expectedVersion: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := pocketidprovider.New(tc.version)()

			req := provider.MetadataRequest{}
			resp := &provider.MetadataResponse{}

			p.Metadata(ctx, req, resp)

			assert.Equal(t, "pocketid", resp.TypeName)
			assert.Equal(t, tc.expectedVersion, resp.Version)
		})
	}
}

func TestProvider_Schema(t *testing.T) {
	ctx := context.Background()
	p := pocketidprovider.New("test")()

	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.NotNil(t, resp.Schema)
	assert.NotEmpty(t, resp.Schema.Description)
	assert.Contains(t, resp.Schema.Description, "Pocket-ID")

	// Check all expected attributes
	expectedAttributes := []string{
		"base_url",
		"api_token",
		"skip_tls_verify",
		"timeout",
	}

	for _, attrName := range expectedAttributes {
		t.Run(attrName, func(t *testing.T) {
			attr, ok := resp.Schema.Attributes[attrName]
			assert.True(t, ok, "Schema should have %s attribute", attrName)
			assert.NotNil(t, attr)
		})
	}

	// Check specific attribute properties
	baseURLAttr := resp.Schema.Attributes["base_url"].(schema.StringAttribute)
	assert.True(t, baseURLAttr.Optional)
	assert.Contains(t, baseURLAttr.Description, "POCKETID_BASE_URL")

	apiTokenAttr := resp.Schema.Attributes["api_token"].(schema.StringAttribute)
	assert.True(t, apiTokenAttr.Optional)
	assert.True(t, apiTokenAttr.Sensitive)
	assert.Contains(t, apiTokenAttr.Description, "POCKETID_API_TOKEN")

	skipTLSAttr := resp.Schema.Attributes["skip_tls_verify"].(schema.BoolAttribute)
	assert.True(t, skipTLSAttr.Optional)

	timeoutAttr := resp.Schema.Attributes["timeout"].(schema.Int64Attribute)
	assert.True(t, timeoutAttr.Optional)
}

func TestProvider_Configure(t *testing.T) {
	// Save current env vars
	originalBaseURL := os.Getenv("POCKETID_BASE_URL")
	originalAPIToken := os.Getenv("POCKETID_API_TOKEN")

	// Restore env vars after test
	t.Cleanup(func() {
		_ = os.Setenv("POCKETID_BASE_URL", originalBaseURL)
		_ = os.Setenv("POCKETID_API_TOKEN", originalAPIToken)
	})

	testCases := []struct {
		name          string
		config        map[string]tftypes.Value
		envVars       map[string]string
		expectError   bool
		errorContains []string
	}{
		{
			name: "valid_config",
			config: map[string]tftypes.Value{
				"base_url":        tftypes.NewValue(tftypes.String, "https://pocketid.example.com"),
				"api_token":       tftypes.NewValue(tftypes.String, "test-token"),
				"skip_tls_verify": tftypes.NewValue(tftypes.Bool, false),
				"timeout":         tftypes.NewValue(tftypes.Number, 30),
			},
			expectError: false,
		},
		{
			name: "valid_config_with_env_vars",
			config: map[string]tftypes.Value{
				"base_url":        tftypes.NewValue(tftypes.String, nil),
				"api_token":       tftypes.NewValue(tftypes.String, nil),
				"skip_tls_verify": tftypes.NewValue(tftypes.Bool, nil),
				"timeout":         tftypes.NewValue(tftypes.Number, nil),
			},
			envVars: map[string]string{
				"POCKETID_BASE_URL":  "https://pocketid.example.com",
				"POCKETID_API_TOKEN": "env-token",
			},
			expectError: false,
		},
		{
			name: "config_overrides_env_vars",
			config: map[string]tftypes.Value{
				"base_url":        tftypes.NewValue(tftypes.String, "https://config.example.com"),
				"api_token":       tftypes.NewValue(tftypes.String, "config-token"),
				"skip_tls_verify": tftypes.NewValue(tftypes.Bool, nil),
				"timeout":         tftypes.NewValue(tftypes.Number, nil),
			},
			envVars: map[string]string{
				"POCKETID_BASE_URL":  "https://env.example.com",
				"POCKETID_API_TOKEN": "env-token",
			},
			expectError: false,
		},
		{
			name: "missing_base_url",
			config: map[string]tftypes.Value{
				"base_url":        tftypes.NewValue(tftypes.String, nil),
				"api_token":       tftypes.NewValue(tftypes.String, "test-token"),
				"skip_tls_verify": tftypes.NewValue(tftypes.Bool, nil),
				"timeout":         tftypes.NewValue(tftypes.Number, nil),
			},
			expectError:   true,
			errorContains: []string{"Missing Pocket-ID Base URL"},
		},
		{
			name: "empty_base_url",
			config: map[string]tftypes.Value{
				"base_url":        tftypes.NewValue(tftypes.String, ""),
				"api_token":       tftypes.NewValue(tftypes.String, "test-token"),
				"skip_tls_verify": tftypes.NewValue(tftypes.Bool, nil),
				"timeout":         tftypes.NewValue(tftypes.Number, nil),
			},
			expectError:   true,
			errorContains: []string{"Missing Pocket-ID Base URL"},
		},
		{
			name: "missing_api_token",
			config: map[string]tftypes.Value{
				"base_url":        tftypes.NewValue(tftypes.String, "https://pocketid.example.com"),
				"api_token":       tftypes.NewValue(tftypes.String, nil),
				"skip_tls_verify": tftypes.NewValue(tftypes.Bool, nil),
				"timeout":         tftypes.NewValue(tftypes.Number, nil),
			},
			expectError:   true,
			errorContains: []string{"Missing Pocket-ID API Token"},
		},
		{
			name: "empty_api_token",
			config: map[string]tftypes.Value{
				"base_url":        tftypes.NewValue(tftypes.String, "https://pocketid.example.com"),
				"api_token":       tftypes.NewValue(tftypes.String, ""),
				"skip_tls_verify": tftypes.NewValue(tftypes.Bool, nil),
				"timeout":         tftypes.NewValue(tftypes.Number, nil),
			},
			expectError:   true,
			errorContains: []string{"Missing Pocket-ID API Token"},
		},
		{
			name: "unknown_base_url",
			config: map[string]tftypes.Value{
				"base_url":        tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"api_token":       tftypes.NewValue(tftypes.String, "test-token"),
				"skip_tls_verify": tftypes.NewValue(tftypes.Bool, nil),
				"timeout":         tftypes.NewValue(tftypes.Number, nil),
			},
			expectError:   true,
			errorContains: []string{"Unknown Pocket-ID Base URL"},
		},
		{
			name: "unknown_api_token",
			config: map[string]tftypes.Value{
				"base_url":        tftypes.NewValue(tftypes.String, "https://pocketid.example.com"),
				"api_token":       tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"skip_tls_verify": tftypes.NewValue(tftypes.Bool, nil),
				"timeout":         tftypes.NewValue(tftypes.Number, nil),
			},
			expectError:   true,
			errorContains: []string{"Unknown Pocket-ID API Token"},
		},
		{
			name: "skip_tls_verify_true",
			config: map[string]tftypes.Value{
				"base_url":        tftypes.NewValue(tftypes.String, "https://pocketid.example.com"),
				"api_token":       tftypes.NewValue(tftypes.String, "test-token"),
				"skip_tls_verify": tftypes.NewValue(tftypes.Bool, true),
				"timeout":         tftypes.NewValue(tftypes.Number, nil),
			},
			expectError: false,
		},
		{
			name: "custom_timeout",
			config: map[string]tftypes.Value{
				"base_url":        tftypes.NewValue(tftypes.String, "https://pocketid.example.com"),
				"api_token":       tftypes.NewValue(tftypes.String, "test-token"),
				"skip_tls_verify": tftypes.NewValue(tftypes.Bool, nil),
				"timeout":         tftypes.NewValue(tftypes.Number, 60),
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear env vars
			_ = os.Unsetenv("POCKETID_BASE_URL")
			_ = os.Unsetenv("POCKETID_API_TOKEN")

			// Set test env vars
			for key, value := range tc.envVars {
				_ = os.Setenv(key, value)
			}

			ctx := context.Background()
			p := pocketidprovider.New("test")()

			// Get schema first
			schemaReq := provider.SchemaRequest{}
			schemaResp := &provider.SchemaResponse{}
			p.Schema(ctx, schemaReq, schemaResp)
			require.False(t, schemaResp.Diagnostics.HasError())

			// Create config
			configValue := tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"base_url":        tftypes.String,
					"api_token":       tftypes.String,
					"skip_tls_verify": tftypes.Bool,
					"timeout":         tftypes.Number,
				},
			}, tc.config)

			config := tfsdk.Config{
				Raw:    configValue,
				Schema: schemaResp.Schema,
			}

			// Configure provider
			req := provider.ConfigureRequest{
				Config: config,
			}
			resp := &provider.ConfigureResponse{}

			p.Configure(ctx, req, resp)

			if tc.expectError {
				assert.True(t, resp.Diagnostics.HasError())
				for _, expected := range tc.errorContains {
					found := false
					for _, diag := range resp.Diagnostics.Errors() {
						if strings.Contains(diag.Summary(), expected) || strings.Contains(diag.Detail(), expected) {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected error containing '%s' not found", expected)
				}
			} else {
				assert.False(t, resp.Diagnostics.HasError())
				assert.NotNil(t, resp.DataSourceData)
				assert.NotNil(t, resp.ResourceData)
			}
		})
	}
}

func TestProvider_DataSources(t *testing.T) {
	ctx := context.Background()
	p := pocketidprovider.New("test")()

	dataSources := p.DataSources(ctx)

	// Should have 4 data sources
	assert.Len(t, dataSources, 4)

	// Verify each data source can be created
	for i, dsFunc := range dataSources {
		t.Run(fmt.Sprintf("data_source_%d", i), func(t *testing.T) {
			ds := dsFunc()
			assert.NotNil(t, ds)
		})
	}
}

func TestProvider_Resources(t *testing.T) {
	ctx := context.Background()
	p := pocketidprovider.New("test")()

	resources := p.Resources(ctx)

	// Should have 3 resources
	assert.Len(t, resources, 3)

	// Verify each resource can be created
	for i, resFunc := range resources {
		t.Run(fmt.Sprintf("resource_%d", i), func(t *testing.T) {
			res := resFunc()
			assert.NotNil(t, res)
		})
	}
}

func TestProvider_Configure_EdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("all_null_values", func(t *testing.T) {
		p := pocketidprovider.New("test")()

		// Get schema first
		schemaReq := provider.SchemaRequest{}
		schemaResp := &provider.SchemaResponse{}
		p.Schema(ctx, schemaReq, schemaResp)
		require.False(t, schemaResp.Diagnostics.HasError())

		// Create config with all null values
		configValue := tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"base_url":        tftypes.String,
				"api_token":       tftypes.String,
				"skip_tls_verify": tftypes.Bool,
				"timeout":         tftypes.Number,
			},
		}, map[string]tftypes.Value{
			"base_url":        tftypes.NewValue(tftypes.String, nil),
			"api_token":       tftypes.NewValue(tftypes.String, nil),
			"skip_tls_verify": tftypes.NewValue(tftypes.Bool, nil),
			"timeout":         tftypes.NewValue(tftypes.Number, nil),
		})

		config := tfsdk.Config{
			Raw:    configValue,
			Schema: schemaResp.Schema,
		}

		req := provider.ConfigureRequest{
			Config: config,
		}
		resp := &provider.ConfigureResponse{}

		p.Configure(ctx, req, resp)
		assert.True(t, resp.Diagnostics.HasError())
		assert.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Missing Pocket-ID Base URL")
	})
}

// Helper function to test different provider versions
func TestProvider_DifferentVersions(t *testing.T) {
	versions := []string{"1.0.0", "2.0.0-beta", "dev", "test", "", "v1.2.3"}

	for _, version := range versions {
		t.Run(version, func(t *testing.T) {
			ctx := context.Background()
			p := pocketidprovider.New(version)()

			// Test metadata returns correct version
			metaReq := provider.MetadataRequest{}
			metaResp := &provider.MetadataResponse{}
			p.Metadata(ctx, metaReq, metaResp)
			assert.Equal(t, version, metaResp.Version)

			// Test schema works with any version
			schemaReq := provider.SchemaRequest{}
			schemaResp := &provider.SchemaResponse{}
			p.Schema(ctx, schemaReq, schemaResp)
			assert.False(t, schemaResp.Diagnostics.HasError())

			// Test data sources and resources work with any version
			assert.NotEmpty(t, p.DataSources(ctx))
			assert.NotEmpty(t, p.Resources(ctx))
		})
	}
}
