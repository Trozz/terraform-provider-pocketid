package resources_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/assert"

	"github.com/Trozz/terraform-provider-pocketid/internal/resources"
)

func TestLDAPSyncResource_Schema(t *testing.T) {
	ctx := context.Background()
	schemaRequest := resource.SchemaRequest{}
	schemaResponse := &resource.SchemaResponse{}

	resources.NewLDAPSyncResource().Schema(ctx, schemaRequest, schemaResponse)

	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema returned diagnostics: %+v", schemaResponse.Diagnostics)
	}

	// Verify optional attributes
	triggersAttr, ok := schemaResponse.Schema.Attributes["triggers"]
	assert.True(t, ok, "triggers attribute should exist")
	assert.True(t, triggersAttr.IsOptional(), "triggers should be optional")

	// Verify computed attributes
	idAttr, ok := schemaResponse.Schema.Attributes["id"]
	assert.True(t, ok, "id attribute should exist")
	assert.True(t, idAttr.IsComputed(), "id should be computed")

	statusAttr, ok := schemaResponse.Schema.Attributes["status"]
	assert.True(t, ok, "status attribute should exist")
	assert.True(t, statusAttr.IsComputed(), "status should be computed")

	lastSyncAttr, ok := schemaResponse.Schema.Attributes["last_sync"]
	assert.True(t, ok, "last_sync attribute should exist")
	assert.True(t, lastSyncAttr.IsComputed(), "last_sync should be computed")

	errorAttr, ok := schemaResponse.Schema.Attributes["error"]
	assert.True(t, ok, "error attribute should exist")
	assert.True(t, errorAttr.IsComputed(), "error should be computed")
}

func TestLDAPSyncResource_SchemaValidation(t *testing.T) {
	ctx := context.Background()
	r := resources.NewLDAPSyncResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())

	// Verify all expected attributes exist
	attrs := resp.Schema.Attributes

	// Check triggers attribute is a map
	triggersAttr, ok := attrs["triggers"].(schema.MapAttribute)
	assert.True(t, ok, "triggers should be MapAttribute")
	assert.True(t, triggersAttr.Optional, "triggers should be optional")

	// Check element type is string
	assert.NotNil(t, triggersAttr.ElementType, "triggers should have an element type")

	// Check computed string attributes
	statusAttr, ok := attrs["status"].(schema.StringAttribute)
	assert.True(t, ok, "status should be StringAttribute")
	assert.True(t, statusAttr.Computed, "status should be computed")

	lastSyncAttr, ok := attrs["last_sync"].(schema.StringAttribute)
	assert.True(t, ok, "last_sync should be StringAttribute")
	assert.True(t, lastSyncAttr.Computed, "last_sync should be computed")

	errorAttr, ok := attrs["error"].(schema.StringAttribute)
	assert.True(t, ok, "error should be StringAttribute")
	assert.True(t, errorAttr.Computed, "error should be computed")
}

func TestLDAPSyncResource_Metadata(t *testing.T) {
	ctx := context.Background()
	r := resources.NewLDAPSyncResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "pocketid",
	}
	resp := &resource.MetadataResponse{}
	r.Metadata(ctx, req, resp)

	assert.Equal(t, "pocketid_ldap_sync", resp.TypeName)
}

func TestLDAPSyncResource_Configure(t *testing.T) {
	ctx := context.Background()
	r := resources.NewLDAPSyncResource()
	configurable := r.(resource.ResourceWithConfigure)

	// Test with nil provider data
	t.Run("NilProviderData", func(t *testing.T) {
		req := resource.ConfigureRequest{
			ProviderData: nil,
		}
		resp := &resource.ConfigureResponse{}

		configurable.Configure(ctx, req, resp)
		assert.False(t, resp.Diagnostics.HasError())
	})

	// Test with invalid provider data type
	t.Run("InvalidProviderDataType", func(t *testing.T) {
		req := resource.ConfigureRequest{
			ProviderData: "invalid-type",
		}
		resp := &resource.ConfigureResponse{}

		configurable.Configure(ctx, req, resp)
		assert.True(t, resp.Diagnostics.HasError())
		assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "Expected *client.Client")
	})
}

func TestLDAPSyncResource_TriggersDescription(t *testing.T) {
	ctx := context.Background()
	r := resources.NewLDAPSyncResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, req, resp)

	// Check that triggers has proper description
	triggersAttr, ok := resp.Schema.Attributes["triggers"].(schema.MapAttribute)
	assert.True(t, ok, "triggers should be MapAttribute")
	assert.NotEmpty(t, triggersAttr.Description, "triggers should have a description")
	assert.Contains(t, triggersAttr.Description, "Map of values", "triggers description should explain its purpose")
}

func TestLDAPSyncResource_SchemaDescriptions(t *testing.T) {
	ctx := context.Background()
	r := resources.NewLDAPSyncResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, req, resp)

	// Check resource description
	assert.NotEmpty(t, resp.Schema.Description, "Resource should have a description")
	assert.Contains(t, resp.Schema.Description, "LDAP sync", "Resource description should mention LDAP sync")

	// Check attribute descriptions
	attrs := resp.Schema.Attributes

	statusAttr, _ := attrs["status"].(schema.StringAttribute)
	assert.NotEmpty(t, statusAttr.Description, "status should have a description")

	lastSyncAttr, _ := attrs["last_sync"].(schema.StringAttribute)
	assert.NotEmpty(t, lastSyncAttr.Description, "last_sync should have a description")

	errorAttr, _ := attrs["error"].(schema.StringAttribute)
	assert.NotEmpty(t, errorAttr.Description, "error should have a description")
}
