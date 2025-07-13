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

func TestLDAPTestDataSource_Metadata(t *testing.T) {
	ctx := context.Background()
	ds := datasources.NewLDAPTestDataSource()

	req := datasource.MetadataRequest{
		ProviderTypeName: "pocketid",
	}
	resp := &datasource.MetadataResponse{}

	ds.Metadata(ctx, req, resp)

	assert.Equal(t, "pocketid_ldap_test", resp.TypeName)
}

func TestLDAPTestDataSource_Schema(t *testing.T) {
	ctx := context.Background()
	ds := datasources.NewLDAPTestDataSource()

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	ds.Schema(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.NotNil(t, resp.Schema)
	assert.NotEmpty(t, resp.Schema.Description)
	assert.Contains(t, resp.Schema.Description, "Tests the current LDAP configuration")

	// Verify all computed attributes exist
	expectedAttributes := []string{
		"id", "connection_successful", "bind_successful",
		"base_dn_accessible", "users_found", "groups_found",
		"error_message", "test_timestamp",
	}

	for _, attr := range expectedAttributes {
		_, ok := resp.Schema.Attributes[attr]
		assert.True(t, ok, "Schema should have %s attribute", attr)
	}

	// Check attribute types and properties
	idAttr, ok := resp.Schema.Attributes["id"].(schema.StringAttribute)
	assert.True(t, ok, "id should be StringAttribute")
	assert.True(t, idAttr.Computed, "id should be computed")

	connectionSuccessfulAttr, ok := resp.Schema.Attributes["connection_successful"].(schema.BoolAttribute)
	assert.True(t, ok, "connection_successful should be BoolAttribute")
	assert.True(t, connectionSuccessfulAttr.Computed, "connection_successful should be computed")

	bindSuccessfulAttr, ok := resp.Schema.Attributes["bind_successful"].(schema.BoolAttribute)
	assert.True(t, ok, "bind_successful should be BoolAttribute")
	assert.True(t, bindSuccessfulAttr.Computed, "bind_successful should be computed")

	baseDNAccessibleAttr, ok := resp.Schema.Attributes["base_dn_accessible"].(schema.BoolAttribute)
	assert.True(t, ok, "base_dn_accessible should be BoolAttribute")
	assert.True(t, baseDNAccessibleAttr.Computed, "base_dn_accessible should be computed")

	testTimestampAttr, ok := resp.Schema.Attributes["test_timestamp"].(schema.StringAttribute)
	assert.True(t, ok, "test_timestamp should be StringAttribute")
	assert.True(t, testTimestampAttr.Computed, "test_timestamp should be computed")

	usersFoundAttr, ok := resp.Schema.Attributes["users_found"].(schema.Int64Attribute)
	assert.True(t, ok, "users_found should be Int64Attribute")
	assert.True(t, usersFoundAttr.Computed, "users_found should be computed")

	groupsFoundAttr, ok := resp.Schema.Attributes["groups_found"].(schema.Int64Attribute)
	assert.True(t, ok, "groups_found should be Int64Attribute")
	assert.True(t, groupsFoundAttr.Computed, "groups_found should be computed")

	errorMessageAttr, ok := resp.Schema.Attributes["error_message"].(schema.StringAttribute)
	assert.True(t, ok, "error_message should be StringAttribute")
	assert.True(t, errorMessageAttr.Computed, "error_message should be computed")
}

func TestLDAPTestDataSource_Configure(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name          string
		providerData  interface{}
		expectError   bool
		errorContains string
	}{
		{
			name:         "valid_client",
			providerData: &client.Client{},
			expectError:  false,
		},
		{
			name:         "nil_provider_data",
			providerData: nil,
			expectError:  false,
		},
		{
			name:          "invalid_provider_data_type",
			providerData:  "invalid",
			expectError:   true,
			errorContains: "Expected *client.Client",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds := datasources.NewLDAPTestDataSource()

			configurable, ok := ds.(datasource.DataSourceWithConfigure)
			require.True(t, ok, "LDAP test data source should implement DataSourceWithConfigure")

			req := datasource.ConfigureRequest{
				ProviderData: tc.providerData,
			}
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

func TestLDAPTestDataSource_SchemaDescriptions(t *testing.T) {
	ctx := context.Background()
	ds := datasources.NewLDAPTestDataSource()

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	ds.Schema(ctx, req, resp)

	// Verify all attributes have descriptions
	attrs := resp.Schema.Attributes

	for name, attr := range attrs {
		switch a := attr.(type) {
		case schema.StringAttribute:
			assert.NotEmpty(t, a.Description, "%s should have a description", name)
		case schema.BoolAttribute:
			assert.NotEmpty(t, a.Description, "%s should have a description", name)
		case schema.Int64Attribute:
			assert.NotEmpty(t, a.Description, "%s should have a description", name)
		default:
			t.Errorf("Unexpected attribute type for %s", name)
		}
	}
}

func TestLDAPTestDataSource_MarkdownDescription(t *testing.T) {
	ctx := context.Background()
	ds := datasources.NewLDAPTestDataSource()

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	ds.Schema(ctx, req, resp)

	// Check that MarkdownDescription is set
	assert.NotEmpty(t, resp.Schema.MarkdownDescription, "Schema should have MarkdownDescription")
	assert.Contains(t, resp.Schema.MarkdownDescription, "Tests the current LDAP configuration")
}

// Test that LDAP test data source is included in the list of data sources to test
func TestLDAPTestDataSource_IncludedInList(t *testing.T) {
	ctx := context.Background()
	ds := datasources.NewLDAPTestDataSource()

	// Verify it has proper metadata
	req := datasource.MetadataRequest{
		ProviderTypeName: "pocketid",
	}
	resp := &datasource.MetadataResponse{}
	ds.Metadata(ctx, req, resp)

	assert.Equal(t, "pocketid_ldap_test", resp.TypeName)

	// Verify it has a schema
	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(ctx, schemaReq, schemaResp)

	assert.False(t, schemaResp.Diagnostics.HasError())
	assert.NotNil(t, schemaResp.Schema)
}
