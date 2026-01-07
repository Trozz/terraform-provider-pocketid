package resources_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"

	"github.com/Trozz/terraform-provider-pocketid/internal/resources"
)

func TestLDAPConfigResource_Schema(t *testing.T) {
	ctx := context.Background()
	schemaRequest := resource.SchemaRequest{}
	schemaResponse := &resource.SchemaResponse{}

	resources.NewLDAPConfigResource().Schema(ctx, schemaRequest, schemaResponse)

	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema returned diagnostics: %+v", schemaResponse.Diagnostics)
	}

	// Verify required attributes
	enabledAttr, ok := schemaResponse.Schema.Attributes["enabled"]
	assert.True(t, ok, "enabled attribute should exist")
	assert.True(t, enabledAttr.IsRequired(), "enabled should be required")

	// Verify computed attributes
	idAttr, ok := schemaResponse.Schema.Attributes["id"]
	assert.True(t, ok, "id attribute should exist")
	assert.True(t, idAttr.IsComputed(), "id should be computed")

	// Verify sensitive attributes
	bindPasswordAttr, ok := schemaResponse.Schema.Attributes["bind_password"]
	assert.True(t, ok, "bind_password attribute should exist")
	assert.True(t, bindPasswordAttr.IsSensitive(), "bind_password should be sensitive")

	// Verify optional attributes with defaults
	skipCertVerifyAttr, ok := schemaResponse.Schema.Attributes["skip_cert_verify"]
	assert.True(t, ok, "skip_cert_verify attribute should exist")
	assert.True(t, skipCertVerifyAttr.IsOptional(), "skip_cert_verify should be optional")
	assert.True(t, skipCertVerifyAttr.IsComputed(), "skip_cert_verify should be computed")

	softDeleteUsersAttr, ok := schemaResponse.Schema.Attributes["soft_delete_users"]
	assert.True(t, ok, "soft_delete_users attribute should exist")
	assert.True(t, softDeleteUsersAttr.IsOptional(), "soft_delete_users should be optional")
	assert.True(t, softDeleteUsersAttr.IsComputed(), "soft_delete_users should be computed")

	// Verify nested attributes exist
	userAttributesAttr, ok := schemaResponse.Schema.Attributes["user_attributes"]
	assert.True(t, ok, "user_attributes attribute should exist")
	assert.True(t, userAttributesAttr.IsOptional(), "user_attributes should be optional")

	groupAttributesAttr, ok := schemaResponse.Schema.Attributes["group_attributes"]
	assert.True(t, ok, "group_attributes attribute should exist")
	assert.True(t, groupAttributesAttr.IsOptional(), "group_attributes should be optional")

	// Verify sync_on_change attribute
	syncOnChangeAttr, ok := schemaResponse.Schema.Attributes["sync_on_change"]
	assert.True(t, ok, "sync_on_change attribute should exist")
	assert.True(t, syncOnChangeAttr.IsOptional(), "sync_on_change should be optional")
	assert.True(t, syncOnChangeAttr.IsComputed(), "sync_on_change should be computed")
}

func TestLDAPConfigResource_Metadata(t *testing.T) {
	ctx := context.Background()
	metadataRequest := resource.MetadataRequest{
		ProviderTypeName: "pocketid",
	}
	metadataResponse := &resource.MetadataResponse{}

	resources.NewLDAPConfigResource().Metadata(ctx, metadataRequest, metadataResponse)

	assert.Equal(t, "pocketid_ldap_config", metadataResponse.TypeName)
}
