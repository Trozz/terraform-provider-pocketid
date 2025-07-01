package resources_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"

	"github.com/trozz/terraform-provider-pocketid/internal/resources"
)

func TestClientResource_Schema(t *testing.T) {
	ctx := context.Background()
	schemaRequest := resource.SchemaRequest{}
	schemaResponse := &resource.SchemaResponse{}

	resources.NewClientResource().Schema(ctx, schemaRequest, schemaResponse)

	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema returned diagnostics: %+v", schemaResponse.Diagnostics)
	}

	// Verify required attributes
	nameAttr, ok := schemaResponse.Schema.Attributes["name"]
	assert.True(t, ok, "name attribute should exist")
	assert.True(t, nameAttr.IsRequired(), "name should be required")

	callbackURLsAttr, ok := schemaResponse.Schema.Attributes["callback_urls"]
	assert.True(t, ok, "callback_urls attribute should exist")
	assert.True(t, callbackURLsAttr.IsRequired(), "callback_urls should be required")

	// Verify computed attributes
	idAttr, ok := schemaResponse.Schema.Attributes["id"]
	assert.True(t, ok, "id attribute should exist")
	assert.True(t, idAttr.IsComputed(), "id should be computed")

	clientSecretAttr, ok := schemaResponse.Schema.Attributes["client_secret"]
	assert.True(t, ok, "client_secret attribute should exist")
	assert.True(t, clientSecretAttr.IsComputed(), "client_secret should be computed")
	assert.True(t, clientSecretAttr.IsSensitive(), "client_secret should be sensitive")

	// Verify optional attributes with defaults
	isPublicAttr, ok := schemaResponse.Schema.Attributes["is_public"]
	assert.True(t, ok, "is_public attribute should exist")
	assert.True(t, isPublicAttr.IsOptional(), "is_public should be optional")
	assert.True(t, isPublicAttr.IsComputed(), "is_public should be computed")

	pkceEnabledAttr, ok := schemaResponse.Schema.Attributes["pkce_enabled"]
	assert.True(t, ok, "pkce_enabled attribute should exist")
	assert.True(t, pkceEnabledAttr.IsOptional(), "pkce_enabled should be optional")
	assert.True(t, pkceEnabledAttr.IsComputed(), "pkce_enabled should be computed")
}

func TestGroupResource_Schema(t *testing.T) {
	ctx := context.Background()
	schemaRequest := resource.SchemaRequest{}
	schemaResponse := &resource.SchemaResponse{}

	resources.NewGroupResource().Schema(ctx, schemaRequest, schemaResponse)

	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema returned diagnostics: %+v", schemaResponse.Diagnostics)
	}

	// Verify required attributes
	nameAttr, ok := schemaResponse.Schema.Attributes["name"]
	assert.True(t, ok, "name attribute should exist")
	assert.True(t, nameAttr.IsRequired(), "name should be required")

	friendlyNameAttr, ok := schemaResponse.Schema.Attributes["friendly_name"]
	assert.True(t, ok, "friendly_name attribute should exist")
	assert.True(t, friendlyNameAttr.IsRequired(), "friendly_name should be required")

	// Verify computed attributes
	idAttr, ok := schemaResponse.Schema.Attributes["id"]
	assert.True(t, ok, "id attribute should exist")
	assert.True(t, idAttr.IsComputed(), "id should be computed")
}

func TestUserResource_Schema(t *testing.T) {
	ctx := context.Background()
	schemaRequest := resource.SchemaRequest{}
	schemaResponse := &resource.SchemaResponse{}

	resources.NewUserResource().Schema(ctx, schemaRequest, schemaResponse)

	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema returned diagnostics: %+v", schemaResponse.Diagnostics)
	}

	// Verify required attributes
	usernameAttr, ok := schemaResponse.Schema.Attributes["username"]
	assert.True(t, ok, "username attribute should exist")
	assert.True(t, usernameAttr.IsRequired(), "username should be required")

	emailAttr, ok := schemaResponse.Schema.Attributes["email"]
	assert.True(t, ok, "email attribute should exist")
	assert.True(t, emailAttr.IsRequired(), "email should be required")

	// Verify computed attributes
	idAttr, ok := schemaResponse.Schema.Attributes["id"]
	assert.True(t, ok, "id attribute should exist")
	assert.True(t, idAttr.IsComputed(), "id should be computed")

	// Verify optional attributes with defaults
	isAdminAttr, ok := schemaResponse.Schema.Attributes["is_admin"]
	assert.True(t, ok, "is_admin attribute should exist")
	assert.True(t, isAdminAttr.IsOptional(), "is_admin should be optional")
	assert.True(t, isAdminAttr.IsComputed(), "is_admin should be computed")

	disabledAttr, ok := schemaResponse.Schema.Attributes["disabled"]
	assert.True(t, ok, "disabled attribute should exist")
	assert.True(t, disabledAttr.IsOptional(), "disabled should be optional")
	assert.True(t, disabledAttr.IsComputed(), "disabled should be computed")
}
