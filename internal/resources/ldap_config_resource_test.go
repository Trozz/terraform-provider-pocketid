package resources_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

	// Verify optional attributes
	urlAttr, ok := schemaResponse.Schema.Attributes["url"]
	assert.True(t, ok, "url attribute should exist")
	assert.True(t, urlAttr.IsOptional(), "url should be optional")

	bindDNAttr, ok := schemaResponse.Schema.Attributes["bind_dn"]
	assert.True(t, ok, "bind_dn attribute should exist")
	assert.True(t, bindDNAttr.IsOptional(), "bind_dn should be optional")

	bindPasswordAttr, ok := schemaResponse.Schema.Attributes["bind_password"]
	assert.True(t, ok, "bind_password attribute should exist")
	assert.True(t, bindPasswordAttr.IsOptional(), "bind_password should be optional")
	assert.True(t, bindPasswordAttr.IsSensitive(), "bind_password should be sensitive")

	baseDNAttr, ok := schemaResponse.Schema.Attributes["base_dn"]
	assert.True(t, ok, "base_dn attribute should exist")
	assert.True(t, baseDNAttr.IsOptional(), "base_dn should be optional")

	// Verify computed attributes
	idAttr, ok := schemaResponse.Schema.Attributes["id"]
	assert.True(t, ok, "id attribute should exist")
	assert.True(t, idAttr.IsComputed(), "id should be computed")

	// Verify nested blocks
	_, ok = schemaResponse.Schema.Blocks["user_attributes"]
	assert.True(t, ok, "user_attributes block should exist")

	_, ok = schemaResponse.Schema.Blocks["group_attributes"]
	assert.True(t, ok, "group_attributes block should exist")
}

func TestLDAPConfigResource_SchemaValidation(t *testing.T) {
	ctx := context.Background()
	r := resources.NewLDAPConfigResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())

	// Verify all expected attributes exist
	attrs := resp.Schema.Attributes

	// Check boolean attributes with defaults
	skipCertVerifyAttr, ok := attrs["skip_cert_verify"].(schema.BoolAttribute)
	assert.True(t, ok, "skip_cert_verify should be BoolAttribute")
	assert.True(t, skipCertVerifyAttr.Optional, "skip_cert_verify should be optional")
	assert.True(t, skipCertVerifyAttr.Computed, "skip_cert_verify should be computed")

	softDeleteUsersAttr, ok := attrs["soft_delete_users"].(schema.BoolAttribute)
	assert.True(t, ok, "soft_delete_users should be BoolAttribute")
	assert.True(t, softDeleteUsersAttr.Optional, "soft_delete_users should be optional")
	assert.True(t, softDeleteUsersAttr.Computed, "soft_delete_users should be computed")

	// Check string attributes
	userSearchFilterAttr, ok := attrs["user_search_filter"].(schema.StringAttribute)
	assert.True(t, ok, "user_search_filter should be StringAttribute")
	assert.True(t, userSearchFilterAttr.Optional, "user_search_filter should be optional")

	userGroupSearchFilterAttr, ok := attrs["user_group_search_filter"].(schema.StringAttribute)
	assert.True(t, ok, "user_group_search_filter should be StringAttribute")
	assert.True(t, userGroupSearchFilterAttr.Optional, "user_group_search_filter should be optional")
}

func TestLDAPConfigResource_Metadata(t *testing.T) {
	ctx := context.Background()
	r := resources.NewLDAPConfigResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "pocketid",
	}
	resp := &resource.MetadataResponse{}
	r.Metadata(ctx, req, resp)

	assert.Equal(t, "pocketid_ldap_config", resp.TypeName)
}

func TestLDAPConfigResource_Configure(t *testing.T) {
	ctx := context.Background()
	r := resources.NewLDAPConfigResource()
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

func TestLDAPConfigResource_UserAttributesBlock(t *testing.T) {
	ctx := context.Background()
	r := resources.NewLDAPConfigResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, req, resp)

	// Get user_attributes block
	userAttrsBlock, ok := resp.Schema.Blocks["user_attributes"].(schema.SingleNestedBlock)
	assert.True(t, ok, "user_attributes should be SingleNestedBlock")

	// Check all user attribute fields
	attrs := userAttrsBlock.Attributes

	uniqueIdAttr, ok := attrs["unique_identifier"].(schema.StringAttribute)
	assert.True(t, ok, "unique_identifier should be StringAttribute")
	assert.True(t, uniqueIdAttr.Optional, "unique_identifier should be optional")

	usernameAttr, ok := attrs["username"].(schema.StringAttribute)
	assert.True(t, ok, "username should be StringAttribute")
	assert.True(t, usernameAttr.Optional, "username should be optional")

	emailAttr, ok := attrs["email"].(schema.StringAttribute)
	assert.True(t, ok, "email should be StringAttribute")
	assert.True(t, emailAttr.Optional, "email should be optional")

	firstNameAttr, ok := attrs["first_name"].(schema.StringAttribute)
	assert.True(t, ok, "first_name should be StringAttribute")
	assert.True(t, firstNameAttr.Optional, "first_name should be optional")

	lastNameAttr, ok := attrs["last_name"].(schema.StringAttribute)
	assert.True(t, ok, "last_name should be StringAttribute")
	assert.True(t, lastNameAttr.Optional, "last_name should be optional")

	profilePictureAttr, ok := attrs["profile_picture"].(schema.StringAttribute)
	assert.True(t, ok, "profile_picture should be StringAttribute")
	assert.True(t, profilePictureAttr.Optional, "profile_picture should be optional")
}

func TestLDAPConfigResource_GroupAttributesBlock(t *testing.T) {
	ctx := context.Background()
	r := resources.NewLDAPConfigResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, req, resp)

	// Get group_attributes block
	groupAttrsBlock, ok := resp.Schema.Blocks["group_attributes"].(schema.SingleNestedBlock)
	assert.True(t, ok, "group_attributes should be SingleNestedBlock")

	// Check all group attribute fields
	attrs := groupAttrsBlock.Attributes

	memberAttr, ok := attrs["member"].(schema.StringAttribute)
	assert.True(t, ok, "member should be StringAttribute")
	assert.True(t, memberAttr.Optional, "member should be optional")

	uniqueIdAttr, ok := attrs["unique_identifier"].(schema.StringAttribute)
	assert.True(t, ok, "unique_identifier should be StringAttribute")
	assert.True(t, uniqueIdAttr.Optional, "unique_identifier should be optional")

	nameAttr, ok := attrs["name"].(schema.StringAttribute)
	assert.True(t, ok, "name should be StringAttribute")
	assert.True(t, nameAttr.Optional, "name should be optional")

	adminGroupNameAttr, ok := attrs["admin_group_name"].(schema.StringAttribute)
	assert.True(t, ok, "admin_group_name should be StringAttribute")
	assert.True(t, adminGroupNameAttr.Optional, "admin_group_name should be optional")
}

func TestLDAPConfigResource_ImportState(t *testing.T) {
	ctx := context.Background()
	r := resources.NewLDAPConfigResource()

	// Check if resource implements ImportState
	importable, ok := r.(resource.ResourceWithImportState)
	assert.True(t, ok, "LDAP config resource should be importable")

	// The actual import logic would be tested in acceptance tests
	_ = importable
	_ = ctx
}
