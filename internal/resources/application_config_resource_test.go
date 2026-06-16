package resources_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
	"github.com/Trozz/terraform-provider-pocketid/internal/resources"
)

func TestNewApplicationConfigResource(t *testing.T) {
	r := resources.NewApplicationConfigResource()
	assert.NotNil(t, r)

	_, ok := r.(resource.ResourceWithConfigure)
	assert.True(t, ok, "should implement ResourceWithConfigure")

	_, ok = r.(resource.ResourceWithImportState)
	assert.True(t, ok, "should implement ResourceWithImportState")
}

func TestApplicationConfigResource_Metadata(t *testing.T) {
	ctx := context.Background()
	r := resources.NewApplicationConfigResource()

	req := resource.MetadataRequest{ProviderTypeName: "pocketid"}
	resp := &resource.MetadataResponse{}

	r.Metadata(ctx, req, resp)

	assert.Equal(t, "pocketid_application_config", resp.TypeName)
}

func TestApplicationConfigResource_Schema(t *testing.T) {
	ctx := context.Background()
	r := resources.NewApplicationConfigResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(ctx, req, resp)

	require.False(t, resp.Diagnostics.HasError())
	assert.NotEmpty(t, resp.Schema.Description)

	// id is computed only.
	idAttr, ok := resp.Schema.Attributes["id"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, idAttr.Computed)
	assert.False(t, idAttr.Optional)

	// Writable attributes are optional + computed.
	for _, name := range []string{"app_name", "session_duration", "ldap_enabled"} {
		attr, ok := resp.Schema.Attributes[name].(schema.StringAttribute)
		require.True(t, ok, "attribute %s should exist", name)
		assert.True(t, attr.Optional, "attribute %s should be optional", name)
		assert.True(t, attr.Computed, "attribute %s should be computed", name)
	}

	// Secrets are marked sensitive.
	for _, name := range []string{"smtp_password", "ldap_bind_password"} {
		attr, ok := resp.Schema.Attributes[name].(schema.StringAttribute)
		require.True(t, ok, "attribute %s should exist", name)
		assert.True(t, attr.Sensitive, "attribute %s should be sensitive", name)
	}
}

func TestApplicationConfigResource_Configure(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name          string
		providerData  interface{}
		expectError   bool
		errorContains string
	}{
		{name: "valid_client", providerData: &client.Client{}, expectError: false},
		{name: "nil_provider_data", providerData: nil, expectError: false},
		{name: "invalid_type", providerData: "invalid", expectError: true, errorContains: "Expected *client.Client"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := resources.NewApplicationConfigResource()
			configurable, ok := r.(resource.ResourceWithConfigure)
			require.True(t, ok)

			req := resource.ConfigureRequest{ProviderData: tc.providerData}
			resp := &resource.ConfigureResponse{}

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
