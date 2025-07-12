package resources_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"

	"github.com/Trozz/terraform-provider-pocketid/internal/resources"
)

func TestNewOneTimeAccessTokenResource(t *testing.T) {
	r := resources.NewOneTimeAccessTokenResource()
	assert.NotNil(t, r)
	assert.Implements(t, (*resource.Resource)(nil), r)
}

func TestOneTimeAccessTokenResource_Metadata(t *testing.T) {
	r := resources.NewOneTimeAccessTokenResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "pocketid",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.TODO(), req, resp)

	assert.Equal(t, "pocketid_one_time_access_token", resp.TypeName)
}

func TestOneTimeAccessTokenResource_Schema(t *testing.T) {
	r := resources.NewOneTimeAccessTokenResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.TODO(), req, resp)

	// Verify schema attributes
	schema := resp.Schema
	assert.NotNil(t, schema)

	// Check required attributes
	assert.Contains(t, schema.Attributes, "id")
	assert.Contains(t, schema.Attributes, "user_id")
	assert.Contains(t, schema.Attributes, "token")
	assert.Contains(t, schema.Attributes, "expires_at")
	assert.Contains(t, schema.Attributes, "created_at")

	// Verify attribute properties
	idAttr := schema.Attributes["id"]
	assert.True(t, idAttr.IsComputed())
	assert.False(t, idAttr.IsRequired())

	userIDAttr := schema.Attributes["user_id"]
	assert.True(t, userIDAttr.IsRequired())
	assert.False(t, userIDAttr.IsComputed())

	tokenAttr := schema.Attributes["token"]
	assert.True(t, tokenAttr.IsComputed())
	assert.True(t, tokenAttr.IsSensitive())

	expiresAtAttr := schema.Attributes["expires_at"]
	assert.True(t, expiresAtAttr.IsOptional())
	assert.True(t, expiresAtAttr.IsComputed())

	createdAtAttr := schema.Attributes["created_at"]
	assert.True(t, createdAtAttr.IsComputed())
	assert.False(t, createdAtAttr.IsRequired())

	// Check skip_recreate attribute
	skipRecreateAttr := schema.Attributes["skip_recreate"]
	assert.True(t, skipRecreateAttr.IsOptional())
	assert.False(t, skipRecreateAttr.IsRequired())
	assert.False(t, skipRecreateAttr.IsComputed())
}

func TestOneTimeAccessTokenResource_Configure(t *testing.T) {
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
			r := resources.NewOneTimeAccessTokenResource()

			// Cast to the concrete type that implements Configure
			tokenResource, ok := r.(*resources.OneTimeAccessTokenResource)
			assert.True(t, ok, "Resource should be of type *OneTimeAccessTokenResource")

			req := resource.ConfigureRequest{
				ProviderData: tt.providerData,
			}
			resp := &resource.ConfigureResponse{}

			tokenResource.Configure(context.TODO(), req, resp)

			if tt.expectError {
				assert.True(t, resp.Diagnostics.HasError())
			} else {
				assert.False(t, resp.Diagnostics.HasError())
			}
		})
	}
}
