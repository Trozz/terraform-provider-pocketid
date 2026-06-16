package resources_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"

	"github.com/Trozz/terraform-provider-pocketid/internal/resources"
)

func TestNewScimServiceProviderResource(t *testing.T) {
	r := resources.NewScimServiceProviderResource()
	assert.NotNil(t, r)
	assert.Implements(t, (*resource.Resource)(nil), r)
}

func TestScimServiceProviderResource_Metadata(t *testing.T) {
	r := resources.NewScimServiceProviderResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "pocketid",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.TODO(), req, resp)

	assert.Equal(t, "pocketid_scim_service_provider", resp.TypeName)
}

func TestScimServiceProviderResource_Schema(t *testing.T) {
	r := resources.NewScimServiceProviderResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.TODO(), req, resp)

	schema := resp.Schema
	assert.NotNil(t, schema)

	assert.Contains(t, schema.Attributes, "id")
	assert.Contains(t, schema.Attributes, "client_id")
	assert.Contains(t, schema.Attributes, "endpoint")
	assert.Contains(t, schema.Attributes, "token")
	assert.Contains(t, schema.Attributes, "last_synced_at")
	assert.Contains(t, schema.Attributes, "created_at")

	idAttr := schema.Attributes["id"]
	assert.True(t, idAttr.IsComputed())
	assert.False(t, idAttr.IsRequired())

	clientIDAttr := schema.Attributes["client_id"]
	assert.True(t, clientIDAttr.IsRequired())
	assert.False(t, clientIDAttr.IsComputed())

	endpointAttr := schema.Attributes["endpoint"]
	assert.True(t, endpointAttr.IsRequired())

	tokenAttr := schema.Attributes["token"]
	assert.True(t, tokenAttr.IsOptional())
	assert.True(t, tokenAttr.IsSensitive())

	lastSyncedAtAttr := schema.Attributes["last_synced_at"]
	assert.True(t, lastSyncedAtAttr.IsComputed())

	createdAtAttr := schema.Attributes["created_at"]
	assert.True(t, createdAtAttr.IsComputed())
}

func TestScimServiceProviderResource_Configure(t *testing.T) {
	tests := []struct {
		name         string
		providerData interface{}
		expectError  bool
	}{
		{
			name:         "nil provider data",
			providerData: nil,
			expectError:  false,
		},
		{
			name:         "invalid provider data type",
			providerData: "invalid",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := resources.NewScimServiceProviderResource()

			cfgResource, ok := r.(resource.ResourceWithConfigure)
			assert.True(t, ok, "Resource should implement ResourceWithConfigure")

			req := resource.ConfigureRequest{
				ProviderData: tt.providerData,
			}
			resp := &resource.ConfigureResponse{}

			cfgResource.Configure(context.TODO(), req, resp)

			if tt.expectError {
				assert.True(t, resp.Diagnostics.HasError())
			} else {
				assert.False(t, resp.Diagnostics.HasError())
			}
		})
	}
}
