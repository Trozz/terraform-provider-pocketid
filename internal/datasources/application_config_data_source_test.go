package datasources_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
	"github.com/Trozz/terraform-provider-pocketid/internal/datasources"
)

func TestNewApplicationConfigDataSource(t *testing.T) {
	d := datasources.NewApplicationConfigDataSource()
	assert.NotNil(t, d)

	_, ok := d.(datasource.DataSourceWithConfigure)
	assert.True(t, ok, "should implement DataSourceWithConfigure")
}

func TestApplicationConfigDataSource_Metadata(t *testing.T) {
	ctx := context.Background()
	d := datasources.NewApplicationConfigDataSource()

	req := datasource.MetadataRequest{ProviderTypeName: "pocketid"}
	resp := &datasource.MetadataResponse{}

	d.Metadata(ctx, req, resp)

	assert.Equal(t, "pocketid_application_config", resp.TypeName)
}

func TestApplicationConfigDataSource_Schema(t *testing.T) {
	ctx := context.Background()
	d := datasources.NewApplicationConfigDataSource()

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(ctx, req, resp)

	require.False(t, resp.Diagnostics.HasError())
	assert.NotEmpty(t, resp.Schema.Description)

	// All attributes are computed.
	for _, name := range []string{"id", "app_name", "ldap_enabled"} {
		attr, ok := resp.Schema.Attributes[name].(schema.StringAttribute)
		require.True(t, ok, "attribute %s should exist", name)
		assert.True(t, attr.Computed, "attribute %s should be computed", name)
	}

	// Secrets are marked sensitive.
	for _, name := range []string{"smtp_password", "ldap_bind_password"} {
		attr, ok := resp.Schema.Attributes[name].(schema.StringAttribute)
		require.True(t, ok, "attribute %s should exist", name)
		assert.True(t, attr.Sensitive, "attribute %s should be sensitive", name)
	}
}

func TestApplicationConfigDataSource_Configure(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name          string
		providerData  interface{}
		expectError   bool
		errorContains string
	}{
		{name: "valid_client", providerData: &client.Client{}, expectError: false},
		{name: "nil_provider_data", providerData: nil, expectError: false},
		{name: "invalid_type", providerData: 123, expectError: true, errorContains: "Expected *client.Client"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := datasources.NewApplicationConfigDataSource()
			configurable, ok := d.(datasource.DataSourceWithConfigure)
			require.True(t, ok)

			req := datasource.ConfigureRequest{ProviderData: tc.providerData}
			resp := &datasource.ConfigureResponse{}

			configurable.Configure(ctx, req, resp)

			if tc.expectError {
				assert.True(t, resp.Diagnostics.HasError())
				assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), tc.errorContains)
			} else {
				assert.False(t, resp.Diagnostics.HasError())
			}
		})
	}
}
