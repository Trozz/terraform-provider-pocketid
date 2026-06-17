package resources_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"

	"github.com/Trozz/terraform-provider-pocketid/internal/resources"
)

func TestNewLdapSyncResource(t *testing.T) {
	r := resources.NewLdapSyncResource()
	assert.NotNil(t, r)
	assert.Implements(t, (*resource.Resource)(nil), r)
}

func TestLdapSyncResource_Metadata(t *testing.T) {
	r := resources.NewLdapSyncResource()

	resp := &resource.MetadataResponse{}
	r.Metadata(context.TODO(), resource.MetadataRequest{ProviderTypeName: "pocketid"}, resp)

	assert.Equal(t, "pocketid_ldap_sync", resp.TypeName)
}

func TestLdapSyncResource_Schema(t *testing.T) {
	r := resources.NewLdapSyncResource()

	resp := &resource.SchemaResponse{}
	r.Schema(context.TODO(), resource.SchemaRequest{}, resp)

	schema := resp.Schema
	assert.NotNil(t, schema)
	assert.Contains(t, schema.Attributes, "id")
	assert.Contains(t, schema.Attributes, "triggers")
	assert.Contains(t, schema.Attributes, "synced_at")

	idAttr := schema.Attributes["id"]
	assert.True(t, idAttr.IsComputed())

	triggersAttr := schema.Attributes["triggers"]
	assert.True(t, triggersAttr.IsOptional())
	assert.False(t, triggersAttr.IsComputed())

	syncedAtAttr := schema.Attributes["synced_at"]
	assert.True(t, syncedAtAttr.IsComputed())
}

func TestLdapSyncResource_Configure(t *testing.T) {
	tests := []struct {
		name         string
		providerData interface{}
		expectError  bool
	}{
		{name: "nil provider data", providerData: nil, expectError: false},
		{name: "invalid provider data type", providerData: "invalid", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, ok := resources.NewLdapSyncResource().(resource.ResourceWithConfigure)
			assert.True(t, ok, "resource should implement ResourceWithConfigure")

			resp := &resource.ConfigureResponse{}
			r.Configure(context.TODO(), resource.ConfigureRequest{ProviderData: tt.providerData}, resp)

			assert.Equal(t, tt.expectError, resp.Diagnostics.HasError())
		})
	}
}
