package resources_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
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

	// Verify computed attributes
	idAttr, ok := schemaResponse.Schema.Attributes["id"]
	assert.True(t, ok, "id attribute should exist")
	assert.True(t, idAttr.IsComputed(), "id should be computed")

	lastSyncAttr, ok := schemaResponse.Schema.Attributes["last_sync"]
	assert.True(t, ok, "last_sync attribute should exist")
	assert.True(t, lastSyncAttr.IsComputed(), "last_sync should be computed")

	// Verify optional attributes
	triggersAttr, ok := schemaResponse.Schema.Attributes["triggers"]
	assert.True(t, ok, "triggers attribute should exist")
	assert.True(t, triggersAttr.IsOptional(), "triggers should be optional")
}

func TestLDAPSyncResource_Metadata(t *testing.T) {
	ctx := context.Background()
	metadataRequest := resource.MetadataRequest{
		ProviderTypeName: "pocketid",
	}
	metadataResponse := &resource.MetadataResponse{}

	resources.NewLDAPSyncResource().Metadata(ctx, metadataRequest, metadataResponse)

	assert.Equal(t, "pocketid_ldap_sync", metadataResponse.TypeName)
}
